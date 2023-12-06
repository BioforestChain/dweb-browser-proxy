package packed

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
