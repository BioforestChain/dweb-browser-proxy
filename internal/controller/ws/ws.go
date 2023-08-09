package ws

import (
	"context"
	"proxyServer/api/ws"

	_ "github.com/gogf/gf/v2/errors/gcode"
	_ "github.com/gogf/gf/v2/errors/gerror"
	v1 "proxyServer/api/client/v1"
)

type Controller struct {
	hub *ws.Hub
}

func New() *Controller {
	return &Controller{}
}

func (a *Controller) Test(ctx context.Context, req *v1.WsReq) (res *v1.WsRes, err error) {

	//ws.ServeWs(a.hub, nil, nil)
	//http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	//	ws.ServeWs(a.hub, w, r)
	//})
	//

	//ws.ServeWs(a.hub, w, r)

	//s.SetServerRoot(gfile.MainPkgPath())
	//s.SetPort(8000)
	//s.Run()

	return
}
