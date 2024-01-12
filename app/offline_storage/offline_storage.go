package offline_storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	clientV1 "github.com/BioforestChain/dweb-browser-proxy/api/client/v1"
	"github.com/BioforestChain/dweb-browser-proxy/app/offline_storage/consts"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/mongodb"
	stringsHelper "github.com/BioforestChain/dweb-browser-proxy/pkg/util/strings"
	timeHelper "github.com/BioforestChain/dweb-browser-proxy/pkg/util/time"
	"github.com/BioforestChain/dweb-browser-proxy/pkg/ws"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
	"time"
)

type LoggerIns struct {
	NewIns *glog.Logger
}

var logger *LoggerIns

func init() {
	logger = &LoggerIns{
		NewIns: NewPath(),
	}
}

func NewPath() *glog.Logger {
	logPath, _ := g.Cfg().Get(context.Background(), "logger.pathOfflineMsg") //
	return glog.New().Path(logPath.String())
}

type OfflineMsgData struct {
	C  interface{}
	RS []string
	D  string
}

// QuotaIsFull
//
//	@Description: 额度判断函数，满则return，不满则把离线数据insert into db
//	@param stoSize
//	@param dbName
//	@param colName
//	@param content
//	@param receiver
//	@return err

type paramsCheckQuota struct {
	storageSize          uint32
	quotaSizeByColName   uint32
	presetCollectionSize uint32
	dbName, colName      string
	content              interface{}
	contentSize          uint32
	receiver             []string
}

// ProcessOfflineStorage
//
//	@Description:
//	@param ctx
//	@param req //req.ClientID : "a.b.com,b.b.com,c.b.com"
//	@param hub
//	@return error
func ProcessOfflineStorage(ctx context.Context, req *clientV1.IpcReq, hub *ws.Hub) error {
	content := req.Body
	clientStr := req.ClientID
	clientArrList := stringsHelper.Explode(",", clientStr)
	//模块 appId 为表名
	//colName = colNameSrc.String()
	//colNameSrc, _ := g.Cfg().Get(ctx, "offlineMsgDb.colName")
	colNameAppId := req.Header[consts.XDwebHostMMID][0]
	dbNameSrc, _ := g.Cfg().Get(ctx, "offlineMsgDb.dbName")
	dbName := dbNameSrc.String()
	quotaColName := consts.QuotaCollectioPrefix + colNameAppId
	presetCollectionSize, _ := g.Cfg().Get(context.Background(), "presetCollectionSize.quotaSize")
	quotaSizeByColName, _ := GetCollectionQuotaByAppId(dbName, quotaColName, presetCollectionSize.Uint32())
	storageSize, _ := GetCollectionStorageSizeByAppId(dbName, colNameAppId)
	contentSize, err := getBodySize(content)
	if err != nil {
		logger.NewIns.Error(ctx, "getBodySize (req.Body) err : ", err, "status", 500)
		//log.Println("mongo.Connect panic is : ", err)
	}
	// 离线状态的 接收者
	receiverList := []string{}

	for _, clientId := range clientArrList {
		client := hub.GetClient(clientId)
		//TODO待优化， 离线消息，额度判断
		if client == nil {
			// find offline clientList
			receiverList = append(receiverList, clientId)
		} else {
			if _, err := ws.SendIPC(ctx, client, req); err != nil {
				logger.NewIns.Error(ctx, "clientIpc.Send err : ", err, "status", 500)
			}
		}
	}

	// Collection Quota check
	params := paramsCheckQuota{
		storageSize,
		quotaSizeByColName,
		presetCollectionSize.Uint32(),
		dbName,
		colNameAppId,
		content,
		contentSize,
		receiverList,
	}

	return QuotaIsFull(params)
}

