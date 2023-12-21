package controller

import (
	"context"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/internal/consts"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/ws"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/gogf/gf/v2/net/ghttp"
	"io"
	"net/http"
)

var PubSub = new(pubsub)

type pubsub struct {
}

func (pb *pubsub) Pub(ctx context.Context, hub *ws.Hub, w http.ResponseWriter, r *ghttp.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer r.Body.Close()

	header := map[string]string{
		"X-Dweb-Host":              r.Header.Get(consts.XDwebHostMMID),
		"X-Dweb-Host-Domain":       r.Header.Get(consts.XDwebHostDomain), //net domain
		"X-Dweb-Pubsub":            r.Header.Get(consts.PubsubMMID),
		"X-Dweb-Pubsub-App":        r.Header.Get(consts.PubsubAppMMID),
		"X-Dweb-Pubsub-Net":        r.Header.Get("X-Dweb-Pubsub-Net"),
		"X-Dweb-Pubsub-Net-Domain": r.Header.Get("X-Dweb-Pubsub-Net-Domain"),
	}
	request := ipc.FromRequestBinary(1, "/pub", http.MethodPost, ipc.NewHeaderWithExtra(header), data, ipc.NewBaseIPC())
	// todo appMmid+topic
	//  X-Dweb-Pubsub-App:testmodule.bagen.com.dweb app_mmid
	// TODO for test
	client := hub.GetClient(r.GetQuery("client_id").String())

	res, err := pkg.DefaultPubSub.Pub(ctx, request, client)
	msg := "fail"
	if res {
		msg = "ok"
	}
	if err != nil {
		msg = err.Error()
	}

	fmt.Fprintf(w, msg)
}

func (pb *pubsub) Sub(ctx context.Context, hub *ws.Hub, w http.ResponseWriter, r *ghttp.Request) {
	// TODO for test
	client := hub.GetClient(r.GetQuery("client_id").String())

	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer r.Body.Close()

	header := map[string]string{
		"X-Dweb-Host":              r.Header.Get(consts.XDwebHostMMID),
		"X-Dweb-Host-Domain":       r.Header.Get(consts.XDwebHostDomain), //net domain
		"X-Dweb-Pubsub":            r.Header.Get(consts.PubsubMMID),
		"X-Dweb-Pubsub-App":        r.Header.Get(consts.PubsubAppMMID),
		"X-Dweb-Pubsub-Net":        r.Header.Get("X-Dweb-Pubsub-Net"),
		"X-Dweb-Pubsub-Net-Domain": r.Header.Get("X-Dweb-Pubsub-Net-Domain"),
	}

	request := ipc.FromRequestBinary(1, "/sub", http.MethodPost, ipc.NewHeaderWithExtra(header), data, ipc.NewBaseIPC())

	msg := "ok"
	err = pkg.DefaultPubSub.Sub(ctx, request, client)
	if err != nil {
		msg = err.Error()
	}
	fmt.Fprintf(w, msg)
}
