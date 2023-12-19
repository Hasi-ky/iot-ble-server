package globalmemo

import (
	cmap "github.com/streamrail/concurrent-map"
)

var MemoCacheGw cmap.ConcurrentMap
var MemoCacheDev cmap.ConcurrentMap
