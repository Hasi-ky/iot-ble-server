package globalmemo

import (
	"github.com/coocood/freecache"
	cmap "github.com/streamrail/concurrent-map"
)

var (
	MemoCacheGw          cmap.ConcurrentMap //`[gw or gw + id] Queue`
	MemoCacheDev         cmap.ConcurrentMap //`[gw] Queue`
	MemoCacheScanTimeOut cmap.ConcurrentMap // 扫描超时处理专用

	MemoCacheService        cmap.ConcurrentMap //缓存服务专用, 仅用于确定服务是否为主服务
	MemoCacheServiceForChar cmap.ConcurrentMap //dev + 服务hanle == key | 

	//终端信息专用缓存
	BleFreeCacheDevInfo = freecache.NewCache(10 * 1024 * 1024)

	BleFreeCache = freecache.NewCache(10 * 1024 * 1024)
)
