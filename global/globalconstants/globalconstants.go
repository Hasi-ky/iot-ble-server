package globalconstants

import (
	"iot-ble-server/global/globalstruct"
	"time"
)

var (
	CtrlAllMsg                    = 0
	CtrlLinkedMsgHeader           = 1
	CtrlLinkedMsgHeadWithBoy      = 3
	CtrlLinkedMsgWithMsgAppHeader = 7
	EncodeInsertIndex             = 4
	EncodeInsertLen               = 4
	MaxMessageFourLimit           = 0x3f3f3f3f //消息数上限 4 字节
	MaxMessageTwoLimit            = 0x3f3f     //消息数上线 2字节

	GwSocketCachePrefix      = "gw:server:socket"
	BleDevCacheMessagePrefix = "ble:server:dev" //下行消息前缀，后缀终端信息
	BleDevCacheConnPrefix    = "ble:server:dev:conn"

	BleDevInfoCachePrefix  = "ble:server:infodev" //存储的是当前终端对应最优网关信息，然后就可以通过这个最优网关信息进行数连接
	CacheSeparator         = ":"
	GwIotModuleCachePrefie = "gw:server:iotstatus"
	BleSendDownQueue       = "ble:dev:down:queue" //后接终端mac
	BleRedisSNTransfer     = "ble:sn:transfer"
	GwRedisSNTransfer      = "gw:sn:transfer" //网关的transfer

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
	BYTE_STR_EIGHT   int = 8
	BYTE_STR_TWELVE  int = 12
	BYTE_STR_SIXTEEN int = 16
)

//消息传递出接口
var (
	ConnectionInfoChan chan globalstruct.ResultMessage
)

//Http Code
const (
	HTTP_CODE_EXCEPTION = -1
	HTTP_CODE_SUCCESS   = 0
	HTTP_CODE_ERROR     = 1
)

//HTTP MESSAGE
const (
	HTTP_MESSAGE_SUCESS           = "success"
	HTTP_MESSAGE_FAILED           = "fail"
	HTTP_MEESAGE_ABSENCE_TERMINAL = "终端对应数据缺失"
	HTTP_MEESAGE_CACHE_EXCEPITON  = "缓存库异常"
	HTTP_ERROR_DATA_ABSENCE       = "部分数据缺失错误"
)

//普通错误
const (
	ERROR_DATA_FORMAT       = "数据格式错误"
	ERROR_DATA_CHANGE       = "数据转化错误"
	ERROR_CACHE_ABSENCE     = "缓存库缺少对应数据"
	ERROR_DATA_GENERATEDOWN = "下行数据生成失败"
	ERROR_DATA_TIMEOUT      = "超时等待"

	ERROR_CACHE_EXCEPTION = "缓存库出现异常"
)

//终端帧校验
const (
	TERMINAL_CORRECT   int = 0
	TERMINAL_ERROR     int = 1
	TERMINAL_EXCEPTION int = -1
)

//服务与特征
const (
	DEV_SERVICE           = "service"
	DEV_CHARACTER         = "character"
	DEV_CHARACTER_DEFAULT = 0x3f3f //二字节默认值
)

//MQTT消息队列
const (
	TopicV3GatewayNetworkIn    = "v3/gateway/network/in"
	TopicV3GatewayNetworkInAck = "v3/gateway/network/in/ack"
	TopicV3GatewayTelemetry    = "v3/gateway/telemetry"
	TopicV3GatewayRPC          = "v3/gateway/rpc"
)

//暂定连接参数
const (
	BLE_CONN_DEV_ADDR_TYPE = 1
	BLE_CONN_OPTYPE        = 1
	BLE_CONN_SCAN_TYPE     = 1
	BLE_CONN_SCAN_INTERVAL = 10
	BLE_CONN_SCAN_WINDOWS  = 10
	BLE_CONN_SCAN_TIMEOUT  = 0
	BLE_CONN_INTERVAL      = 6
	BLE_CONN_LATENCY       = 0
	BLE_CONN_TIMEOUT       = 6
)
