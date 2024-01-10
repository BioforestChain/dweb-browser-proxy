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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
	"time"
)

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

func ProcessOfflineStorage(ctx context.Context, req *clientV1.IpcReq, hub *ws.Hub) error {
	content := req.Body
	//req.ClientID : "a.b.com,b.b.com,c.b.com"
	clientStr := req.ClientID
	clientArrList := stringsHelper.Explode(",", clientStr)

	appId := req.Header[consts.XDwebHostMMID][0]
	//colNameSrc, _ := g.Cfg().Get(ctx, "offlineMsgDb.colName")
	//colName = colNameSrc.String()
	//dbName := "local"
	dbNameSrc, _ := g.Cfg().Get(ctx, "offlineMsgDb.dbName")
	dbName := dbNameSrc.String()
	quotaColName := consts.QuotaCollectioPrefix + appId
	presetCollectionSize, _ := g.Cfg().Get(context.Background(), "presetCollectionSize.quotaSize")
	quotaSizeByColName, _ := GetCollectionQuotaByAppId(dbName, quotaColName, presetCollectionSize.Uint32())
	storageSize, _ := GetCollectionStorageSizeByAppId(dbName, appId)
	contentSize, err := getBodySize(content)
	if err != nil {
		log.Println("mongo.Connect panic is : ", err)
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
				log.Println("clientIpc.Send panic is", err)
				//return nil, err
			}
			//return resIpc, nil
		}
	}

	// Collection Quota check
	params := paramsCheckQuota{
		storageSize,
		quotaSizeByColName,
		presetCollectionSize.Uint32(),
		dbName,
		appId,
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
		log.Println("ping panic is ", err)
	}
	fmt.Println("Connected to MongoDB!")
	defer func() {
		// 断开连接
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
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
		log.Println(req.colName+" Insert offlineMsgCol panic: ", err)
	}
	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("InsertedID is not a primitive.ObjectID")
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
			log.Println(quotaColName+" update quota panic is ", err)
		}
	}
	fmt.Printf("InsertOne id's hexID is : %#v\n", hexID)
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
		log.Println(colName+" statsResult panic is", err)
	}
	storageSizeInt32, ok := result["storageSize"].(int32)
	if !ok {
		log.Println("storageSize not found or not a int32")
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
		log.Println(colName+" insert data panic is ", err)
	}
	collection := client.Database(dbName).Collection(colName)
	if exists {
		var result QuotaCollection
		if err := collection.FindOne(context.Background(), bson.D{}).Decode(&result); err != nil {
			log.Println("found document panic", err)
		}
		log.Println("found document ", result)
		return result.Quota, nil
	} else {
		inData := QuotaCollection{
			Quota: presetCollectionSize,
		}
		_, err = collection.InsertOne(context.Background(), inData)
		if err != nil {
			log.Println(colName+" insert data panic is ", err)
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
