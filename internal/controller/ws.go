package controller

import (
	"context"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/ws"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
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

	clientIPC := ipc.NewReadableStreamIPC(ipc.CLIENT, ipc.SupportProtocol{
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

	clientIPC.OnRequest(func(data any, ipcObj ipc.IPC) {
		request := data.(*ipc.Request)

		if len(request.Header.Get("X-Dweb-Pubsub")) > 0 {
			if err := pkg.DefaultPubSub.Handler(context.Background(), request, client); err != nil {
				// TODO
				log.Println("handlerPubSub err: ", err)

				body := []byte(fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error()))
				err = clientIPC.PostMessage(context.Background(), ipc.FromResponseBinary(request.ID, http.StatusOK, ipc.NewHeader(), body, ipcObj))
				fmt.Println("PostMessage err: ", err)
				return
			}

			body := []byte(fmt.Sprint(`{"success": true, "message": "ok"}`))
			err = clientIPC.PostMessage(context.Background(), ipc.FromResponseBinary(request.ID, http.StatusOK, ipc.NewHeader(), body, ipcObj))
			fmt.Println("PostMessage err: ", err)
		}
	})
}
