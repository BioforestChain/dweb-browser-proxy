package packed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"os"
	"os/exec"
	v1 "proxyServer/api/client/v1"
	"proxyServer/internal/model"
	"proxyServer/internal/service"
	ws "proxyServer/internal/service/ws"
	"proxyServer/ipc"
	"runtime"
	"strings"
)

var (
	err         error
	DevEnv      string
	TestConfig  string
	ProjectName string
)

// Proxy2Ipc
//
//	@Description: The request goes to the IPC object for processing
//	@param ctx
//	@param hub
//	@param req
//	@return res
//	@return err
func Proxy2Ipc(ctx context.Context, hub *ws.Hub, req *v1.IpcReq) (res *ipc.Response, err error) {
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return nil, errors.New("the service is unavailable")
	}
	// Verify req.ClientID exists in the database
	valCheckUser := service.User().IsUserExist(ctx, model.CheckUserInput{UserIdentification: req.ClientID})
	if !valCheckUser {
		return nil, gerror.Newf(`Sorry, your user "%s" is not registered yet`, req.ClientID)
	}
	var (
		clientIpc = client.GetIpc()
		//req.Header map[string]string{"Content-Type": req.Header, "xx": "1"},
		overallHeader = make(map[string]string)
	)
	for k, v := range req.Header {
		overallHeader[k] = v[0]
	}
	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: overallHeader,
		Body:   req.Body,
	})
	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		return nil, err
	}
	return resIpc, nil
}

func IpcErrResponse(code int, msg string) *ipc.Response {
	body := fmt.Sprintf(`{"code": %d, "message": "%s", "data": null}`, code, msg)
	newIpc := ipc.NewBaseIPC()
	res := ipc.NewResponse(
		1,
		400,
		ipc.NewHeaderWithExtra(map[string]string{
			"Content-Type": "application/json",
		}),
		ipc.NewBodySender([]byte(body), newIpc),
		newIpc,
	)
	return res
}

func ProjectPath() (path string) {
	// default linux/mac os
	var (
		sp = "/"
		ss []string
	)
	if runtime.GOOS == "windows" {
		sp = "\\"
	}

	// GOMOD
	// in go source code:
	// // Check for use of modules by 'go env GOMOD',
	// // which reports a go.mod file path if modules are enabled.
	// stdout, _ := exec.Command("go", "env", "GOMOD").Output()
	// gomod := string(bytes.TrimSpace(stdout))
	stdout, _ := exec.Command("go", "env", "GOMOD").Output()
	path = string(bytes.TrimSpace(stdout))
	if path == "/dev/null" {
		return ""
	}
	if path != "" {
		ss = strings.Split(path, sp)
		ss = ss[:len(ss)-1]
		path = strings.Join(ss, sp) + sp
		return
	}

	// GOPATH
	fileDir, _ := os.Getwd()
	path = os.Getenv("GOPATH") // < go 1.17 use
	ss = strings.Split(fileDir, path)
	if path != "" {
		ss2 := strings.Split(ss[1], sp)
		path += sp
		for i := 1; i < len(ss2); i++ {
			path += ss2[i] + sp
			return path
		}
	}
	return
}
