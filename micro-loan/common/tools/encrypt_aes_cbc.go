package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"

	"github.com/astaxie/beego/logs"
)

const (
	AesCBCKey string = "12f24fca6d6298fd1b7ff147559be43f"
	AesCBCIV  string = "b095a68b22ec1debea760810a4515505"
)

const (
	AesPKCS5 int = iota
	AesPKCS7
)

func AesDecryptUrlCode(ciphertext string, key string, iv string) (string, error) {
	decodeData, err := UrlDecode(ciphertext)
	if err != nil {
		logs.Warning("urldecode has wrong.")
		return "", err
	}

	return AesDecryptCBC(decodeData, key, iv)
}

func AesDecryptCBC(ciphertext string, key string, iv string) (string, error) {
	//logs.Debug("ciphertext:", ciphertext)
	base64Data, err := Base64Decode(ciphertext)
	if err != nil {
		logs.Warning("base64decode has wrong.")
		return "", err
	}

	aesKey, _ := hex.DecodeString(key)
	aesIV, _ := hex.DecodeString(iv)

	ciphertextByte, err := Decrypter(base64Data, aesKey, aesIV, AesPKCS5)
	if err != nil {
		logs.Warning("call Decrypter has wrong.")
		return "", err
	}
	//logs.Debug("ciphertext:", string(ciphertextByte))

	return string(ciphertextByte), nil
}

//解密
func Decrypter(crypted []byte, key []byte, iv []byte, paddingType int) ([]byte, error) {
	var err error
	emptyBytes := []byte{}

	sourceBlock, err := aes.NewCipher(key)
	if err != nil {
		return emptyBytes, err
	}
	if len(crypted)%sourceBlock.BlockSize() != 0 {
		err = errors.New("crypto/cipher: input not full blocks")
		return emptyBytes, err
	}

	source := make([]byte, len(crypted))
	sourceAes := cipher.NewCBCDecrypter(sourceBlock, iv)
	sourceAes.CryptBlocks(source, crypted)
	if paddingType == AesPKCS5 {
		source = PKCS5UnPadding(source)
	} else {
		source = PKCS7UnPadding(source)
	}

	return source, err
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	check := length - unpadding
	if check < 0 {
		return []byte{}
	}
	return origData[:check]
}

func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

// 包装过的简单加密函数
func AesEncryptCBC(src string, key string, iv string) (string, error) {
	aesKey, _ := hex.DecodeString(key)
	aesIV, _ := hex.DecodeString(iv)
	ciphertextByte, err := Encrypter([]byte(src), aesKey, aesIV, AesPKCS5)
	if err != nil {
		return "", err
	}

	return UrlEncode(Base64Encode(ciphertextByte)), nil
}

//加密
func Encrypter(source []byte, key []byte, iv []byte, paddingType int) ([]byte, error) {
	sourceBlock, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	if paddingType == AesPKCS5 {
		source = PKCS5Padding(source, sourceBlock.BlockSize()) //补全位数，长度必须是 16 的倍数
	} else {
		source = PKCS7Padding(source, sourceBlock.BlockSize())
	}

	sourceCrypted := make([]byte, len(source))
	sourceAes := cipher.NewCBCEncrypter(sourceBlock, iv)
	sourceAes.CryptBlocks(sourceCrypted, source)
	return sourceCrypted, err
}

// 补位
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//base64编码
func Base64Encode(data []byte) string {
	base64encodeBytes := base64.StdEncoding.EncodeToString(data)
	//logs.Debug("base64encode:", base64encodeBytes)
	return base64encodeBytes
}

//base64解码
func Base64Decode(data string) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(data)
	return decodeBytes, err
}

//url 编码
func UrlEncode(data string) string {
	encode := url.QueryEscape(data)
	return encode
}

//url 解码
func UrlDecode(data string) (string, error) {
	decodeurl, err := url.QueryUnescape(data)
	return decodeurl, err
}
