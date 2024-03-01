package startup

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func InitMongoDB(l logger.LoggerV1) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			l.Info(
				"mongodb",
				logger.Field{
					Key: "command",
					Val: evt.Command,
				},
			)
		},
	}
	opts := options.Client().ApplyURI(config.Config.DB.Mongo.Url).SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		l.Error("mongodb连接失败", logger.Error(err))
		panic(any(err))
	}
	mdb := client.Database("webook")

	err = initCollection(mdb)
	if err != nil {
		l.Error("mongodb初始化集合失败", logger.Error(err))
		panic(any(err))
	}
	return mdb
}

func initCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	col := mdb.Collection("articles")
	_, err := col.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{bson.E{Key: "id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{bson.E{Key: "author_id", Value: 1}},
			},
		},
	)
	if err != nil {
		return err
	}
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{bson.E{Key: "id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{bson.E{Key: "author_id", Value: 1}},
			},
		},
	)
	return err
}
