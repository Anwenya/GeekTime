package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}

	opts := options.Client().ApplyURI("mongodb://root:root@192.168.2.128:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	col := client.Database("webook").Collection("articles")
	// 插入一个文档
	insertRes, err := col.InsertOne(
		ctx,
		Article{
			Id:       1,
			Title:    "我的标题",
			Content:  "我的内容",
			AuthorId: 111,
		},
	)

	assert.NoError(t, err)
	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("id", oid)

	// 查询
	filter := bson.D{
		bson.E{
			Key:   "id",
			Value: 1,
		},
	}
	findRes := col.FindOne(ctx, filter)
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没查到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	// 更新一个
	updateFilter := bson.D{
		bson.E{Key: "id", Value: 1},
	}
	set := bson.D{
		bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}},
	}
	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	// 更新多个
	updateManyRes, err := col.UpdateMany(
		ctx,
		updateFilter,
		bson.D{
			bson.E{Key: "$set", Value: Article{Content: "新的内容"}},
		},
	)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	// 删除
	deleteFilter := bson.D{
		bson.E{Key: "id", Value: 1},
	}
	delRes, err := col.DeleteMany(ctx, deleteFilter)
	assert.NoError(t, err)
	t.Log("删除文档数量", delRes.DeletedCount)

}

type Article struct {
	Id         int64  `bson:"id,omitempty"`
	Title      string `bson:"title,omitempty"`
	Content    string `bson:"content,omitempty"`
	AuthorId   int64  `bson:"author_id,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
	CreateTime int64  `bson:"create_time,omitempty"`
	// 更新时间
	UpdateTime int64 `bson:"update_time,omitempty"`
}
