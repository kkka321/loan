package short_url

import (
	"fmt"
	"strconv"

	"micro-loan/common/models"
	"micro-loan/common/tools"
)

var urlKey = "mobi_url"

var urlChars = []string{
	"a", "b", "c", "d", "e", "f", "g", "h",
	"i", "j", "k", "l", "m", "n", "o", "p",
	"q", "r", "s", "t", "u", "v", "w", "x",
	"y", "z", "0", "1", "2", "3", "4", "5",
	"6", "7", "8", "9", "A", "B", "C", "D",
	"E", "F", "G", "H", "I", "J", "K", "L",
	"M", "N", "O", "P", "Q", "R", "S", "T",
	"U", "V", "W", "X", "Y", "Z",
}

func generateUrl(url string) []string {
	hex := tools.Md5(urlKey + url)
	shortUrl := make([]string, 4)

	for i := 0; i < 4; i++ {
		subStr := hex[i*8 : i*8+8]
		subint, err := strconv.ParseInt(subStr, 16, 0)
		fmt.Println(err)
		hexint := 0x3FFFFFFF & subint
		subUrl := ""
		for j := 0; j < 6; j++ {
			index := 0x0000003D & hexint
			subUrl = subUrl + urlChars[index]
			hexint = hexint >> 5
		}
		shortUrl[i] = subUrl
	}

	return shortUrl
}

func crc8CheckNum(str string) uint8 {
	var crc uint8
	var i uint8
	crc = 0

	for j := 0; j < len(str); j++ {
		crc ^= str[j]
		for i = 0; i < 8; i++ {
			if crc&0x01 != 0 {
				crc = (crc >> 1) ^ 0x8c
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

func GenerateShortUrl(url string, host string) string {
	md5Url := tools.Md5(url)
	m, err := models.GetShortUrlByMd5(md5Url)
	if err == nil {
		return host + "/t/" + m.ShortUrl
	}

	tmpUrls := generateUrl(url)
	shortUrl := ""
	for _, v := range tmpUrls {
		tmpUrl := v
		_, err := models.GetShortUrl(tmpUrl)
		if err != nil {
			shortUrl = tmpUrl
			break
		}
	}

	if shortUrl != "" {
		m := models.ShortUrl{}
		m.ShortUrl = shortUrl
		m.Ctime = tools.GetUnixMillis()
		m.Url = url
		m.UrlMd5 = md5Url
		m.Insert()
	}

	if shortUrl == "" {
		return ""
	}

	return host + "/t/" + shortUrl
}

func GetShortUrl(shortUrl string) (models.ShortUrl, error) {
	return models.GetShortUrl(shortUrl)
}

/*
public static string[] ShortUrl(string url)
{
	//可以自定义生成MD5加密字符传前的混合KEY
	string key = "Leejor";
	//要使用生成URL的字符
	string[] chars = new string[]{
 		"a","b","c","d","e","f","g","h",
     	"i","j","k","l","m","n","o","p",
		"q","r","s","t","u","v","w","x",
   		"y","z","0","1","2","3","4","5",
 		"6","7","8","9","A","B","C","D",
  		"E","F","G","H","I","J","K","L",
  		"M","N","O","P","Q","R","S","T",
		"U","V","W","X","Y","Z"
	};

 	//对传入网址进行MD5加密
  	string hex = System.Web.Security.FormsAuthentication.HashPasswordForStoringInConfigFile(key + url, "md5");

	string[] resUrl = new string[4];

	for (int i = 0; i < 4; i++)
	{
		//把加密字符按照8位一组16进制与0x3FFFFFFF进行位与运算
    	int hexint = 0x3FFFFFFF & Convert.ToInt32("0x" + hex.Substring(i * 8, 8), 16);
		string outChars = string.Empty;
   		for (int j = 0; j < 6; j++)
    	{
        	//把得到的值与0x0000003D进行位与运算，取得字符数组chars索引
        	int index = 0x0000003D & hexint;
       		//把取得的字符相加
     		outChars += chars[index];
  			//每次循环按位右移5位
       		hexint = hexint >> 5;
    	}
		//把字符串存入对应索引的输出数组
 		resUrl[i] = outChars;
 	}
	return resUrl;
}

    ShortUrl(http://www.me3.cn)[0];  //得到值fAVfui
    ShortUrl(http://www.me3.cn)[1];  //得到值3ayQry
    ShortUrl(http://www.me3.cn)[2];  //得到值UZzyUr
    ShortUrl(http://www.me3.cn)[3];  //得到值36rQZn
*/
