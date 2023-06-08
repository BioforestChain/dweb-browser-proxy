package user

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	v1 "frpConfManagement/api/client/v1"
	"frpConfManagement/internal/model"
	"frpConfManagement/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// SignUp is the API for user sign up.
func (c *Controller) ClientReg(ctx context.Context, req *v1.ClientRegReq) (res *v1.ClientRegRes, err error) {
	g.Array{
		sss
	}
	err = service.User().Create(ctx, model.UserCreateInput{
		Name:           req.Name,
		Identification: req.Identification,
		Remark:         req.Remark,
	})
	return
}
