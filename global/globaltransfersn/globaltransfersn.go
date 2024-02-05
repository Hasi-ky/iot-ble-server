package globaltransfersn

import "sync"

//维护终端、网关下行消息数量
var TransferSN = struct {
	sync.RWMutex
	SN map[string]int
}{SN: make(map[string]int)}