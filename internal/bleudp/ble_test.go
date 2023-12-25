package bleudp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestMain(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "33.33.33.244:6381",
		DB:       0,
		Password: "",
	})
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	defer cf()
	_, err := redisClient.Ping(ctx).Result()
	redisClient.Set(ctx, "jianlai", "chenchen", -1)
	if err != nil {
		fmt.Println("无法连接到redis服务器", err)
		return
	}
	fmt.Println("成功连接")
}
