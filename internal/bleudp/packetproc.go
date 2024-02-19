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
	"sort"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//`解码上行数据使用`
//hello ack消息处理
func procHelloAck(ctx context.Context, jsoninfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procHelloAck> : devEui: %s start proc hello msg\n", devEui)
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
func procIotModuleStatus(ctx context.Context, jsoninfo packets.JsonUdpInfo, devEui string) {
	var err error
	globallogger.Log.Infof("<procIotModuleStatus> : devEui: %s start proc iotmodule status msg\n", devEui)
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
				globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s redis has error %v\n", devEui, err)
				return
			}
		}
		err = globalredis.RedisCache.Set(ctx, cacheKey, byteIotModuleInfo, globalconstants.TTLDuration).Err()
	} else {
		_, err = globalmemo.BleFreeCache.Get([]byte(cacheKey))
		if err != nil {
			globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s memo get value has error %v\n", devEui, err)
			return
		}
		err = globalmemo.BleFreeCache.Set([]byte(cacheKey), byteIotModuleInfo, int(globalconstants.TTLDuration.Seconds()))
	}
	if err != nil {
		globallogger.Log.Errorf("<procIotModuleStatus>: devEui: %s set cache has error %v\n", devEui, err)
	} else {
		globallogger.Log.Infof("<procIotModuleStatus> : devEui: %s update iotmodule status success\n", devEui)
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
	globallogger.Log.Infof("<procBleBoardCast> DevEuiL %s start deal ble boardcast message", devEui)
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
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't compress data %v\n", devEui, err)
				return
			}
		} else {
			err = json.Unmarshal(devCacheByte, &devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't resolve %v\n", devEui, err)
				return
			}
			devInfo.TimeStamp = time.Now()
			devCacheByte, err = json.Marshal(devInfo)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> DevEui %s the memo msg can't resolve data %v\n", devEui, err)
				return
			}
		}
		globalmemo.BleFreeCacheDevInfo.Set([]byte(jsoninfo.MessageAppBody.TLV.TLVPayload.DevMac), devCacheByte, 0) //刷新和写入
	}
}

//处理应答信息
func procBleResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleBoardCast> DevEui %s start deal ble response message\n", devEui)
	switch jsonInfo.MessageAppHeader.Type {
	case packets.TLVScanRespMsg: //此时，需要将消息进行确认掉
		procScanResponse(ctx, jsonInfo, devEui)
	case packets.TLVConnectRespMsg:
		procBleConnResponse(ctx, jsonInfo, devEui)
	case packets.TLVMainServiceRespMsg:
		procBleMainServiceResponse(ctx, jsonInfo, devEui)
	case packets.TLVCharRespMsg:
		procBleCharFindResponse(ctx, jsonInfo, devEui)
	default:
		globallogger.Log.Errorf("<procBleResponse>: Ble DevEui:%s received unrecognized  message type%s:\n", devEui, jsonInfo.MessageAppHeader.Type)
	}
}

//处理连接应答
//处理应答需要pop数据
//MTU交换、PHY协商操作,暂时省略
func procBleConnResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleConnResponse> DevEui %s start deal ble connect response message\n", devEui)
	if jsonInfo.MessageAppBody.ErrorCode != packets.Success {
		globallogger.Log.Errorf("<procBleConnResponse> DevEui %s has an error %v\n", devEui, packets.GetResult(jsonInfo.MessageAppBody.ErrorCode, packets.English).String())
		globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_ERROR, Message: packets.GetResult(jsonInfo.MessageAppBody.ErrorCode, packets.Chinese).String()}
		return
	}
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	if CheckUpMessageFrame(ctx, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, int(curSN)) != globalconstants.TERMINAL_CORRECT {
		globallogger.Log.Errorf("<procBleConnResponse> DevEui %s frame ")
		return //消息响应校对，若校准正确则继续执行，否则抛出
	}
	var (
		terminal globalstruct.TerminalInfo
		strInfo  string
		byteInfo []byte
		err      error
	)
	if config.C.General.UseRedis {
		strInfo, err = globalredis.RedisCache.HGet(ctx, globalconstants.BleDevInfoCachePrefix, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac).Result()
		if err != nil {
			if err != redis.Nil {
				globallogger.Log.Errorf("<procBleConnResponse> DevEui %s has redis error %v\n", devEui, err)
				globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_ERROR, Message: globalconstants.ERROR_CACHE_EXCEPTION}
			} else {
				globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_ERROR, Message: globalconstants.ERROR_CACHE_ABSENCE}
			}
			return
		}
		byteInfo = []byte(strInfo)
	} else {
		byteInfo, err = globalmemo.BleFreeCacheDevInfo.Get([]byte(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac))
		if err != nil {
			globallogger.Log.Errorf("<procBleConnResponse> DevEui %s has cache error %v\n", devEui, err)
			globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_ERROR, Message: globalconstants.ERROR_CACHE_ABSENCE}
			return
		}
	}
	err = json.Unmarshal(byteInfo, &terminal)
	if err != nil {
		globallogger.Log.Errorf("<procBleConnResponse> DevEui %s has data change error %v\n", devEui, err)
		globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_ERROR, Message: globalconstants.ERROR_DATA_CHANGE}
		return
	}
	connStatu, err1 := strconv.ParseInt(jsonInfo.MessageAppBody.TLV.TLVPayload.ConnStatus, 16, 8)
	if err1 != nil {
		globallogger.Log.Errorf("<procBleConnResponse> DevEui %s conn data change error %v\n", devEui, err1)
		return
	}
	terminal.ConnectStatus = uint8(connStatu)
	cacheInfo, err2 := json.Marshal(terminal)
	if err2 != nil {
		globallogger.Log.Errorf("<procBleConnResponse> DevEui %s data convert json error %v\n", devEui, err2)
		return
	}
	if config.C.General.UseRedis { //此处不处理错误
		globalredis.RedisCache.HSet(ctx, globalconstants.BleDevInfoCachePrefix, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, cacheInfo)
	} else {
		globalmemo.BleFreeCacheDevInfo.Set([]byte(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac), cacheInfo, 0)
	}
	globalconstants.ConnectionInfoChan <- globalstruct.ResultMessage{Code: globalconstants.HTTP_CODE_SUCCESS, Message: globalconstants.HTTP_MESSAGE_SUCESS}
}

