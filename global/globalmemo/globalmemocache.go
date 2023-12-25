package globalmemo

import (
	"github.com/coocood/freecache"
	cmap "github.com/streamrail/concurrent-map"
)

var (
	MemoCacheGw  cmap.ConcurrentMap //`[gw or gw + id] Queue`
	MemoCacheDev cmap.ConcurrentMap //`[gw] Queue`
	BleFreeCache = freecache.NewCache(10 * 1024 * 1024)
)
