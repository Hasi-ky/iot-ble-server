package globalstruct

import "time"

type SocketInfo struct {
	ID         uint      `json:"id" bson:"-" gorm:"primary_key"`
	Mac        string    `json:"ACMac" bson:"ACMac,omitempty" gorm:"column:acmac"`
	APMac      string    `json:"APMac" bson:"APMac,omitempty" gorm:"column:apmac"`
	Family     string    `json:"family" bson:"family,omitempty" gorm:"column:family"` //源socket的IP协议
	IPAddr     string    `json:"IPAddr" bson:"IPAddr,omitempty" gorm:"column:ipaddr"` //源socket的IP
	IPPort     int       `json:"IPPort" bson:"IPPort,omitempty" gorm:"column:ipport"` //源socket的port
	UpdateTime time.Time `json:"updateTime" bson:"updateTime,omitempty" gorm:"column:updatetime"`
	CreateTime time.Time `json:"createTime" bson:"createTime,omitempty" gorm:"column:createtime"`
}