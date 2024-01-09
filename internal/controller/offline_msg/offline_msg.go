package offline_msg

import (
	"context"
	"fmt"
	v1 "github.com/BioforestChain/dweb-browser-proxy/api/offline_msg/v1"
	"github.com/BioforestChain/dweb-browser-proxy/internal/model"
	"github.com/BioforestChain/dweb-browser-proxy/internal/pkg/page"
	"github.com/BioforestChain/dweb-browser-proxy/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// OfflineMsgList
//
//	@Description:
//	@receiver c
//	@param ctx
//	@param req
//	@return res
//	@return err
func (c *Controller) OfflineMsgList(ctx context.Context, req *v1.Req) (res *v1.OfflineMsgListRes, err error) {
	start := time.Now()
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		fmt.Println("OfflineMsgList Validator", err)
	}
	condition := model.OfflineMsgListQueryInput{}
	condition.Page, condition.Limit, condition.Offset = page.InitCondition(req.Page, req.Limit)
	condition.Receiver = req.ClientID
	colName, _ := g.Cfg().Get(ctx, "offlineMsgDb.colName")
	dbName, _ := g.Cfg().Get(ctx, "offlineMsgDb.dbName")

	condition.ColName = colName.String()
	condition.DbName = dbName.String()

	list, total, err := service.OfflineMsg().GetOfflineMsgList(ctx, condition)

	end := time.Now()
	interval := end.Sub(start)
	fmt.Println("Collection FindOne 程序执行间隔时间:", interval)

	if err != nil {
		return
	}
	res = new(v1.OfflineMsgListRes)
	res.List = list
	res.Total = total
	res.Page = condition.Page
	res.LastPage = page.GetLastPage(int64(total), condition.Limit)

	return res, err
}
