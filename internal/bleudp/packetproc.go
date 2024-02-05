package bleudp

import (
	"context"
	"encoding/json"
	"errors"
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
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//`解码上行数据使用`
//hello ack消息处理
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
	newJsonInfo.PendCtrl = globalconstants.CtrlLinkedMsgHeader
	sendBytes := EnCodeForDownUdpMessage(newJsonInfo)
	SendDownMessage(sendBytes, jsoninfo.MessageBody.GwMac, jsoninfo.MessageBody.GwMac)
}

//此事devEUI可以表示为网关插卡+物联网模块
//压入redis当中是字节数组，走json编码格式
func procIotModuleStatus(ctx context.Context, jsoninfo packets.JsonUdpInfo, devEui string) {
	var err error
	globallogger.Log.Infof("<procIotModuleStatus> : devEui: %s start proc iotmodule status msg", devEui)
	cacheKey := globalutils.CreateCacheKey(globalconstants.GwIotModuleCachePrefie, jsoninfo.MessageBody.GwMac, jsoninfo.MessageBody.ModuleID)
	moduleId, _ := strconv.ParseUint(jsoninfo.MessageBody.ModuleID, 16, 16)
	moduleStatus, _ := strconv.ParseUint(jsoninfo.MessageBody.TLV.TLVPayload.IotModuleStatus, 16, 16)
	tempVal := globalstruct.IotModuleInfo{
		GwMac:           jsoninfo.MessageBody.GwMac,
		IotModuleId:     uint16(moduleId),
		IotModuleStatus: uint(moduleStatus),
	}
	byteIotModuleInfo, _ := json.Marshal(tempVal)
	if config.C.General.UseRedis {
		_, err = globalredis.RedisCache.Get(ctx, cacheKey).Result()
		if err != nil {
			if err != redis.Nil {
				globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s redis has error %v", devEui, err)
				return
			}
		}
		err = globalredis.RedisCache.Set(ctx, cacheKey, byteIotModuleInfo, globalconstants.TTLDuration).Err()
	} else {
		_, err = globalmemo.BleFreeCache.Get([]byte(cacheKey))
		if err != nil {
			globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s memo get value has error %v", devEui, err)
			return
		}
		err = globalmemo.BleFreeCache.Set([]byte(cacheKey), byteIotModuleInfo, int(globalconstants.TTLDuration.Seconds()))
	}
	if err != nil {
		globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s set cache has error %v", devEui, err)
	} else {
		globallogger.Log.Infof("<procIotModuleStatus> : devEui: %s update iotmodule status success", devEui)
	}
}

//更新网络信息
func updateSockets(ctx context.Context, gwMac string, rinfo dgram.RInfo, msgAliveTime int) error {
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
	if err != nil {
		return errors.New("<updateSockets> gateway can't resolve message" + err.Error())
	}
	json.Unmarshal(cachedSocketByte, socketInfo)
	if rinfo.Address != socketInfo.IPAddr || rinfo.Family != socketInfo.Family || rinfo.Port != socketInfo.IPPort {
		globallogger.Log.Warnf("<updateSockets> DevEui: %v, Update Socket information", gwMac)
		if config.C.General.UseRedis {
			globalredis.RedisCache.Set(ctx, cacheKey, cacheDataByte, -1)
		} else {
			globalmemo.BleFreeCache.Set([]byte(cacheKey), cacheDataByte, msgAliveTime+5)
		}
		pgSocketSet := map[string]interface{}{
			"gwmac":      gwMac,
			"family":     preCacheData.Family,
			"ipaddr":     preCacheData.IPAddr,
			"ipport":     preCacheData.IPPort,
			"updatetime": preCacheData.UpdateTime,
		}
		err = storage.FindSocketAndUpdatePG(pgSocketSet)
		if err != nil {
			return errors.New("<updateSockets> : Failed to save to database")
		}
	}
	return nil
}

