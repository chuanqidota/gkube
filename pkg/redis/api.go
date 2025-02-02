package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gkube/config"

	"time"
)

type ManagerRedis struct {
	Client *redis.Client
}

// NewManagerRedis
//
//	@Description: 构造函数
//	@param addr
//	@param password
//	@param db
//	@return *ManagerRedis
func NewManagerRedis(addr, password string, db int) *ManagerRedis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &ManagerRedis{Client: client}
}

var RedisClient *redis.Client

func Init() {
	m := NewManagerRedis(config.Conf.Redis.Addr, config.Conf.Redis.Password, config.Conf.Redis.DB)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	RedisClient = m.Client
	// 检查连接是否正常
	_, err := m.Client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("redis连接失败：%v\n", err.Error())
		return
	}
	fmt.Printf("redis连接成功")
}

// Set
//
//	@Description: 设置字符串
//	@param key
//	@param value
//	@param expiration
//	@return error
func (m *ManagerRedis) Set(key string, value any, expiration time.Duration) error {
	if err := m.Client.Set(context.Background(), key, value, expiration).Err(); err != nil {
		return err
	} else {
		return nil
	}
}

// Get
//
//	@Description: 获取字符串
//	@param key
//	@return string
//	@return error
func (m *ManagerRedis) Get(key string) (string, error) {
	result, err := m.Client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

// SetMap
//
//	@Description: 设置value为map
//	@param key
//	@param value
//	@return error
func (m *ManagerRedis) SetMap(key string, value map[string]any) error {
	// 使用 HSet 设置值为 map 类型的哈希
	err := m.Client.HSet(context.Background(), key, value).Err()
	if err != nil {
		fmt.Printf("设置值错误：%s\n", err.Error())
		return err
	}
	return nil
}

// GetMap
//
//	@Description: 获取value为map
//	@param key
//	@return map[string]any
//	@return error
func (m *ManagerRedis) GetMap(key string) (map[string]any, error) {
	// 获取哈希类型的值
	val, err := m.Client.HGetAll(context.Background(), key).Result()
	if err != nil {
		fmt.Printf("获取值错误：%s", err.Error())
		return nil, err
	}
	// 将值转换为 map[string]string 类型
	result := make(map[string]any)
	for field, value := range val {
		result[field] = value
	}
	return result, nil
}

// DeleteKey
//
//	@Description: 删除key
//	@receiver m
//	@param key
//	@return error
func (m *ManagerRedis) DeleteKey(key string) error {
	// 删除指定的键
	err := m.Client.Del(context.Background(), key).Err()
	if err != nil {
		fmt.Printf("删除键：%s", err.Error())
		return err
	}
	return nil
}
