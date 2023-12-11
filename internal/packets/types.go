package packets

type OpType uint8
type AddressType uint8
type ConnType uint8
type ScanType uint8
type State uint8
type PHY uint8
type Version uint16
type ModuleID uint16
type LinkMsgType uint16
type IOTMsgType uint16
type AppMsgType uint16
type SN uint32
type MAC [6]byte

const (
	Version3 Version = 0x0103
)

const (
	Public AddressType = 0
	Random AddressType = 1
)

const (
	Passive ScanType = 0
	Active  ScanType = 1
)

const (
	Disconnect ConnType = 0
	Connect    ConnType = 1
)

const (
	Disconnected State = 0
	Connected    State = 1
)

const (
	PHY1M    PHY   = 0x01
	PHYCoded State = 0x04
)

const (
	Hello LinkMsgType = 0x0101
)

const (
	IBeaconMsg               IOTMsgType = 0x0101
	BeaconMsg                IOTMsgType = 0x0106
	RFIDMsg                  IOTMsgType = 0x0108
	GeneralIOTMsg            IOTMsgType = 0x0109
	GatewayTypeMsg           IOTMsgType = 0x0201
	GatewaySNMsg             IOTMsgType = 0x0202
	GatewayMACMsg            IOTMsgType = 0x0203
	GatewayModuleMsg         IOTMsgType = 0x0204
	GatewayModuleSNMsg       IOTMsgType = 0x0205
	GatewayModuleMACMsg      IOTMsgType = 0x0206
	GatewayDescribeMsg       IOTMsgType = 0x0207
	GatewayModuleDescribeMsg IOTMsgType = 0x0208
	GatewayModuleEventMsg    IOTMsgType = 0x0209
	ServiceMsg               IOTMsgType = 0x020A
	CharacteristicMsg        IOTMsgType = 0x020B
	DeviceListMsg            IOTMsgType = 0x020C
	NotifyMsg                IOTMsgType = 0x020D
	ScanMsg                  IOTMsgType = 0x020E
	ScanRespMsg              IOTMsgType = 0x020F
	ConnectMsg               IOTMsgType = 0x0210
	ConnectRespMsg           IOTMsgType = 0x0212
	MainServiceReqMsg        IOTMsgType = 0x0213
	MainServiceRespMsg       IOTMsgType = 0x0214
	CharReqMsg               IOTMsgType = 0x0215
	CharRespMsg              IOTMsgType = 0x0216
	CharConfReqMsg           IOTMsgType = 0x0217
	CharConfRespMsg          IOTMsgType = 0x0218
	CharOptReqMsg            IOTMsgType = 0x0219
	CharOptRespMsg           IOTMsgType = 0x021A
	BroadcastMsg             IOTMsgType = 0x021B
	DisconnectMsg            IOTMsgType = 0x021C
	CharDescribeMsg          IOTMsgType = 0x021E
	MainServiceByUUIDReqMsg  IOTMsgType = 0x021F
	MainServiceByUUIDRespMsg IOTMsgType = 0x0220
)

const (
	AppReq         AppMsgType = 0x0301
	AppResp        AppMsgType = 0x0302
	ConnDevList    AppMsgType = 0x0303
	AppCharConfReq AppMsgType = 0x0304
	AppBroadcast   AppMsgType = 0x0305
	AppEvent       AppMsgType = 0x0306
)

const (
	Require         OpType = 1
	RequireWithResp OpType = 2
	Response        OpType = 3
)

// version | length | sn | type | opType | gwMac | moduleID |
type MsgHead struct {
	Version Version     `json:"version"`
	Len     uint16      `json:"len"`
	SN      SN          `json:"sn"`
	Type    LinkMsgType `json:"type"`
	OpType  OpType      `json:"opType"`
}

type GatewayInfo struct {
	MAC      MAC      `json:"mac"`
	ModuleID ModuleID `json:"moduleID"`
}

// 	TotalLen | SN| CtrlField | FragOffset | AppMsgType | OpType
type AppMsgHead struct {
	TotalLen   uint16
	SN         SN
	CtrlField  uint16
	FragOffset uint16
	Type       AppMsgType
	OpType     OpType
}

// IOTMsgType | Length | Reserved0 | AddressType | MAC | ConnType | Reserved1 | ScanType | Reserved2
// ScanInterval | ScanWindow | ScanTimeout | ConnInterval | ConnInterval | connLatency | supTimeout
type ConnInfo struct {
	IOTMsgType   IOTMsgType
	Length       uint16
	Reserved0    uint8
	AddressType  AddressType
	MAC          MAC
	ConnType     ConnType
	Reserved1    uint8
	ScanType     ScanType
	Reserved2    uint8
	ScanInterval uint16
	ScanWindow   uint16
	ScanTimeout  uint16
	ConnInterval uint16
	ConnLatency  uint16
	SupTimeout   uint16
}

// IOTMsgType | Length | Reserved0 | AddressType | MAC | ConnType | Reserved1 | ScanType | Reserved2
// ScanInterval | ScanWindow | ScanTimeout | ConnInterval | ConnInterval | connLatency | supTimeout
type ConnAckInfo struct {
	IOTMsgType   IOTMsgType
	Length       uint16
	MAC          MAC
	ErrorCode    ErrorCode
	State        State
	PHY          PHY
	ConnHandle   uint16
	ConnInterval uint16
	connLatency  uint16
	SupTimeout   uint16
	MTU          uint16
}

type IOTHead struct {
	Type   IOTMsgType `json:"type"`
	OpType OpType     `json:"opType"`
}

type ConnectReqMsg struct {
	MsgHead     MsgHead
	GatewayInfo GatewayInfo
	AppMsgHead  AppMsgHead
	ConnInfo    ConnInfo
}

type ConnectAckMsg struct {
	MsgHead     MsgHead
	GatewayInfo GatewayInfo
	AppMsgHead  AppMsgHead
	ConnAckInfo ConnAckInfo
}
