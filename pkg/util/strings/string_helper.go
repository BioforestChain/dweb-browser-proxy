package strings

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"reflect"
	"strings"
	"unsafe"
)

func GetURL(r *ghttp.Request) (Url string) {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	return strings.Join([]string{scheme, r.Host, r.RequestURI}, "")
}

func GetHost(r *ghttp.Request) (Host string) {
	return strings.Join([]string{r.Host}, "")
}

// Explode 用一个字符串separator分割另外一个字符串str
func Explode(separator, str string) []string {
	return strings.Split(str, separator)
}

// Implode 切片转为指定字符串连接的为字符串
func Implode(separator string, data []string) string {
	return strings.Join(data, separator)
}

func Empty(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return reflect.DeepEqual(val, reflect.Zero(v.Type()).Interface())
}

// StrToBytes string转bytes
func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// BytesToStr bytes转string
func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// SubStr string截取
func SubStr(str string, offset int, length int) string {
	runeData := []rune(str)
	runeLen := len(runeData)

	if length < 0 {
		return ""
	}
	if (offset == 0 && length == runeLen) || runeLen <= length {
		return str
	}
	if offset >= 0 && length > 0 {
		limit := length + offset
		return string(runeData[offset:limit])
	}

	return ""
}
