package globalconstants

import "time"

var (
	CtrlAllMsg                    = 0
	CtrlLinkedMsgHeader           = 1
	CtrlLinkedMsgHeadWithBoy      = 3
	CtrlLinkedMsgWithMsgAppHeader = 7
	EncodeInsertIndex             = 4
	EncodeInsertLen               = 4
	MaxMessageFourLimit           = 0x3f3f3f3f //消息数上限 4 字节
	MaxMessageTwoLimit            = 0x3f3f     //消息数上线 2字节

	GwSocketCachePrefix    = "gw:server:socket:"
	BleDevCachePrefix      = "ble:server:dev:"
	BleDevInfoCachePrefix  = "ble:server:infodev"
	CacheSeparator         = ":"
	GwIotModuleCachePrefie = "gw:server:iotstatus:"
	BleSendDownQueue       = "ble:dev:down:queue:" //后接终端mac
	BleRedisSNTransfer     = "ble:sn:transfer:"
	GwRedisSNTransfer      = "gw:sn:transfer:" //网关的transfer

	LimitMessageTime = time.Hour //消息截至过期时间
	TTLDuration      = time.Hour * 7 * 24
	AgingTTLDuration = time.Second * 30
)

const (
	JudgeGetError = iota //取值错误
	JudgeGetNil          //空
	JudgeGetRes          //取到
)

//暂时封装
var (
	CtrlField  = "0000"
	FragOffset = "0000"
)

//字节对应字符串长度
const (
	BYTE_STR_ONE     int = 1
	BYTE_STR_TWO     int = 2
	BYTE_STR_FOUR    int = 4
	BYTE_STR_TWELVE  int = 12
	BYTE_STR_SIXTEEN int = 16
)
