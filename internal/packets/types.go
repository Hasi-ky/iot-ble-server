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
	Version3 string = "0103"
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
	Hello string = "0101"
)

const (
	IBeaconMsg               string = "0101"
	BeaconMsg                string = "0106"
	RFIDMsg                  string = "0108"
	GeneralIOTMsg            string = "0109"
	GatewayTypeMsg           string = "0201"
	GatewaySNMsg             string = "0202"
	GatewayMACMsg            string = "0203"
	GatewayModuleMsg         string = "0204"
	GatewayModuleSNMsg       string = "0205"
	GatewayModuleMACMsg      string = "0206"
	GatewayDescribeMsg       string = "0207"
	GatewayModuleDescribeMsg string = "0208"
	GatewayModuleEventMsg    string = "0209"
	ServiceMsg               string = "020A"
	CharacteristicMsg        string = "020B"
	DeviceListMsg            string = "020C"
	NotifyMsg                string = "020D"
	ScanMsg                  string = "020E"
	ScanRespMsg              string = "020F"
	ConnectMsg               string = "0210"
	ConnectRespMsg           string = "0212"
	MainServiceReqMsg        string = "0213"
	MainServiceRespMsg       string = "0214"
	CharReqMsg               string = "0215"
	CharRespMsg              string = "0216"
	CharConfReqMsg           string = "0217"
	CharConfRespMsg          string = "0218"
	CharOptReqMsg            string = "0219"
	CharOptRespMsg           string = "021A"
	BroadcastMsg             string = "021B"
	DisconnectMsg            string = "021C"
	CharDescribeMsg          string = "021E"
	MainServiceByUUIDReqMsg  string = "021F"
	MainServiceByUUIDRespMsg string = "0220"
)

const (
	AppReq         string = "0301"
	AppResp        string = "0302"
	ConnDevList    string = "0303"
	AppCharConfReq string = "0304"
	AppBroadcast   string = "0305"
	AppEvent       string = "0306"
)

const (
	Require         string = "01"
	RequireWithResp string = "02"
	Response        string = "03"
)
const (
	ChannelControl  string = "01"
	GatewayManager  string = "02"
	TerminalManager string = "03"
)

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
