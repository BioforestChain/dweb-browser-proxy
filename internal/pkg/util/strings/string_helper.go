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
