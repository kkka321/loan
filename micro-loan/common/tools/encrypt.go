package tools

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/astaxie/beego/logs"
)

// SignatureSecret 参数签名的盐
//const SignatureSecret string = "hy0le#GML0k"

//md5方法
func Md5(s string) string {
	return Md5Bytes([]byte(s))
}

func Md5Bytes(buf []byte) string {
	h := md5.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil))
}

func Sha1(data string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(data))
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

//Guid方法
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return Md5(base64.URLEncoding.EncodeToString(b))
}

func PasswordEncrypt(password string, salt int64) string {
	saltStr := strconv.FormatInt(salt, 10)
	md51 := Md5(password + "$" + saltStr)
	md52 := Md5(saltStr)

	result1 := SubString(md52, 24, 8)
	result2 := SubString(md51, 0, 24)

	return Md5(result1 + result2)
}

func Signature(params map[string]interface{}, secret string) string {
	paramLen := len(params)
	if paramLen <= 0 {
		return ""
	}

	cntr := make([]string, paramLen)
	var i int = 0
	for k, _ := range params {
		cntr[i] = k
		i++
	}

	// 按字典序列排序
	sort.Strings(cntr)

	str := "" // 待签名字符串
	for i = 0; i < paramLen; i++ {
		key := cntr[i]
		str += fmt.Sprintf("%s=%s&", key, params[key].(string))
	}
	str += secret
	logs.Debug("signature str:", str)

	return Md5(str)
}
