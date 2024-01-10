package controller

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage"
	"github.com/BioforestChain/dweb-browser-proxy/internal/consts"
	error2 "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/error"
	helperIPC "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/ipc"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
	"github.com/gogf/gf/v2/net/ghttp"
	"io"
	"log"
	"time"
)

var Proxy = new(proxy)

type proxy struct {
}

// Forward
//
//	@Description:
//
// https://mongodb.net.cn/manual/reference/method/db.collection.storageSize/
// appId 为 表名 colName
//
//	@receiver p
//	@param ctx
//	@param r
//	@param hub
func (p *proxy) Forward(ctx context.Context, r *ghttp.Request, hub *ws.Hub) {
	req := &v1.IpcReq{}
	req.Header = r.Header
	req.Method = r.Method
	req.URL = r.GetUrl()
	req.Host = r.GetHost()
	// TODO 需要优化body https://medium.com/@owlwalks/sending-big-file-with-minimal-memory-in-golang-8f3fc280d2c
	req.Body = r.GetBody()
	//TODO 暂定用 query 参数传递
	req.ClientID = req.Host

	//TODO test offline msg by post base on https
	var overallHeader = make(map[string]string)
	for k, v := range req.Header {
		overallHeader[k] = v[0]
	}
	//client := hub.GetClient(req.ClientID)
	if overallHeader[consts.XDwebOffline] == "1" {
		err := offline_storage.ProcessOfflineStorage(ctx, req, hub)
		if err != nil {
			log.Println("QuotaIsFull panic is ", err)
			r.Response.WriteStatus(400, err)
			return
		}

		r.Response.WriteStatus(200, "OK")
	} else {
		resIpc, err := Proxy2Ipc(ctx, req, hub)
		if err != nil {
			resIpc = helperIPC.ErrResponse(error2.ServiceIsUnavailable, err.Error())
		}
		for k, v := range resIpc.Header {
			r.Response.Header().Set(k, v)
		}
		bodyStream := resIpc.Body.Stream()
		if bodyStream == nil {
			if _, err = io.Copy(r.Response.Writer, resIpc.Body); err != nil {
				r.Response.WriteStatus(400, "请求出错")
			}
		} else {
			data, err := helperIPC.ReadStreamWithTimeout(bodyStream, 10*time.Second)
			if err != nil {
				r.Response.WriteStatus(400, err)
			} else {
				r.Response.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
				_, _ = r.Response.Writer.Write(data)
			}
		}
	}
}

// Proxy2Ipc
//
//	@Description: The request goes to the IPC object for processing
//	@param ctx
//	@param hub
//	@param req
//	@return res
//	@return err
func Proxy2Ipc(ctx context.Context, req *v1.IpcReq, hub *ws.Hub) (res *ipc.Response, err error) {
	client := hub.GetClient(req.ClientID)
	if client == nil {
		return nil, errors.New("the service is unavailable~")
	}

	return ws.SendIPC(ctx, client, req)
}
