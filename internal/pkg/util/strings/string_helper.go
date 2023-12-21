package strings

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"strings"
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
