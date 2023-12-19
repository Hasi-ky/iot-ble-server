package packets

import (
	"iot-ble-server/dgram"
)

//UDP packets
type JsonUdpInfo struct {
	MessageHeader    MessageHeader
	MessageBody      MessageBody
	MessageAppHeader MessageAppHeader
	MessageAppBody   MessageAppBody
	Rinfo            dgram.RInfo
}

type MessageType struct {
	UpMsg   UpMsg
	DownMsg DownMsg
}

//upStream
type UpMsg struct {
	WgHelloEvent   IOTMsgType
	ScanRespMsg    IOTMsgType
	ConnectRespMsg IOTMsgType
}

//downStream
type DownMsg struct {
	ScanMsgReq IOTMsgType
	ConnectMsg IOTMsgType
}

// version | length | sn | type | opType
type MessageHeader struct {
	Version           string
	LinkMessageLength string
	LinkMsgFrameSN    string
	LinkMsgType       string
	OpType            OpType
}

//| gwMac | moduleID |
type MessageBody struct {
	GwMac    string
	ModuleID string
}

// TotalLen | SN| CtrlField | FragOffset | AppMsgType | OpType
type MessageAppHeader struct {
	TotalLen   string
	SN         string
	CtrlField  string
	FragOffset string
	Type       string
	OpType     OpType
}

// errorCode | respSN | TLV
type MessageAppBody struct {
	ErrorCode    ErrorCode
	RespondFrame uint32
	TLV          TLV
}

//appmsg payload
type TLV struct {
	TLVMsgType string
	TLVLen     uint16
	TLVPayload TLVFeature
}

//tlv message content
type TLVFeature struct {
	DevMac          MAC
	ErrorCode       ErrorCode
	UUID            string
	MajorID         string
	MinorID         string
	MeasurePower    int8
	RSSI            int8
	TimeStamp       string
	ReserveOne      uint8
	ReserveTwo      uint8
	ReserveThree    uint16
	AddrType        uint8
	OpType          uint8
	ADType          uint8
	ScanType        uint8
	ScanStatus      uint8
	ScanAble        uint8
	ScanPhys        uint8
	ScanInterval    uint16
	ScanWindow      uint16
	ScanTimeout     uint16
	ConnStatus      uint8
	ServiceHandle   uint16
	CharHandle      uint16
	ValueHandle     uint16
	CCCDHandle      uint16
	StartHandle     uint16
	EndHandle       uint16
	ConnHandle      uint16
	DisConnReason   uint16
	FeatureHandle   uint16 //特征handle
	FeatureSum      uint16
	FeatureCfg      uint8
	ConnInterval    uint16
	ConnLatency     uint16
	ConnTimeout     uint16
	MTUSize         uint16
	ServiceSum      uint16
	ParaLength      uint16
	ParaValue       string
	PHY             uint8
	AnnounceContent string
	ReserveTLV      []TLVFeature //保留
}
