package bleudp

import (
	"fmt"
	"testing"

	"github.com/coocood/freecache"
)

func TestMain(t *testing.T) {

	var tempBleFreeCache = freecache.NewCache(10 * 1024 * 1024)
	tempBleFreeCache.Set([]byte("jianlai"), []byte("jianlai"), 0)
	tempBleFreeCache.Set([]byte("1"), []byte("1"), 0)
	tempBleFreeCache.Set([]byte("3"), []byte("jia2nlai"), 0)
	tempBleFreeCache.Set([]byte("4"), []byte("jia3nlai"), 0)
	tempBleFreeCache.Set([]byte("5"), []byte("jian2lai"), 0)
	iter := tempBleFreeCache.NewIterator()
	for ;; {
		tempRes := iter.Next()
		if tempRes == nil {
			break
		}
		fmt.Println(tempRes.Key, tempRes.Value)
	}
	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr:     "33.33.33.244:6381",
	// 	DB:       0,
	// 	Password: "",
	// })
	// ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cf()
	// _, err := redisClient.Ping(ctx).Result()
	// redisClient.Set(ctx, "jianlai", "chenchen", -1)
	// if err != nil {
	// 	fmt.Println("无法连接到redis服务器", err)
	// 	return
	// }
	// fmt.Println("成功连接")
}