// QuotaIsFull
//
//	@Description: 判断是否额度满了
//	@param req
//	@return err
func QuotaIsFull(req paramsCheckQuota) (err error) {
	// body size check ，body的大小是否合法
	if req.contentSize > req.presetCollectionSize {
		return
	}
	// 当前col的大小和 预设的额度大小比较
	// TODO 返回存储超额的信息给前端
	if req.storageSize >= req.presetCollectionSize {
		return
	}
	//不满时
	nowTime := timeHelper.Date(timeHelper.Time(), time.DateTime)
	client := mongodb.DB.Mongo
	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		logger.NewIns.Error(context.Background(), "clientIpc.Ping err : ", err, "status", 500)
		//log.Println("ping panic is ", err)
	}
	fmt.Println("Connected to MongoDB!")
	defer func() {
		// 断开连接
		if err = client.Disconnect(context.Background()); err != nil {
			logger.NewIns.Error(context.Background(), "clientIpc.Disconnect err : ", err, "status", 500)
		}
		fmt.Println("Connection to MongoDB closed.")
	}()
	//1. 查询额度
	// TODO
	//queryQuota,quota_test_app_id,0,create table
	if req.quotaSizeByColName == 0 {
		//quotaColName := consts.QuotaCollectioPrefix + req.colName
		//quotaCol := client.Database(req.dbName).Collection(quotaColName)
		//r := QuotaSizeData{req.presetCollectionSize}
		//res, err := quotaCol.InsertOne(context.Background(), r)
		//if err != nil {
		//	log.Println(quotaColName+" init Insert quotaSize panic: ", err)
		//}
		//id := res.InsertedID
		//fmt.Printf(quotaColName+" Insert quotaSize pkId is : %#v\n", id)
	}
	//2. offline msg table,0,
	// insert into db
	offlineMsgCol := client.Database(req.dbName).Collection(req.colName)
	//插入
	bsonOne := OfflineMsgData{req.content, req.receiver, nowTime}
	res, err := offlineMsgCol.InsertOne(context.Background(), bsonOne)
	if err != nil {
		logger.NewIns.Error(context.Background(), req.colName+" Insert offlineMsgCol err : ", err, "status", 500)
		//log.Println(req.colName+" Insert offlineMsgCol panic: ", err)
	}
	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		logger.NewIns.Error(context.Background(), req.colName+" InsertedID is not a primitive.ObjectID err : ", "status", 500)
		//log.Println("InsertedID is not a primitive.ObjectID")
	}
	// Get the hexadecimal representation of the ObjectID
	hexID := insertedID.Hex()

	length := len(hexID)
	//扣减 配额
	if length > 0 {
		//update
		quotaColName := consts.QuotaCollectioPrefix + req.colName
		quotaCol := client.Database(req.dbName).Collection(quotaColName)
		//定额 - 已使用配额
		//r := QuotaCollection{req.quotaSizeByColName - req.storageSize}
		update := bson.M{"$set": bson.M{"quota": req.presetCollectionSize - req.storageSize}}
		filter := bson.M{}
		_, err = quotaCol.UpdateOne(context.Background(), filter, update)
		if err != nil {
			logger.NewIns.Error(context.Background(), quotaColName+" update quota err : ", err, "status", 500)
			//log.Println(quotaColName+" update quota panic is ", err)
		}
	}
	return nil
}

// db.collection.storageSize（）
//
// GetCollectionStorageSizeByAppId
//
//	@Description: 计算当前使用离线消息的文档占用大小
//	@param dbName
//	@param colName
//	@return storageSize
//	@return err
func GetCollectionStorageSizeByAppId(dbName, colName string) (storageSize uint32, err error) {
	client := mongodb.DB.Mongo
	statsResult := client.Database(dbName).RunCommand(context.Background(), bson.D{{"collStats", colName}})
	var result map[string]interface{}
	if err := statsResult.Decode(&result); err != nil {
		logger.NewIns.Error(context.Background(), colName+" statsResult err : ", err, "status", 500)
	}
	storageSizeInt32, ok := result["storageSize"].(int32)
	if !ok {
		logger.NewIns.Error(context.Background(), colName+" storageSize not found or not a int32 ", "status", 500)
	}
	storageSize = uint32(storageSizeInt32)
	return storageSize, nil
}

// GetCollectionQuotaByAppId
//
//	@Description: 查出额度表里面的额度值
//	@param dbName
//	@param colName
//	@return int
//	@return error

type QuotaCollection struct {
	Quota uint32 `bson:"quota"`
}

func collectionExists(client *mongo.Client, dbName, colName string) (bool, error) {
	// List collection names
	//bson.M{"name": collName}
	filter := bson.M{"name": colName}
	collections, err := client.Database(dbName).ListCollectionNames(context.Background(), filter)
	if err != nil {
		return false, err
	}

	// Check if the desired collection name is present
	for _, name := range collections {
		if strings.Compare(name, colName) == 0 {
			return true, nil
		}
	}

	return false, nil
}

// GetCollectionQuotaByAppId
//
//	@Description: 额度表
//	@param dbName
//	@param colName
//	@param presetCollectionSize
//	@return int
//	@return error
func GetCollectionQuotaByAppId(dbName, colName string, presetCollectionSize uint32) (uint32, error) {
	client := mongodb.DB.Mongo
	exists, err := collectionExists(client, dbName, colName)
	if err != nil {
		logger.NewIns.Error(context.Background(), colName+" insert data err : ", err, "status", 500)
	}
	collection := client.Database(dbName).Collection(colName)
	if exists {
		var result QuotaCollection
		if err := collection.FindOne(context.Background(), bson.D{}).Decode(&result); err != nil {
			logger.NewIns.Error(context.Background(), colName+" found document err : ", err, "status", 500)
		}
		log.Println("found document ", result)
		return result.Quota, nil
	} else {
		inData := QuotaCollection{
			Quota: presetCollectionSize,
		}
		_, err = collection.InsertOne(context.Background(), inData)
		if err != nil {
			logger.NewIns.Error(context.Background(), colName+" insert data err : ", err, "status", 500)
		}
		return presetCollectionSize, nil
	}
	return 0, nil
}

// getBodySize
//
//	@Description: 计算body字节大小
//	@param v
//	@return int
func getBodySize(body any) (uint32, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(body); err != nil {
		return 0, err
	}

	return uint32(buf.Len()), nil
}
