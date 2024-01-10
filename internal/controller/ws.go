package controller

import (
	"context"
	pubsub2 "github.com/BioforestChain/dweb-browser-proxy/app/pubsub"
	"github.com/BioforestChain/dweb-browser-proxy/internal/logic/net"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/rsa"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
	"github.com/gogf/gf/v2/frame/g"
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
	// secret check
	secretSrc := r.URL.Query().Get("secret")
	secret, _ := g.Cfg().Get(context.Background(), "auth.secret")
	if secretSrc != secret.String() {
		return
	}
	// clientId check
	clientId := r.URL.Query().Get("client_id")
	resData, err := net.GetNetPublicKeyByBroAddr(context.Background(), model.NetModulePublicKeyInput{
		BroadcastAddress: clientId,
	})
	sign := r.URL.Query().Get("s")
	publicKeyPem := "-----BEGIN PUBLIC KEY-----\n" + resData.PublicKey + "\n-----END PUBLIC KEY-----\n"
	// RsaVerySign
	resRsaVerySign := rsa.RsaVerySignWithSha256(clientId, sign, publicKeyPem)
	if !resRsaVerySign {
		return
	}

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
			if err = pubsub2.ProcessPubSub(context.Background(), client, request, ipcObj); err != nil {
				log.Println("ProcessPubSub err: ", err)
			}
		}
	})
}
