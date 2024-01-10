package mongodb

import (
	"context"
	"fmt"
	_ "fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)

// mgo
// @Description:
type mgo struct {
	database   string //要连接的数据库
	collection string //要连接的集合
}

// NewMgo
//
//	@Description:
//	@param database
//	@param collection
//	@return *mgo
func NewMgo(database, collection string) *mgo {

	return &mgo{
		//uri,
		database,
		collection,
	}
}

// FindOne
//
//	@Description: 查询单个
//	@receiver m
//	@param key
//	@param value
//	@return *mongo.SingleResult
func (m *mgo) FindOne(key string, value interface{}) *mongo.SingleResult {
	client := DB.Mongo
	collection, _ := client.Database(m.database).Collection(m.collection).Clone()
	//collection.
	filter := bson.D{{key, value}}
	singleResult := collection.FindOne(context.TODO(), filter)
	return singleResult
}

// InsertOne
//
//	@Description: 插入单个
//	@receiver m
//	@param value
//	@return *mongo.InsertOneResult
func (m *mgo) InsertOne(value interface{}) *mongo.InsertOneResult {
	client := DB.Mongo
	collection := client.Database(m.database).Collection(m.collection)
	insertResult, err := collection.InsertOne(context.TODO(), value)
	if err != nil {
		fmt.Println(err)
	}
	return insertResult
}

// CollectionCount
//
//	@Description:  查询集合里有多少数据
//	@receiver m
//	@return string
//	@return int64
func (m *mgo) CollectionCount() (string, int64) {
	client := DB.Mongo
	collection := client.Database(m.database).Collection(m.collection)
	name := collection.Name()
	size, _ := collection.EstimatedDocumentCount(context.TODO())
	return name, size
}

// CollectionDocuments
//
//	@Description: filter := bson.D{{key,value}} 按选项查询集合 Skip 跳过 Limit 读取数量 sort 1 ，-1 . 1 为最初时间读取 ， -1 为最新时间读取
//	@receiver m
//	@param Filter
//	@param Skip
//	@param Limit
//	@param sort
//	@return cur
//	@return err
func (m *mgo) CollectionDocuments(Filter bson.M, Skip, Limit int64, sort int) (cur *mongo.Cursor, err error) {
	client := DB.Mongo
	collection := client.Database(m.database).Collection(m.collection)

	SORT := bson.D{{"_id", sort}}
	findOptions := options.Find().SetSort(SORT).SetLimit(Limit).SetSkip(Skip)

	//findOptions.SetLimit(i)
	temp, err := collection.Find(context.Background(), Filter, findOptions)
	if err != nil {
		log.Println("CollectionDocuments collection.Find err is", err)
		return nil, err
	}
	return temp, nil
}

// ParsingId
//
//	@Description: 获取集合创建时间和编号
//	@receiver m
//	@param result
//	@return time.Time
//	@return uint64
func (m *mgo) ParsingId(result string) (time.Time, uint64) {
	temp1 := result[:8]
	timestamp, _ := strconv.ParseInt(temp1, 16, 64)
	dateTime := time.Unix(timestamp, 0) //这是截获情报时间 时间格式 2019-04-24 09:23:39 +0800 CST
	temp2 := result[18:]
	count, _ := strconv.ParseUint(temp2, 16, 64) //截获情报的编号
	return dateTime, count
}

// DeleteAndFind
//
//	@Description: 删除doc和查询doc
//	@receiver m
//	@param key
//	@param value
//	@return int64
//	@return *mongo.SingleResult
func (m *mgo) DeleteAndFind(key string, value interface{}) (int64, *mongo.SingleResult) {
	client := DB.Mongo

	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}
	singleResult := collection.FindOne(context.TODO(), filter)
	DeleteResult, err := collection.DeleteOne(context.TODO(), filter, nil)
	if err != nil {
		fmt.Println("删除时出现错误，你删不掉的~")
	}
	return DeleteResult.DeletedCount, singleResult
}

// Delete
//
//	@Description: 删除doc
//	@receiver m
//	@param key
//	@param value
//	@return int64
func (m *mgo) Delete(key string, value interface{}) int64 {
	client := DB.Mongo
	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}
	count, err := collection.DeleteOne(context.TODO(), filter, nil)
	if err != nil {
		fmt.Println(err)
	}
	return count.DeletedCount

}

// DeleteMany
//
//	@Description: 删除多个
//	@receiver m
//	@param key
//	@param value
//	@return int64
func (m *mgo) DeleteMany(key string, value interface{}) int64 {
	client := DB.Mongo
	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}

	count, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
	}
	return count.DeletedCount
}
