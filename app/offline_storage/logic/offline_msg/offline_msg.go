package offline_msg

import (
	"context"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/api/offline_msg/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/model"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/service"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/mongodb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	sOfflineMsg struct{}
)

func init() {
	service.RegisterOfflineMsg(New())
	logger = &LoggerIns{
		NewIns: NewPath(),
	}
}
func New() service.IOfflineMsg {
	return &sOfflineMsg{}
}

type LoggerIns struct {
	NewIns *glog.Logger
}

var logger *LoggerIns

func NewPath() *glog.Logger {
	logPath, _ := g.Cfg().Get(context.Background(), "logger.pathOfflineMsg") //
	return glog.New().Path(logPath.String())
}

// DelOfflineMsgById
//
//	@Description: 物理删除
//	@receiver s
//	@param ctx
//	@param in
//	@return err
//
//	func (s *sOfflineMsg) DelOfflineMsgById(ctx context.Context, in model.OfflineMsgDelInput) (err error) {
//		return dao.OfflineMsg.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
//			_, err = dao.OfflineMsg.Ctx(ctx).Unscoped().Delete("id", in.Id)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}

func (s *sOfflineMsg) GetOfflineMsgList(ctx context.Context, in model.OfflineMsgListQueryInput) (results []*v1.Res, total int, err error) {
	findOptions := options.Find()
	findOptions.SetLimit(int64(in.Limit))
	colName := in.ColName
	dbName := in.DbName
	clientID := in.Receiver
	//  ID: ObjectID("65966eda2374337c29554a46"),  C: xxxxxxxxxxxxxxxx, RS: [a.b.com b.b.com c.b.com]
	//	ID: ObjectID("65966f415496267801465653"),  C: xxxxxxxxxxxxxxxx, RS: [a.b.com b.b.com c.b.com]
	//	ID: ObjectID("659671086c74840da8831f31"),  C: xxxxxxxxxxxxxxxx, RS: [a.b.com b.b.com c.b.com]

	filter := bson.M{"rs": clientID}
	cur, err := mongodb.NewMgo(dbName, colName).CollectionDocuments(filter, int64(in.Offset), int64(in.Limit), -1)
	if err != nil {
		logger.NewIns.Error(ctx, "CollectionDocuments err: ", err, "status", 500)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		// 创建一个值，将单个文档解码为该值
		var result v1.Res
		err := cur.Decode(&result)
		if err != nil {
			logger.NewIns.Error(ctx, "CollectionDocuments cur Decode err: ", err, "status", 500)
		}
		results = append(results, &result)
	}
	if err := cur.Err(); err != nil {
		logger.NewIns.Error(ctx, "CollectionDocuments cur err: ", err, "status", 500)
	}
	// 完成后关闭游标
	cur.Close(context.Background())
	return results, len(results), nil
}
