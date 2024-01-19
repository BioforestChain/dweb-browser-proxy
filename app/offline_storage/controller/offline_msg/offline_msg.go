package offline_msg

import (
	"context"
	"fmt"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/api/offline_msg/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/model"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/service"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/page"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

type LoggerIns struct {
	NewIns *glog.Logger
}

var logger *LoggerIns

// Init
//
//	@Description: 初始化
func init() {
	logger = &LoggerIns{
		NewIns: NewPath(),
	}
}

func NewPath() *glog.Logger {
	logPath, _ := g.Cfg().Get(context.Background(), "logger.pathOfflineMsg") //
	return glog.New().Path(logPath.String())
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
	if err != nil {
		logger.NewIns.Error(ctx, "OfflineMsgList err: ", err, "status", 500)
		return
	}
	res = new(v1.OfflineMsgListRes)
	res.List = list
	res.Total = total
	res.Page = condition.Page
	res.LastPage = page.GetLastPage(int64(total), condition.Limit)

	return res, err
}
