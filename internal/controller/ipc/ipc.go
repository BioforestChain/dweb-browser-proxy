package ipc

import (
	"context"
	"encoding/json"
	"log"
	"proxyServer/ipc"

	_ "github.com/gogf/gf/v2/errors/gcode"
	_ "github.com/gogf/gf/v2/errors/gerror"
	"proxyServer/api/client/v1"
	"proxyServer/api/ws"
)

// ipc管理
var Ipc = cIpc{}

type cIpc struct{}

type Controller struct {
	hub *ws.Hub
}

func New(hub *ws.Hub) *Controller {
	return &Controller{hub: hub}
}

func (a *Controller) Test(ctx context.Context, req *v1.IpcReq) (res *v1.IpcRes, err error) {
	res = &v1.IpcRes{}

	client := a.hub.GetClient("test")
	if client == nil {
		res.Ipc = "The service is unavailable"
		return res, nil
	}

	clientIpc := client.GetIpc()
	//url: http://127.0.0.1:8000/user/client-reg
	//
	reqIpc := clientIpc.Request("https://www.example.com/search?p=feng", ipc.RequestArgs{
		Method: "GET",
		Header: map[string]string{"Content-Type": "application/json"},
	})
	resIpc, err := clientIpc.Send(reqIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		//res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		res.Ipc = err.Error()
		return res, err
	}
	//todo
	//for k, v := range resIpc.Header {
	//	w.Header().Set(k, v)
	//}

	resStr, err := json.Marshal(resIpc)
	if err != nil {
		//res.Ipc = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		res.Ipc = err.Error()
		return res, err
	}
	res.Ipc = string(resStr)
	return res, err
}

//http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//	ws.ServeWs(hub, w, r)
//})
