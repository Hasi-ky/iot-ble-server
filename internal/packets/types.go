package packets

type AddressType uint8
type ConnType uint8
type ScanType uint8
type State uint8
type PHY uint8
type Version string
type ModuleID uint16
type GwACK string

type AppMsgType string
type MAC [6]byte

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

//`mac版本`
const (
	Version3 string = "0103"
)

//`三类消息`
const (
	ChannelControl  string = "01"
	GatewayManager  string = "02"
	TerminalManager string = "03"
)

//`通道` | `网关` | `终端` 三类消息
const (
	Hello string = "0101"

	GatewayDevInfo        string = "0201"
	IotModuleRset         string = "0202"
	IotModuleStatusChange string = "0203"

	BleRequest              string = "0301"
	BleConfirm              string = "0301"
	BleResponse             string = "0302"
	BleGetConnDevList       string = "0303"
	BleCharacteristicNotice string = "0304"
	BleBoardcast            string = "0305"
	BleTerminalEvent        string = "0306"
)

//`TLV`
const (
	TLVIBeaconMsg    string = "0101"
	TLVBeaconMsg     string = "0106"
	TLVRFIDMsg       string = "0108"
	TLVGeneralIOTMsg string = "0109"

	TLVGatewayTypeMsg           string = "0201"
	TLVGatewaySNMsg             string = "0202"
	TLVGatewayMACMsg            string = "0203"
	TLVIotModuleMsg             string = "0204"
	TLVIotModuleSNMsg           string = "0205"
	TLVIotModuleMACMsg          string = "0206"
	TLVGatewayDescribeMsg       string = "0207"
	TLVIotModuleDescribeMsg     string = "0208"
	TLVIotModuleEventMsg        string = "0209"
	TLVServiceMsg               string = "020A"
	TLVCharacteristicMsg        string = "020B"
	TLVDeviceListMsg            string = "020C"
	TLVNotifyMsg                string = "020D"
	TLVScanMsg                  string = "020E"
	TLVScanRespMsg              string = "020F"
	TLVConnectMsg               string = "0210"
	TLVConnectRespMsg           string = "0212"
	TLVMainServiceReqMsg        string = "0213"
	TLVMainServiceRespMsg       string = "0214"
	TLVCharReqMsg               string = "0215"
	TLVCharRespMsg              string = "0216"
	TLVCharConfReqMsg           string = "0217"
	TLVCharConfRespMsg          string = "0218"
	TLVCharOptReqMsg            string = "0219"
	TLVCharOptRespMsg           string = "021A"
	TLVBroadcastMsg             string = "021B"
	TLVDisconnectMsg            string = "021C"
	TLVCharDescribeMsg          string = "021E"
	TLVMainServiceByUUIDReqMsg  string = "021F"
	TLVMainServiceByUUIDRespMsg string = "0220"
)

// `Message`
const (
	Request         string = "01"
	RequireWithResp string = "02"
	Response        string = "03"
)

// ``

// // version | length | sn | type | opType | gwMac | moduleID |
// type MsgHead struct {
// 	Version Version     `json:"version"`
// 	Len     uint16      `json:"len"`
// 	SN      FrameSN          `json:"sn"`
// 	Type    LinkMsgType `json:"type"`
// 	OpType  OpType      `json:"opType"`
// }

// type GatewayInfo struct {
// 	MAC      MAC      `json:"mac"`
// 	ModuleID ModuleID `json:"moduleID"`
// }

// // 	TotalLen | SN| CtrlField | FragOffset | string | OpType
// type AppMsgHead struct {
// 	TotalLen   uint16
// 	SN         FrameSN
// 	CtrlField  uint16
// 	FragOffset uint16
// 	Type       AppMsgType
// 	OpType     OpType
// }

// // string | Length | Reserved0 | AddressType | MAC | ConnType | Reserved1 | ScanType | Reserved2
// // ScanInterval | ScanWindow | ScanTimeout | ConnInterval | ConnInterval | connLatency | supTimeout
// type ConnInfo struct {
// 	string   string
// 	Length       uint16
// 	Reserved0    uint8
// 	AddressType  AddressType
// 	MAC          MAC
// 	ConnType     ConnType
// 	Reserved1    uint8
// 	ScanType     ScanType
// 	Reserved2    uint8
// 	ScanInterval uint16
// 	ScanWindow   uint16
// 	ScanTimeout  uint16
// 	ConnInterval uint16
// 	ConnLatency  uint16
// 	SupTimeout   uint16
// }

// // string | Length | Reserved0 | AddressType | MAC | ConnType | Reserved1 | ScanType | Reserved2
// // ScanInterval | ScanWindow | ScanTimeout | ConnInterval | ConnInterval | connLatency | supTimeout
// type ConnAckInfo struct {
// 	string   string
// 	Length       uint16
// 	MAC          MAC
// 	ErrorCode    ErrorCode
// 	State        State
// 	PHY          PHY
// 	ConnHandle   uint16
// 	ConnInterval uint16
// 	connLatency  uint16
// 	SupTimeout   uint16
// 	MTU          uint16
// }

// type IOTHead struct {
// 	Type   string `json:"type"`
// 	OpType OpType     `json:"opType"`
// }

// type ConnectReqMsg struct {
// 	MsgHead     MsgHead
// 	GatewayInfo GatewayInfo
// 	AppMsgHead  AppMsgHead
// 	ConnInfo    ConnInfo
// }

// type ConnectAckMsg struct {
// 	MsgHead     MsgHead
// 	GatewayInfo GatewayInfo
// 	AppMsgHead  AppMsgHead
// 	ConnAckInfo ConnAckInfo
// }
