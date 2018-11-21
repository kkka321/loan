package tools

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"golang.org/x/text/message"

	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const charset string = "abcdefghzkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ3456789" //随机因子

func GenerateRandomStr(length int) string {
	retStr := ""
	csLeng := len(charset)
	for i := 0; i < length; i++ {
		randNum := GenerateRandom(0, csLeng)
		retStr += string(charset[randNum])
	}

	return retStr
}

// 生成一个区间范围的随机数,左闭右开
func GenerateRandom(min, max int) int {
	if min >= max {
		return max
	}

	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max - min)
	randNum += min

	return randNum
}

//! 手机验证在4-8位之间
func GenerateMobileCaptcha(length int) string {
	if length < 4 || length > 8 {
		return ""
	}

	minStr := "1" + strings.Repeat("0", length-1)
	maxStr := "1" + strings.Repeat("0", length)

	min, _ := strconv.Atoi(minStr)
	max, _ := strconv.Atoi(maxStr)

	captcha := GenerateRandom(min, max)
	return strconv.Itoa(captcha)
}

func GetCurrentEnv() string {
	return beego.AppConfig.String("runmode")
}

func IsProductEnv() bool {
	return GetCurrentEnv() == "prod"
}

func DBDriver() string {
	return beego.AppConfig.String("db_type")
}

func GetLocalUploadPrefix() string {
	return beego.AppConfig.String("upload_prefix")
}

func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}

	return data
}

// CheckRequiredParameter 通用的检查必要参数的方法,只检测参数存在,不关心参数值
func CheckRequiredParameter(parameter map[string]interface{}, requiredParameter map[string]bool) bool {
	var requiredCheck int
	var rpCopy = make(map[string]bool)
	for rp, v := range requiredParameter {
		rpCopy[rp] = v
	}

	for k := range parameter {
		if requiredParameter[k] {
			requiredCheck++
			delete(rpCopy, k)
		}
	}

	if len(requiredParameter) != requiredCheck {
		var lostParam []string
		for l := range rpCopy {
			lostParam = append(lostParam, l)
		}
		logs.Error("request lost required parameter, parameter:", parameter, fmt.Sprintf("lostParam: [%s]", strings.Join(lostParam, ", ")))
		return false
	}

	return true
}

func ThreeElementExpression(status bool, exp1 interface{}, exp2 interface{}) (result interface{}) {
	if status {
		return exp1
	} else {
		return exp2
	}
}

func FullStack() string {
	var buf [2 << 11]byte
	runtime.Stack(buf[:], true)
	return string(buf[:])
}

func ClearOnSignal(handler func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		handler()
		os.Exit(0)
	}()
}

// IntsSliceToWhereInString 将状态或者IDs集合转换为string
// interface{}支持所有int, int8 etc.: %d
func IntsSliceToWhereInString(intsSlice interface{}) (s string, err error) {
	sl, err := ToSlice(intsSlice)
	if err != nil {
		return
	}
	for _, i := range sl {
		s += fmt.Sprintf("%d,", i)
	}
	if len(s) > 0 {
		s = strings.TrimSuffix(s, ",")
	} else {
		err = fmt.Errorf("[IntsSliceToWhereInString] generate empty string, will occur sql error, with param %v", intsSlice)
	}
	return
}

// ToSlice 转化 泛型为 slice
func ToSlice(arr interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("[ToSlice] should be slice param, but %#v", arr)
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret, nil
}

//MobileFormat 去除空格，并且不能加拨0或62
func MobileFormat(mobile string) string {
	// 去除空格
	str := strings.Replace(mobile, " ", "", -1)
	if strings.HasPrefix(str, "08") {
		str = strings.Replace(str, "08", "8", 1)
	}
	if strings.HasPrefix(str, "628") {
		str = strings.Replace(str, "628", "8", 1)
	}
	return str
}

// NumberFormat 输出格式化数字，千分位以逗号分割
func NumberFormat(number interface{}) string {
	p := message.NewPrinter(message.MatchLanguage("en"))
	return p.Sprint(number)
}

// SliceInt64ToMap 输出格式化数字，千分位以逗号分割
func SliceInt64ToMap(s []int64) map[int64]interface{} {
	m := make(map[int64]interface{}, len(s))
	for _, v := range s {
		m[v] = nil
	}
	return m
}
