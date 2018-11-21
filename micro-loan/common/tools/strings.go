package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// 字串截取
func SubString(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[pos:l])
}

func Strim(str string) string {
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)

	return str
}

// StrReplace 在 origin 中搜索 search 组,替换成 replace
func StrReplace(origin string, search []string, replace string) (s string) {
	s = origin
	for _, find := range search {
		s = strings.Replace(s, find, replace, -1)
	}

	return
}

func TrimRealName(name string) string {
	// 将名字里面的标点符号替换成1个空格

	name = ReplaceInvalidRealName(name)
	reg := regexp.MustCompile(`[\pP]+?`)
	name = reg.ReplaceAllString(name, " ")

	// 将连续的空白替换成一个空格
	reg = regexp.MustCompile(`\s{2,}`)
	name = reg.ReplaceAllString(name, " ")

	name = strings.TrimSpace(name)

	return name
}

func ReplaceInvalidRealName(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z\\s]+")
	ret := reg.ReplaceAllString(name, "")
	return ret
}

// IsIndonesiaName 判断是否是合法的印尼名字
func IsIndonesiaName(name string) (valid bool) {
	var exp, _ = regexp.Compile(`^[a-zA-Z ]+$`)
	if exp.MatchString(name) {
		return true
	}
	return false
}

// IsNumber 判断是否都是数字
func IsNumber(str string) (valid bool) {
	var exp, _ = regexp.Compile(`^[0-9]+$`)
	if exp.MatchString(str) {
		return true
	}
	return false
}

func ContainNumber(str string) (valid bool) {
	var exp, _ = regexp.Compile(`[\d]`)
	if exp.MatchString(str) {
		return true
	}
	return false
}

func Str2Int64(str string) (int64, error) {
	number, err := strconv.ParseInt(str, 10, 64)
	return number, err
}

func Int642Str(number int64) string {
	return strconv.FormatInt(number, 10)
}

func Str2Int(str string) (int, error) {
	number, err := strconv.ParseInt(str, 10, 0)
	return int(number), err
}

func Int2Str(number int) string {
	return strconv.FormatInt(int64(number), 10)
}

func Float2Str(f float32) string {
	return Float642Str(float64(f))
}

