package bleudp

import (
	"context"
	"encoding/json"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalsocket"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"iot-ble-server/internal/storage"
	"time"

	"github.com/redis/go-redis/v9"
)

//
func procHelloAck(ctx context.Context, jsoninfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procHelloAck> : devEui: %s start proc hello msg", devEui)
	newJsonInfo := packets.JsonUdpInfo{}
	reMsgHeader := packets.MessageHeader{
		Version:           jsoninfo.MessageHeader.Version,
		LinkMessageLength: "10",
		LinkMsgFrameSN:    jsoninfo.MessageHeader.LinkMsgFrameSN,
		LinkMsgType:       jsoninfo.MessageAppHeader.Type,
		OpType:            packets.Response,
	}
	updateSockets(ctx, jsoninfo.MessageBody.GwMac, jsoninfo.Rinfo, config.C.General.KeepAliveTime)
	newJsonInfo.MessageHeader = reMsgHeader
	sendBytes := EnCodeForDownUdpMessage(newJsonInfo, globalconstants.CtrlLinkedMsgHeader)
	SendDownMessage(sendBytes, jsoninfo.MessageBody.GwMac, jsoninfo.MessageBody.GwMac)
}

func updateSockets(ctx context.Context, gwMac string, rinfo dgram.RInfo, msgAliveTime int) {
	cacheKey := globalutils.CreateCacheKey(globalconstants.GwSocketCachePrefix, gwMac)
	socketInfo := &globalstruct.SocketInfo{}
	preCacheData := globalstruct.SocketInfo{
		Mac:        gwMac,
		Family:     rinfo.Family,
		IPAddr:     rinfo.Address,
		IPPort:     rinfo.Port,
		UpdateTime: time.Now(),
	}
	cachedSocketByte, err := globalmemo.BleFreeCache.Get([]byte(cacheKey))
	cacheDataByte, _ := json.Marshal(preCacheData)
	if err == nil {
		json.Unmarshal(cachedSocketByte, socketInfo)
	}
	if err != nil || rinfo.Address != socketInfo.IPAddr || rinfo.Family != socketInfo.Family || rinfo.Port != socketInfo.IPPort {
		globallogger.Log.Warnf("<updateSockets> DevEui: %v, Update Socket information", gwMac)
		globalmemo.BleFreeCache.Set([]byte(cacheKey), cacheDataByte, msgAliveTime+5)
		if config.C.General.UseRedis {
			globalredis.RedisCache.Set(ctx, cacheKey, cacheDataByte, -1)
		}
		pgSocketSet := map[string]interface{}{
			"gwmac":      gwMac,
			"family":     rinfo.Family,
			"ipaddr":     rinfo.Address,
			"ipport":     rinfo.Port,
			"updatetime": preCacheData.UpdateTime,
		}
		err = storage.FindSocketAndUpdatePG(pgSocketSet)
		if err != nil {
			globallogger.Log.Errorln("<updateSockets> : Failed to save to database")
		}
	}
}

// `0` deveui | `1` mac | `2` module id |
func SendDownMessage(data []byte, devEui, mac string) {
	var (
		socketInfo          *globalstruct.SocketInfo
		err, err1, err2     error
		socketInfoBytes     []byte
		cacheKey, socketVal string
	)
	cacheKey = globalutils.CreateCacheKey(globalconstants.GwSocketCachePrefix, mac)
	socketInfoBytes, err = globalmemo.BleFreeCache.Get([]byte(cacheKey))
	if err != nil {
		if config.C.General.UseRedis {
			socketVal, err1 = globalredis.RedisCache.Get(context.TODO(), cacheKey).Result()
			if err1 != redis.Nil {
				socketInfo, err2 = storage.FindSocketByGwMac(mac)
				if err2 != nil {
					globallogger.Log.Errorf("<SendDownMessage> DevEui: %s send fail, can't find right socket to send msg, reseaon:%v", devEui, err2)
					return
				}
			} else {
				json.Unmarshal([]byte(socketVal), &socketInfo)
			}
		}
	} else {
		json.Unmarshal(socketInfoBytes, &socketInfo)
	}
	if socketInfo == nil {
		globallogger.Log.Errorln("<SendDownMessage> DevEui: %s missing corresponding socket information", devEui)
	} else if socketInfo.Family == "IPv6" {
		globallogger.Log.Errorln("<SendDownMessage> DevEui: %s current not support IPv6", devEui)
	} else {
		err = globalsocket.ServiceSocket.Send(data, socketInfo.IPPort, socketInfo.IPAddr)
		if err != nil {
			globallogger.Log.Errorf("<SendDownMessage> DevEui: %s send message occur error %v", devEui, err)
			return
		}
		globallogger.Log.Infof("<SendDownMessage> DevEui: %s send message success", devEui)
	}
}
