package packets

import (
	"iot-ble-server/dgram"
)

//UDP packets `PendFrame` 期待帧 | `PendCtrl` 期待编码格式 |
type JsonUdpInfo struct {
	MessageHeader    MessageHeader
	MessageBody      MessageBody
	MessageAppHeader MessageAppHeader
	MessageAppBody   MessageAppBody
	Rinfo            dgram.RInfo
	PendCtrl         int
}

type MessageType struct {
	UpMsg   UpMsg
	DownMsg DownMsg
}

//upStream
type UpMsg struct {
	WgHelloEvent   string
	ScanRespMsg    string
	ConnectRespMsg string
}

//downStream
type DownMsg struct {
	ScanMsgReq string
	ConnectMsg string
}

// version | length | sn | type | opType
type MessageHeader struct {
	Version           string
	LinkMessageLength string
	LinkMsgFrameSN    string
	LinkMsgType       string
	OpType            string
}

//| gwMac | moduleID | Status | Reason | ErrorCode |
type MessageBody struct {
	GwMac     string
	ModuleID  string
	ErrorCode string
	TLV       TLV
	MutiTLV   []TLV
}

// TotalLen | SN | CtrlField | FragOffset | AppMsgType | OpType
type MessageAppHeader struct {
	TotalLen   string
	SN         string
	CtrlField  string
	FragOffset string
	Type       string
	OpType     string
}

// errorCode | respSN | TLV
type MessageAppBody struct {
	ErrorCode    string
	RespondFrame string
	DevSum       string
	GwMac        string
	Reserve      string
	TLV          TLV
	MultiTLV      []TLV
}

//appmsg payload
type TLV struct {
	TLVMsgType string
	TLVLen     string
	TLVPayload TLVFeature
}

//tlv message content
type TLVFeature struct {
	AnnounceContent       string
	DevMac                string
	GwType                string
	GwSN                  string
	GwMac                 string
	IotModuleType         string
	IotModuleSN           string
	IotModuleMac          string
	IotModuleId           string
	IotModuleStatus       string
	IotModuleChangeReason string
	ErrorCode             string
	ServiceUUID           string
	CharacterUUID         string
	MajorID               string
	MinorID               string
	MeasurePower          string
	RSSI                  string
	TimeStamp             string
	ReserveOne            string
	ReserveTwo            string
	ReserveThree          string
	AddrType              string
	OpType                string
	ADType                string
	ConnStatus            string
	CharHandle            string
	CCCDHandle            string
	ConnHandle            string
	DescriptorHandle      string
	CharHandleValue       string
	CCCDHandleValue       string
	Handle                string
	EndHandle             string
	Event                 string
	DisConnReason         string
	FeatureHandle         string //特征handle
	FeatureSum            string
	FeatureCfg            string //特征属性配置
	ConnInterval          string
	ConnLatency           string
	ConnTimeout           string
	MTUSize               string
	ServiceSum            string
	ScanType              string
	ScanStatus            string
	ScanAble              string
	ScanPhys              string
	ScanInterval          string
	ScanWindow            string
	ScanTimeout           string
	ServiceHandle         string
	StartHandle           string
	SupervisionWindow     string
	ParaLength            string //参数值长度
	ParaValue             string //应答的参数值
	PHY                   string
	Port                  string
	Primary               string
	Properties            string
	NoticeType            string       //通告类型
	NoticeContent         ManufactData //通告内容
	TLVReserve            []TLV        //携带TLV
}

//广播报文中的厂商数据 -------
type ManufactData struct {
	Length string
	Data   Data
}

type Data struct {
	AdType string
	AdData ADData
}

type ADData struct {
	CompanyID string
	CompData  CompanyData
}

type CompanyData struct {
	MsgType    string
	Device     DeviceID
	DeviceData string
}

type DeviceID struct {
	DevType string
	Addr    string
}
