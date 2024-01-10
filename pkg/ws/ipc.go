package ws

import (
	"context"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
)

func SendIPC(ctx context.Context, client *Client, req *v1.IpcReq) (res *ipc.Response, err error) {
	var overallHeader = make(map[string]string)
	for k, v := range req.Header {
		overallHeader[k] = v[0]
	}

	var clientIpc = client.GetIpc()

	reqIpc := clientIpc.Request(req.URL, ipc.RequestArgs{
		Method: req.Method,
		Header: overallHeader,
		Body:   req.Body,
	})

	resIpc, err := clientIpc.Send(ctx, reqIpc)
	if err != nil {
		return nil, err
	}

	return resIpc, nil
}
