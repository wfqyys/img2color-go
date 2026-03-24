package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"img2color-go/app/internal/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBStorage MongoDB存储实现
type MongoDBStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
	enabled    bool
}

var (
	mongoInstance *MongoDBStorage
	mongoOnce     sync.Once
)

// NewMongoStorage 创建MongoDB存储
func NewMongoStorage(uri, database, collection string) (*MongoDBStorage, error) {
	var initErr error
	mongoOnce.Do(func() {
		// 创建客户端
		clientOptions := options.Client().ApplyURI(uri)
		clientOptions.SetConnectTimeout(5 * time.Second)

		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			initErr = fmt.Errorf("MongoDB连接失败: %w", err)
			logger.Error("MongoDB连接失败: %v", err)
			return
		}

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx, nil); err != nil {
			initErr = fmt.Errorf("MongoDB Ping失败: %w", err)
			logger.Error("MongoDB Ping失败: %v", err)
			return
		}

		mongoInstance = &MongoDBStorage{
			client:     client,
			collection: client.Database(database).Collection(collection),
			enabled:    true,
		}
		logger.Info("MongoDB连接成功: %s/%s", database, collection)
	})

	if initErr != nil {
		// 返回禁用状态的存储
		logger.Warn("MongoDB不可用，禁用持久化")
		return &MongoDBStorage{enabled: false}, nil
	}

	if mongoInstance == nil {
		return &MongoDBStorage{enabled: false}, nil
	}

	return mongoInstance, nil
}

// Save 保存颜色记录
func (s *MongoDBStorage) Save(ctx context.Context, url, color string) error {
	if !s.enabled || s.collection == nil {
		return nil
	}

	doc := bson.M{
		"url":       url,
		"color":     color,
		"created_at": time.Now(),
	}

	_, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		logger.Error("MongoDB保存记录失败: %v", err)
		return err
	}

	return nil
}

// Close 关闭连接
func (s *MongoDBStorage) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Disconnect(context.Background())
}

// IsEnabled 检查是否启用
func (s *MongoDBStorage) IsEnabled() bool {
	return s.enabled
}