//处理扫描应答
func procScanResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procScanResponse> DevEui %s start deal ble scan response message\n", devEui)
	if jsonInfo.MessageAppBody.ErrorCode != packets.Success {
		globallogger.Log.Errorf("<procScanResponse> DevEui %s has an error %v\n", devEui, packets.GetResult(jsonInfo.MessageAppBody.ErrorCode, packets.English).String())
		return
	}
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	verifyFrame := CheckUpMessageFrame(ctx, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, int(curSN))
	if verifyFrame != globalconstants.TERMINAL_CORRECT { //目前仅后面
		return //消息响应校对，若校准正确则继续执行，否则抛出
	}
	curTime := time.Now()
	passTime, ok := globalmemo.MemoCacheScanTimeOut.Get(jsonInfo.MessageBody.GwMac + jsonInfo.MessageBody.ModuleID)
	tempTime, _ := strconv.ParseInt(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanTimeout, 16, 32)
	limitTime := time.Duration(tempTime) * time.Millisecond
	if !ok || globalutils.CompareTimeIsExpire(curTime, passTime.(time.Time), limitTime) {
		globallogger.Log.Errorf("<procScanResponse> DevEui %s scan timeout, unable to perform subsequent connections\n", devEui)
		return
	}
}

//处理主服务报文应答, 这个deveui应当描述的是设备mac
//主服务先存缓存，然后再存数据库， 查看缓存中的信息是否发生更改，更改则清空数据库相关内容，重新写入
//自身缓存则多加一级: 是否为主服务放置在最后
func procBleMainServiceResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleMainServiceResponse> DevEui %s start deal ble mainservice response message\n", devEui)
	var err error
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleMainServiceResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) {
		tempServiceHandle := make([]string, 0)
		if config.C.General.UseRedis {
			pipe := globalredis.RedisCache.Pipeline()
			for _, serviceTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
				tempServiceHandle = append(tempServiceHandle, serviceTLV.TLVPayload.ServiceHandle)
				pipe.HSet(ctx, devEui+globalconstants.DEV_SERVICE, serviceTLV.TLVPayload.ServiceHandle, serviceTLV.TLVPayload.Primary)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s redis has error %v \n", devEui, err)
				return
			}
		} else {
			serviceMap := make(map[string]string)
			for _, serviceTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
				serviceMap[serviceTLV.TLVPayload.ServiceHandle] = serviceTLV.TLVPayload.Primary
				tempServiceHandle = append(tempServiceHandle, serviceTLV.TLVPayload.ServiceHandle)
				globalmemo.MemoCacheService.Set(devEui+globalconstants.DEV_SERVICE, serviceMap)
			}
		}
		if len(tempServiceHandle) != 0 {
			sort.Strings(tempServiceHandle) //防止应答乱序
			err = ResumeCharacterFind(ctx, devEui, tempServiceHandle[0], tempServiceHandle[len(tempServiceHandle)-1])
			if err != nil {
				globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s has error %v \n", devEui, err)
			}
		} else {
			globallogger.Log.Warnln("<procBleMainServiceResponse> not have service handle")
		}
	}
}

//处理特征值响应
//devEui通常为设备mac
//存储缓存为设备服务唯一标识： 特征handle : 对应值(默认)
//特征值映射完成
func procBleCharFindResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleCharFindResponse> DevEui %s start deal ble characterValue response message\n", devEui)
	var err error
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleCharFindResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) {
		if config.C.General.UseRedis {
			pipe := globalredis.RedisCache.Pipeline()
			for _, characterTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
				cacheKey := globalutils.CreateCacheKey(devEui, globalconstants.DEV_CHARACTER, characterTLV.TLVPayload.ServiceHandle)
				pipe.HSet(ctx, cacheKey, characterTLV.TLVPayload.CharHandle, globalconstants.DEV_CHARACTER_DEFAULT)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s redis has error %v \n", devEui, err)
				return
			}
		} else {
			if len(jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve) != 0 {
				characterMap := make(map[string]int)
				cacheKey := globalutils.CreateCacheKey(devEui, globalconstants.DEV_CHARACTER,  jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve[0].TLVPayload.ServiceHandle)
				for _, characterTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
					characterMap[characterTLV.TLVPayload.CharHandle] = globalconstants.DEV_CHARACTER_DEFAULT
				}
				globalmemo.MemoCacheServiceForChar.Set(cacheKey, characterMap)
			}
		}
	}
}
