package controller

import (
	"context"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/ws"
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var WsIns = new(webSocket)

type webSocket struct {
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Connect handles websocket requests from the peer.
func (wst *webSocket) Connect(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	upGrader.CheckOrigin = func(r *http.Request) bool {
		// Origin header have a pattern that *.xxx.com
		// TODO return r.Header.Get("Origin") == '*.xxx.com'
		return true
	}
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	clientIPC := ipc2.NewReadableStreamIPC(ipc2.CLIENT, ipc2.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	client := ws.NewClient(r.URL.Query().Get("client_id"), hub, conn, clientIPC)

	client.GetHub().Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.WritePump()
	go client.ReadPump()

	go func() {
		defer func() {
			client.Close()
			if err := recover(); err != nil {
				// TODO 日志上报
				log.Println("clientIPC.BindInputStream panic: ", err)
			}
		}()

		if err := clientIPC.BindInputStream(client.GetInputStream()); err != nil {
			log.Println("clientIPC.BindInputStream: ", err)
		}
	}()

	clientIPC.OnRequest(func(data any, ipcObj ipc2.IPC) {
		request := data.(*ipc2.Request)

		if len(request.Header.Get("X-Dweb-Pubsub")) > 0 {
			if err := pkg.DefaultPubSub.Handler(context.Background(), request, client); err != nil {
				// TODO
				log.Println("handlerPubSub err: ", err)

				body := []byte(fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error()))
				err = clientIPC.PostMessage(context.Background(), ipc2.FromResponseBinary(request.ID, http.StatusOK, ipc2.NewHeader(), body, ipcObj))
				fmt.Println("PostMessage err: ", err)
				return
			}

			body := []byte(fmt.Sprint(`{"success": true, "message": "ok"}`))
			err = clientIPC.PostMessage(context.Background(), ipc2.FromResponseBinary(request.ID, http.StatusOK, ipc2.NewHeader(), body, ipcObj))
			fmt.Println("PostMessage err: ", err)
		}
	})
}
