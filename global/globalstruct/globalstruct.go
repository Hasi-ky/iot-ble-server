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
	IotModuleId   uint16    `json:"iotModuleId,omitempty"`
	RSSI          int8      `json:"rssi,omitempty"`
	SupportConn   bool      `json:"supportConn,omitempty"`
	ConnectStatus uint8     `json:"connectStatus,omitempty"`
	TimeStamp     time.Time `json:"timestamp,omitempty"`
}

 