package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
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

func (a *Controller) Test(ctx context.Context, req *v1.IpcTestReq) (res *v1.IpcTestRes, err error) {
	res = &v1.IpcTestRes{}

	client := a.hub.GetClient("test")
	if client == nil {
		res.Msg = fmt.Sprintf(`{"msg": "%s"}`, "The service is unavailable")
		g.RequestFromCtx(ctx).Response.Writeln(res)
		return
	}

	clientIpc := client.GetIpc()
	reqIpc := clientIpc.Request("https://www.example.com/search?p=feng", ipc.RequestArgs{
		Method: "GET",
		Header: map[string]string{"Content-Type": "application/json"},
	})
	resIpc, err := clientIpc.Send(reqIpc)
	if err != nil {
		log.Println("ipc response err: ", err)
		res.Msg = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		g.RequestFromCtx(ctx).Response.Writeln(res)
		return
	}
	//for k, v := range res.Header {
	//	w.Header().Set(k, v)
	//}

	resStr, err := json.Marshal(resIpc)
	if err != nil {
		res.Msg = fmt.Sprintf(`{"msg": "%s"}`, err.Error())
		g.RequestFromCtx(ctx).Response.Writeln(res)
		return
	}

	res.Msg = fmt.Sprintf(`{"msg": %s}`, string(resStr))
	g.RequestFromCtx(ctx).Response.Writeln(res)
	return

}

//http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//	ws.ServeWs(hub, w, r)
//})
