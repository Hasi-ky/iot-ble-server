package globalconstants

import "time"

var (
	CtrlAllMsg                    = 0
	CtrlLinkedMsgHeader           = 1
	CtrlLinkedMsgHeadWithBoy      = 3
	CtrlLinkedMsgWithMsgAppHeader = 7
	EncodeInsertIndex             = 4
	EncodeInsertLen               = 4
	GwSocketCachePrefix           = "ble:server:socket:"
	BleDevCachePrefix             = "ble:server:dev:"
	BleDevInfoCachePrefix         = "ble:server:infodev:"
	CacheSeparator                = ":"
	GwIotModuleCachePrefie        = "ble:gwIot:"

	TTLDuration = time.Hour * 7 * 24
)