func Float642Str(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func Str2Float64(s string) (f float64, err error) {
	f, err = strconv.ParseFloat(s, 64)

	return
}

func Str2Float(s string) (f float32, err error) {
	f64, err := Str2Float64(s)
	f = float32(f64)

	return
}

func JsonEncode(d interface{}) (jsonStr string, err error) {
	bson, err := json.Marshal(d)
	jsonStr = string(bson)

	return
}

func Unicode(rs string) string {
	json := ""
	for _, r := range rs {
		rint := int(r)
		if rint < 128 {
			json += string(r)
		} else {
			json += "\\u" + strconv.FormatInt(int64(rint), 16)
		}
	}

	return json
}

func Escape(html string) string {
	return template.HTMLEscapeString(html)
}

func AddSlashes(str string) string {
	str = strings.Replace(str, `\`, `\\`, -1)
	str = strings.Replace(str, "'", `\'`, -1)
	str = strings.Replace(str, `"`, `\"`, -1)

	return str
}

func StripSlashes(str string) string {
	str = strings.Replace(str, `\'`, `'`, -1)
	str = strings.Replace(str, `\"`, `"`, -1)
	str = strings.Replace(str, `\\`, `\`, -1)

	return str
}

func RawUrlEncode(s string) (r string) {
	r = UrlEncode(s)
	r = strings.Replace(r, "+", "%20", -1)
	return
}

// 直接json.Marshal ，  会把 < > & 转成 unicode 编码
// JSONMarshal 解决直接json.Marshal 后单引号，双引号，< > & 符号的问题
func JSONMarshal(v interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	jsonStr := string(buf.String())
	return jsonStr, err
}

//将slice 转化成字符串
//[]int{1, 2, 3, 4, 5}  => 1,2,3,4,5 或
//[]string{"1", "2", "3", "4", "5"}  => 1,2,3,4,5
//"AAA bbb" 转为  AAA,bbb
//其他类型返回空字符串
func ArrayToString(a interface{}, delim string) (newStr string) {
	vtype := reflect.TypeOf(a).String()
	if vtype == "[]int" || vtype == "[]int64" {
		newStr = strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	} else if vtype == "[]string" {

		// 单独处理[]string类型是因为，我们要处理的字符串值可能有空格，但是也会被当成需要分格的对象，所以先替换处理，然后再替换回去
		specialChar := "^^"
		newSlice := make([]string, 0)
		for _, v := range a.([]string) {
			newSlice = append(newSlice, strings.Replace(v, " ", specialChar, -1))
		}
		fileds := strings.Fields(fmt.Sprint(newSlice))
		join := strings.Replace(strings.Join(fileds, delim), specialChar, " ", -1)
		newStr = strings.Trim(join, "[]")
	} else if vtype == "string" {
		newStr = strings.Replace(fmt.Sprint(a.(string)), " ", delim, -1)
	} else {
		newStr = ""
	}
	return
}

func GetIntKeysOfMap(mymap map[int]string) (keys []int) {
	keys = make([]int, 0, len(mymap))
	for k := range mymap {
		keys = append(keys, k)
	}
	return
}

// IsValidIndonesiaMobile 是否是印尼的有效电话号
// 08 开头, 10-13位数字, 2018.08,有13位手机号段了
// 2018.11 增加15位手机号判断
func IsValidIndonesiaMobile(mobile string) (yes bool, err error) {
	mobileLen := len(mobile)
	if mobileLen < 10 || mobileLen > 15 {
		err = fmt.Errorf("mobile length is invalid, mobile: %s", mobile)
		return
	}

	if "08" != SubString(mobile, 0, 2) {
		err = fmt.Errorf("mobile is invalid, mobile: %s", mobile)
		return
	}

	_, errNum := Str2Int64(mobile)
	if errNum != nil {
		err = fmt.Errorf("mobile has invalid char, mobile: %s, errNum: %v", mobile, errNum)
		return
	}

	yes = true

	return yes, nil
}

// ParseTableName 从SQL语句中解析主表名
func ParseTableName(sql string) (name string, err error) {
	re := regexp.MustCompile(`(?i).+FROM\s+(\S+)`)
	allMatch := re.FindAllStringSubmatch(sql, -1)

	if len(allMatch) == 0 {
		err = fmt.Errorf("parse sql has error.")
		return
	}

	tableName := strings.Replace(allMatch[0][1], "`", "", -1)
	names := strings.Split(tableName, ".")
	name = names[len(names)-1]

	return
}

// 手机号脱敏处理(只保留前两位和后四位，中间的每个字符都替换为"*")
// 例如：08123456789 改为 08*****6789, 0812345645678 改为 08*******5678
func MobileDesensitization(src string) (dst string) {
	length := len(src)

	if length >= 7 {
		prefix := src[0:3]
		var middle []string
		for i := 0; i < length-7; i++ {
			middle = append(middle, "*")
		}
		suffix := src[length-4 : length]

		dst = fmt.Sprintf("%s%s%s", prefix, strings.Join(middle, ""), suffix)
	}

	return
}

func ParseTargetList(str string) []string {
	list := make([]string, 0)
	if str == "" {
		return list
	}

	listStr := strings.Split(str, "\n")

	c := ","
	if strings.Contains(listStr[0], "\r") {
		c = "\r"
	} else if strings.Contains(listStr[0], ",") {
		c = ","
	}

	if len(listStr) == 1 {
		vec := strings.Split(listStr[0], c)

		list = append(list, strings.Trim(vec[0], " "))

		return list
	}

	for _, v := range listStr {
		vec := strings.Split(v, c)
		if len(vec) < 1 {
			continue
		}

		list = append(list, strings.Trim(vec[0], " "))
	}

	return list
}

/*
func GetUnpaidAmountKey(str string, t int64) (key string) {
	key = fmt.Sprintf("%s:%s", str, MDateMHSDate(t))
	return
}
*/
