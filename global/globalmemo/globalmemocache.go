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
	MemoCacheServiceForChar cmap.ConcurrentMap //dev + 服务uuid 服务hanle == key |  field : character uuid + handle | value : characterNode |

	//广播所需以及其下行连接 `key = devMac |  map[module + id] TerminalInfo `
	MemoCacheDownPriority cmap.ConcurrentMap

	//上下行交互专用
	BleFreeCacheUpDown = freecache.NewCache(10 * 1024 * 1024) //纯缓存使用

	//终端信息专用缓存
	BleFreeCacheDevInfo = freecache.NewCache(10 * 1024 * 1024)

	//终端连接信息专用
	BleFreeCacheDevConnInfo = freecache.NewCache(10 * 1024 * 1024)

	BleFreeCache = freecache.NewCache(10 * 1024 * 1024)
)