// 消息下行接口
//这里传入的mac应该是网关对应的mac
func SendDownMessage(data []byte, devEui, mac string) {
	var (
		socketInfo          *globalstruct.SocketInfo
		err                 error
		socketInfoBytes     []byte
		cacheKey, socketVal string
	)
	cacheKey = globalutils.CreateCacheKey(globalconstants.GwSocketCachePrefix, mac)
	if config.C.General.UseRedis {
		socketVal, err = globalredis.RedisCache.Get(context.TODO(), cacheKey).Result()
		socketInfoBytes = []byte(socketVal)
	} else {
		socketInfoBytes, err = globalmemo.BleFreeCache.Get([]byte(cacheKey))
	}
	if err != nil && err != redis.Nil {
		globallogger.Log.Errorf("<SendDownMessage> DevEui: %s send fail :%v\n", devEui, err)
		return
	}
	if len(socketInfoBytes) == 0 {
		socketInfo, err = storage.FindSocketByGwMac(mac)
		if err != nil {
			globallogger.Log.Errorf("<SendDownMessage> DevEui: %s send fail, can't find right socket to send msg, reseaon:%v\n", devEui, err)
			return
		}
	} else {
		json.Unmarshal(socketInfoBytes, &socketInfo)
	}
	if socketInfo == nil {
		globallogger.Log.Errorf("<SendDownMessage> DevEui: %s missing corresponding socket information\n", devEui)
	} else if socketInfo.Family == "IPv6" {
		globallogger.Log.Errorf("<SendDownMessage> DevEui: %s current not support IPv6\n", devEui)
	} else {
		err = globalsocket.ServiceSocket.Send(data, socketInfo.IPPort, socketInfo.IPAddr)
		if err != nil {
			globallogger.Log.Errorf("<SendDownMessage> DevEui: %s send message occur error %v\n", devEui, err)
			return
		}
		globallogger.Log.Infof("<SendDownMessage> DevEui: %s send message success\n", devEui)
	}
}

//终端广播报文能处理, 唯一映射即可
//连接|断联以后才会有连接状态 首次情况下肯定没有
//`收到连接报文的时候还需要重新设置设备相关连接状态`
func procBleBoardCast(ctx context.Context, jsoninfo packets.JsonUdpInfo, devEui string) {
	var (
		devCacheByte []byte
		devCacheStr  string
		devInfo      globalstruct.TerminalInfo
		err          error
	)
	moduleId, _ := strconv.ParseUint(jsoninfo.MessageBody.ModuleID, 16, 16)
	rssi, _ := strconv.ParseInt(jsoninfo.MessageBody.TLV.TLVPayload.RSSI, 16, 8)
	if config.C.General.UseRedis {
		devCacheStr, err = globalredis.RedisCache.HGet(ctx, globalconstants.BleDevInfoCachePrefix, jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac).Result()
		judgeNum := globalutils.JudgeGet(err)
		if judgeNum == globalconstants.JudgeGetError {
			globallogger.Log.Errorf("<procBleBoardCast> DevEui: %s redis has error %v\n", devEui, err)
			return
		} else if judgeNum == globalconstants.JudgeGetNil { //缺少数据压入
			devInfo = globalstruct.TerminalInfo{ //支持连接应当在content内容中附带，默认不支持
				TerminalName: jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac,
				TerminalMac:  jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac,
				GwMac:        jsoninfo.MessageBody.GwMac,
				IotModuleId:  uint16(moduleId),
				RSSI:         int8(rssi),
				TimeStamp:    time.Now(),
			}
			devCacheByte, err = json.Marshal(devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the redis msg can't compress data %v\n", devEui, err)
				return
			}
		} else { //更新
			err = json.Unmarshal([]byte(devCacheStr), &devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the redis msg can't resolve %v\n", devEui, err)
				return
			}
			devInfo.TimeStamp = time.Now()
			devCacheByte, err = json.Marshal(devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the redis msg can't resolve data %v", devEui, err)
				return
			}
		}
		globalredis.RedisCache.HSet(ctx, globalconstants.BleDevInfoCachePrefix, jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac, devCacheByte) //刷新或重新键入
	} else {
		devCacheByte, err = globalmemo.BleFreeCacheDevInfo.Get([]byte(jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac))
		if err != nil {
			devInfo = globalstruct.TerminalInfo{ //支持连接应当在content内容中附带，默认不支持
				TerminalName: jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac,
				TerminalMac:  jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac,
				GwMac:        jsoninfo.MessageBody.GwMac,
				IotModuleId:  uint16(moduleId),
				RSSI:         int8(rssi),
				TimeStamp:    time.Now(),
			}
			devCacheByte, err = json.Marshal(devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't compress data %v", devEui, err)
				return
			}
		} else {
			err = json.Unmarshal(devCacheByte, &devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't resolve %v", devEui, err)
				return
			}
			devInfo.TimeStamp = time.Now()
			devCacheByte, err = json.Marshal(devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't resolve data %v", devEui, err)
				return
			}
		}
		globalmemo.BleFreeCacheDevInfo.Set([]byte(jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac), devCacheByte, 0) //刷新和写入
	}
}
