package globalstruct

import (
	"time"
)

//网络信息接口
type SocketInfo struct {
	ID         uint      `json:"id"  db:"id"`
	Mac        string    `json:"GwMac"  db:"gwmac"`
	Family     string    `json:"family"  db:"family"` //源socket的IP协议
	IPAddr     string    `json:"IPAddr" db:"ipaddr"`  //源socket的IP
	IPPort     int       `json:"IPPort"  db:"ipport"` //源socket的port
	UpdateTime time.Time `json:"updateTime" db:"updatetime"`
}

//网关插卡信息
type IotModuleInfo struct {
	GwMac           string `json:"GwMac"`
	IotModuleId     uint16 `json:"iotModuleId"`
	IotModuleStatus uint   `json:"iotModuleStatus"`
}

//终端设备信息
type TerminalInfo struct {
	TerminalName  string    `json:"terminalName,omitempty"`
	TerminalMac   string    `json:"terminalMac,omitempty"`
	GwMac         string    `json:"gwMac,omitempty"`
	IotModuleId   string    `json:"iotModuleId,omitempty"`
	RSSI          int8      `json:"rssi,omitempty"`
	SupportConn   bool      `json:"supportConn,omitempty"`
	ConnectStatus string    `json:"connectStatus,omitempty"`
	TimeStamp     time.Time `json:"timestamp,omitempty"`
	ConnHandle    string    `form:"connHandle" json:"connHandle" binding:"connHandle"`
}

//设备连接
//连接成功以后存储下来
type DevConnection struct {
	DevMac            string `form:"devMac" json:"devMac" binding:"devMac"`
	GwMac             string `form:"gwMac" json:"gwMac" binding:"gwMac"`
	ModuleId          string `form:"moduleId" json:"moduleId" binding:"moduleId"`
	ConnStatus        string `form:"connStatus" json:"connStatus" binding:"connStatus"`
	ConnHandle        string `form:"connHandle" json:"connHandle" binding:"connHandle"`
	PHY               string `form:"phy" json:"phy" binding:"phy"`
	MTU               string `form:"mtu" json:"mtu" binding:"mtu"`
	AddrType          uint8  `form:"addrType" json:"addrType" binding:"addrType"`
	OpType            uint8  `form:"opType" json:"opType" binding:"opType"`
	ScanType          uint8  `form:"scanType" json:"scanType" binding:"scanType"`
	ScanInterval      uint16 `form:"scanInterval" json:"scanInterval" binding:"scanInterval"`
	ScanWindow        uint16 `form:"scanWindow" json:"scanWindow" binding:"scanWindow"`
	ScanTimeout       uint16 `form:"scanTimeout" json:"scanTimeout" binding:"scanTimeout"`
	ConnInterval      uint16 `form:"connInterval" json:"connInterval" binding:"connInterval"`
	ConnLatency       uint16 `form:"connLatency" json:"connLatency" binding:"connLatency"`
	SupervisionWindow uint16 `form:"supervisionWindow" json:"supervisionWindow" binding:"supervisionWindow"`
}

//扫描所需参数
type ScanInfo struct {
	EnableScan   uint8  `form:"enableScan" json:"enableScan" binding:"enableScan"`
	ScanType     uint8  `form:"scanType" json:"scanType" binding:"scanType"`
	ScanInterval uint16 `form:"scanInterval" json:"scanInterval" binding:"scanInterval"`
	ScanWindow   uint16 `form:"scanWindow" json:"scanWindow" binding:"scanWindow"`
	ScanTimeout  uint16 `form:"ScanTimeout" json:"ScanTimeout" binding:"ScanTimeout"`
	GwMac        string `form:"gwMac"  json:"gwMac" binding:"gwMac"`
	IotModuleId  uint16 `form:"iotModuleId" json:"iotModuleId" binding:"iotModuleId"`
}

//响应web界面端
type ResultMessage struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

//设备主服务消息
type ServiceInfo struct {
	PrimaryService bool   `json:"primaryService" db:"primaryService"` //是否为主服务
	UUIDService    uint16 `json:"uuidService" db:"uuidService"`
	HandleService  uint16 `json:"handleService" db:"handleService"`
	DevMac         string `json:"devMac" db:"devMac"`
}

//handle请求
type CharacterInfo struct {
	DevMac      string
	StartHandle string
	EndHandle   string
}

//RPCIotware
type RPCIotware struct {
	Device string `json:"device"`
	Data   string `json:"data"`
}

//TelemetryIotware
type TelemetryIotware struct {
	Device string `json:"device"`
	Data   string `json:"data"`
}

//ack
type NetworkInAckIotware struct {
	Device string `json:"device"`
	OK     bool   `json:"isok"`
}

//networkIn
type NetworkInIotware struct {
	Device string `json:"device"`
}

//存服务特征的最终节点
type ServiceCharacterNode struct {
	CharUUID        string `json:"charUUID"`
	CharacterHandle string `json:"characterHandle"` //
	CharacterValue  string `json:"characterValue"`  //标准string存储方式
	CCCDHanle       string `json:"cccdHandle"`      //
	CCCDHanleValue  string `json:"cccdHandleValue"`
	Properties      string `json:"properties"` //特征值属性位域
}
