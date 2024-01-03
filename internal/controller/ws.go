package controller

import (
	"context"
	"fmt"
	pubsub2 "github.com/BioforestChain/dweb-browser-proxy/app/pubsub"
	"github.com/BioforestChain/dweb-browser-proxy/internal/logic/net"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/rsa"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	ws2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
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
// sign:   SfqOJDVyBWrGB6IK3Z1q0Co2dTtmXTUlmhFx52cb3jco7jyBzbk224brucS3Du5M69p0nfVJ28jzlvlAtAVoJxzLMy2DUF2C+nUnNCMm7/0qyseumooBSE3rwdBrDHffwnNS+6DXS2TfM29xaz0HnBhfI5aChYU4qGz1+RoJxhCmfLKkymeQyLmOh7zHiJUy41XPCAVuhPXpyJgMqIq2hoXoPkq9jm1Qf0Vc8SS9hRsK4gom9b9INyb0z4jbjtg9acQmyKUqKMOiWfwY+ieSFlo42sMV/vMVJeDxsoo6cyOEGLh1IK75CaUBjgEf/+5y2nUMnahtUFgRMM2Ld+d0pg==
// prikey:MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCpZBLkcQBFy0JXByA5xMH83HNVX1CTZ792g1CD3AJN814+jOJ0iKD/Qttqnp8Sx0SUdMjPQosMmtd07vIcCNSApEw7b2VYvhzzJWhfgsTgnozrCtegnVqdmH+MKJbxcEIOWbwo0yyDeiGgMzqYpB48Lnuyk0MyIvQ3SMcBMmdmT9KYeAMIDVD5FjDX9e51aKA6TCCab5zcJzw4nrSLxQSMO+rY4gHZF2oymTeuv2JNdhiXzaOqHvYLnK8WeU+Bq98iBUYZYxuMdVfPOgbeiavXMZlGStB/Vrck4ZI3COfnozrrJVL4HNUfvolXSUL955bhWV99JFNTzLBQz5tr8REpAgMBAAECggEAAPPTmV47Sqksa9HZ8Ak3sATLTzkUeme0b2Won70udCLArmhbY51bDmRhWHWg3lmpfI41jNjKcF00he9MBDVWgIyr8w+aycdz4TgeIJ9bHQo3L6Elej04Q7eWTAL0alIfxPXjNXWOhAS9SKnRFAekNeFrB6OFVrRUnDT4DX0rkKsy4VKpIcuQpOyxzDsqOljj1WptiaWi3XjvQoHz/DSU3PP0wu7ypK9mSwJ/nyGM8qoldSuawHqstFOAjzzmeV6ZEX2fuq31pAM/qPWA95/sMKi1iS4uYXZpr+y0n6GPtqQRwleosWFv9lziE9hjKmvpDw1kv/TZVLQIBYyhTxfJgQKBgQDkn+1khM/Ez5zUAl0H8ZUvd/xzNdHcNl/WE/NZ8gNgvxNSWM5sdBorwsppoIDN8lIIkLjzWPg/az0E/+5Hh4fuVIYIiejrbEduzlgup6buyuSCb+kTYGd51GVI/eliEp1O27Xdw6vXAMmC5lmJSGGsVRgazus+UlRPhjWlJZwdYQKBgQC9rHFLLaeIWSxxYUJwDwh1wcg91x9Oo9V6rWEGbkpzJdwHYg8Dwr68Pe8EgA6W/D9rKWz4gi3hWbef2RbeG2+0/wj8veRBg0CMHZlMfSCyOS4KD+RNLEcEHSbNrbSlTQLdiF8bj1uQ19tcxhAFkKaUxv2r8u4ziWGk0Rh/Ux4AyQKBgQCOh2l+5hGFWA0kWwjWf/SKsFnRFXdsuvVKSAvJQkh60wRfrP+bu1HpgDmiWi6StgQQVPEIvKmfF+LlsAxDyamjmkwpHJj50/pAiSGOjHRUoGaPLud2bf50hEZUl/8cZhBt7ilWRLtngZUfJy4gmOBTiIVLiT49DySCo1/kQisuYQKBgAHXNpJAMywDkYbYJsjnnHFoHAVdnRQqStwR6qshTt+nMmdv8C1dKnSxNSyaAYo9kG/9yuzudnuFX17RwIMPSRo8j13Eif6Iw4uYjfBMFpEkNOosFU8aauYDUmkUkng4MxrrQ+EElyLktWBFG8qyCKvQ8o1EokMlxijPTqmNqPDhAoGACq28xUt7Mwc7DghhJwRKIHQ4s6Iu4R3yK8kRTyJcRnfhFK+/29y9okF3OQpu07wkOSzF90SWHgBpxug/TU9p+1tllGkV7weHmNvT8pM6ZPTTYO68yVJVBppCAidnfXiwBFfQ2MIFvMMuZjFW6rQ//AAnL3LCcvkfpTFqHev7HOI=
// publicKey:  MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqWQS5HEARctCVwcgOcTB/NxzVV9Qk2e/doNQg9wCTfNePozidIig/0Lbap6fEsdElHTIz0KLDJrXdO7yHAjUgKRMO29lWL4c8yVoX4LE4J6M6wrXoJ1anZh/jCiW8XBCDlm8KNMsg3ohoDM6mKQePC57spNDMiL0N0jHATJnZk/SmHgDCA1Q+RYw1/XudWigOkwgmm+c3Cc8OJ60i8UEjDvq2OIB2RdqMpk3rr9iTXYYl82jqh72C5yvFnlPgavfIgVGGWMbjHVXzzoG3omr1zGZRkrQf1a3JOGSNwjn56M66yVS+BzVH76JV0lC/eeW4VlffSRTU8ywUM+ba/ERKQIDAQAB
func (wst *webSocket) Connect(hub *ws2.Hub, w http.ResponseWriter, r *http.Request) {
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

	client := ws2.NewClient(r.URL.Query().Get("client_id"), hub, conn, clientIPC)

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
			if err := pubsub2.DefaultPubSub.Handler(context.Background(), request, client); err != nil {
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
