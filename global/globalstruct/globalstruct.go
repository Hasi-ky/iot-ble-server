package globalstruct

import "time"

type SocketInfo struct {
	ID         uint      `json:"id" bson:"-" db:"id"`
	Mac        string    `json:"GwMac" bson:"ACMac,omitempty" db:"gwmac"`
	Family     string    `json:"family" bson:"family,omitempty" db:"family"` //源socket的IP协议
	IPAddr     string    `json:"IPAddr" bson:"IPAddr,omitempty" db:"ipaddr"` //源socket的IP
	IPPort     int       `json:"IPPort" bson:"IPPort,omitempty" db:"ipport"` //源socket的port
	UpdateTime time.Time `json:"updateTime" bson:"updateTime,omitempty" db:"updatetime"`
}