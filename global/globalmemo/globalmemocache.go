package globalmemo

import (
	"github.com/coocood/freecache"
	cmap "github.com/streamrail/concurrent-map"
)

var (
	MemoCacheGw  cmap.ConcurrentMap //`[gw or gw + id] Queue`
	MemoCacheDev cmap.ConcurrentMap //`[gw] Queue`
	//终端信息专用缓存
	BleFreeCacheDevInfo = freecache.NewCache(10 * 1024 * 1024)

	BleFreeCache = freecache.NewCache(10 * 1024 * 1024)
)
