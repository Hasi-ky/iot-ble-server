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
	"strings"
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
		LinkMessageLength: "0010",
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
//暂时不使用redis进行记录
func procBleBoardCast(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleBoardCast> DevEuiL %s start deal ble boardcast message", devEui)
	var curNode globalstruct.TerminalInfo
	rssi, _ := strconv.ParseInt(jsonInfo.MessageBody.TLV.TLVPayload.RSSI, 16, 8)
	devMac := jsonInfo.MessageBody.TLV.TLVPayload.DevMac
	secondCacheKey := globalutils.CreateCacheKey(jsonInfo.MessageBody.GwMac, jsonInfo.MessageBody.ModuleID)
	for _, curTLV := range jsonInfo.MessageAppBody.MultiTLV {
		curNode = globalstruct.TerminalInfo{ //支持连接应当在content内容中附带，默认不支持
			TerminalName: curTLV.TLVPayload.DevMac,
			TerminalMac:  curTLV.TLVPayload.DevMac,
			GwMac:        jsonInfo.MessageBody.GwMac,
			IotModuleId:  jsonInfo.MessageBody.ModuleID,
			RSSI:         int8(rssi),
			TimeStamp:    time.Now(),
		}
		gwInfo, ok := globalmemo.MemoCacheDownPriority.Get(devMac)
		if !ok {
			tempGwMap := make(map[string]globalstruct.TerminalInfo)
			tempGwMap[secondCacheKey] = curNode
		} else {
			var (
				pastNode globalstruct.TerminalInfo
				exist    bool
			)
			if pastNode, exist = gwInfo.(map[string]globalstruct.TerminalInfo)[secondCacheKey]; !exist {
				gwInfo.(map[string]globalstruct.TerminalInfo)[secondCacheKey] = curNode
				globalmemo.MemoCacheDownPriority.Set(devMac, gwInfo)
			} else {
				if pastNode.RSSI < curNode.RSSI {
					gwInfo.(map[string]globalstruct.TerminalInfo)[secondCacheKey] = curNode
					globalmemo.MemoCacheDownPriority.Set(devMac, gwInfo)
				}
			}
		}
		//分析解析后的内容0102则开启连接
		if curTLV.TLVPayload.NoticeContent.Data.AdData.CompData.MsgType == "0102" {
			gwInfo, _ := globalmemo.MemoCacheDownPriority.Get(devMac)
			priorityNode := globalstruct.TerminalInfo{
				RSSI: -128,
			}
			for _, node := range gwInfo.(map[string]globalstruct.TerminalInfo) {
				if node.RSSI > priorityNode.RSSI {
					priorityNode = node
				}
			}
			err := ConnectBleDev(ctx, priorityNode)
			if err != nil {
				globallogger.Log.Errorf("<procBleBoardCast> deal dev connect fail , error is %v\n", err)
			}
		}
	}
}

func procBleConfirm(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleConfirm> DevEui %s start deal ble confirm message, the app FrameSN: %v\n", devEui, jsonInfo.MessageAppHeader.SN)
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
	case packets.TLVMainServiceByUUIDRespMsg:
		procBleUUIDServiceResponse(ctx, jsonInfo, devEui)
	case packets.TLVCharConfRespMsg:
		procBleCharacterConfResponse(ctx, jsonInfo, devEui)
	default:
		globallogger.Log.Errorf("<procBleResponse>: Ble DevEui:%s received unrecognized  message type%s:\n", devEui, jsonInfo.MessageAppHeader.Type)
	}
}

//处理连接应答
//处理应答需要pop数据
// 连接成功前检测该消息是否超时

func procBleConnResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleConnResponse> DevEui %s start deal ble connect response message\n", devEui)
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleConnResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) {
		connDevInfo := globalstruct.TerminalInfo{
			TerminalMac:   jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac,
			GwMac:         jsonInfo.MessageBody.GwMac,
			IotModuleId:   jsonInfo.MessageBody.ModuleID,
			ConnectStatus: jsonInfo.MessageAppBody.TLV.TLVPayload.ConnStatus,
			ConnHandle:    jsonInfo.MessageAppBody.TLV.TLVPayload.ConnHandle,
		}
		connCacheKey := globalutils.CreateCacheKey(globalconstants.BleDevCacheConnPrefix, connDevInfo.ConnHandle)
		InfoByte, err := json.Marshal(connDevInfo)
		if err != nil {
			globallogger.Log.Errorf("<procBleConnResponse> DevEui %s change data has error %v\n", err)
			return
		}
		if config.C.General.UseRedis {
			globalredis.RedisCache.Set(ctx, connCacheKey, InfoByte, 0)
		} else {
			globalmemo.BleFreeCacheDevConnInfo.Set([]byte(connCacheKey), InfoByte, 0)
		}
		//推送连接结果？
		/* mqtt 消息推送*/

		//开始主服务发现
		err = SearchMainService(ctx, connDevInfo)
		if err != nil {
			globallogger.Log.Errorf("<procBleConnResponse> has error %v\n", err)
		}
		//
	}
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
//这一步之前需要将所有的信息处理完成后才进行后续操作
func procBleMainServiceResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleMainServiceResponse> DevEui %s start deal ble mainservice response message\n", devEui)
	var err error
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleMainServiceResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) {
		serviceUUIDWithHandle := make(map[string]string, 0)
		if config.C.General.UseRedis {
			pipe := globalredis.RedisCache.Pipeline()
			for _, serviceTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
				serviceUUIDWithHandle[serviceTLV.TLVPayload.ServiceUUID] = serviceTLV.TLVPayload.ServiceHandle
				pipe.HSet(ctx, globalutils.CreateCacheKey(globalconstants.DEV_SERVICE, devEui), serviceTLV.TLVPayload.ServiceHandle, serviceTLV.TLVPayload.Primary)
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
				serviceUUIDWithHandle[serviceTLV.TLVPayload.ServiceUUID] = serviceTLV.TLVPayload.ServiceHandle
				globalmemo.MemoCacheService.Set(devEui+globalconstants.DEV_SERVICE, serviceMap)
			}
		}
		if len(serviceUUIDWithHandle) != 0 {
			//sort.Strings(tempServiceHandle) //防止应答乱序
			//err = ResumeCharacterFind(ctx, devEui, tempServiceHandle[0], tempServiceHandle[len(tempServiceHandle)-1])
			var terminalInfo globalstruct.TerminalInfo
			cacheKey := globalutils.CreateCacheKey(globalconstants.BleDevCacheConnPrefix, devEui)
			if config.C.General.UseRedis {
				strInfo, err1 := globalredis.RedisCache.Get(ctx, cacheKey).Result()
				if err1 != nil {
					globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s loss connect dev infomation\n", devEui)
					return
				}
				err1 = json.Unmarshal([]byte(strInfo), &terminalInfo)
				if err1 != nil {
					globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s connect redis infomation has error %v\n", devEui, err1)
					return
				}
			} else {
				byteInfo, err1 := globalmemo.BleFreeCacheDevConnInfo.Get([]byte(cacheKey))
				if err1 != nil {
					globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s loss connect dev infomation\n", devEui)
					return
				}
				err1 = json.Unmarshal(byteInfo, &terminalInfo)
				if err1 != nil {
					globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s connect memo infomation has error %v\n", devEui, err1)
					return
				}
			}
			for uuid, _ := range serviceUUIDWithHandle {
				if uuid == "0001" {
					err1 := SearchMainServiceUUID(ctx, terminalInfo, uuid)
					if err1 != nil {
						globallogger.Log.Errorf("<procBleMainServiceResponse> DevEui %s search main service error %v\n", err1)
					}
					break //可以及时退出
				}
			}
		} else {
			globallogger.Log.Warnln("<procBleMainServiceResponse> not have right service")
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
				cacheKey := globalutils.CreateCacheKey(devEui, globalconstants.DEV_CHARACTER, jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve[0].TLVPayload.ServiceHandle)
				for _, characterTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
					characterMap[characterTLV.TLVPayload.CharHandle] = globalconstants.DEV_CHARACTER_DEFAULT
				}
				globalmemo.MemoCacheServiceForChar.Set(cacheKey, characterMap)
			}
		}
	}
}

//针对特定UUID的缓存生成
//服务key(devmac + service)
//默认一个uuid服务
func procBleUUIDServiceResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleUUIDServiceResponse> DevEui %s start deal ble uuid service response message\n", devEui)
	var (
		err           error
		reserveLen    = len(jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve)
		nusServiceKey string
	)
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleUUIDServiceResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) &&
		reserveLen > 0 {
		allMessageTLV := make(map[string]map[string]globalstruct.ServiceCharacterNode)
		uuidForhandleSvc := make(map[string]string) //handle对应uuid
		uuidForhandleChar := make(map[string]string)
		serviceTLV, characterTLV, characterDesp := make([]packets.TLV, 0), make([]packets.TLV, 0), make([]packets.TLV, 0)
		for _, curTLV := range jsonInfo.MessageAppBody.TLV.TLVPayload.TLVReserve {
			if curTLV.TLVMsgType == packets.TLVServiceMsg {
				serviceTLV = append(serviceTLV, curTLV)
				uuidForhandleSvc[curTLV.TLVPayload.ServiceHandle] = curTLV.TLVPayload.ServiceUUID
			} else if curTLV.TLVMsgType == packets.TLVCharacteristicMsg {
				characterTLV = append(characterTLV, curTLV)
				uuidForhandleChar[curTLV.TLVPayload.CharHandle] = curTLV.TLVPayload.CharacterUUID
			} else {
				characterDesp = append(characterDesp, curTLV)
			}
		}
		for _, service := range serviceTLV {
			cacheKey := globalutils.CreateCacheKey(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, service.TLVPayload.ServiceUUID, service.TLVPayload.ServiceHandle)
			allMessageTLV[cacheKey] = make(map[string]globalstruct.ServiceCharacterNode)
		}
		for _, character := range characterTLV {
			serviceKey := globalutils.CreateCacheKey(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, uuidForhandleSvc[character.TLVPayload.ServiceHandle], character.TLVPayload.ServiceHandle)
			nusServiceKey = serviceKey //填充
			cacheKey := globalutils.CreateCacheKey(character.TLVPayload.CharacterUUID, character.TLVPayload.CharHandle)
			allMessageTLV[serviceKey][cacheKey] = globalstruct.ServiceCharacterNode{
				CharUUID:        character.TLVPayload.CharacterUUID,
				CharacterHandle: character.TLVPayload.CharHandle,
				Properties:      character.TLVPayload.Properties,
			}
		}
		for _, charDesp := range characterDesp {
			serviceKey := globalutils.CreateCacheKey(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac, uuidForhandleSvc[charDesp.TLVPayload.ServiceHandle], charDesp.TLVPayload.ServiceHandle)
			characterKey := globalutils.CreateCacheKey(charDesp.TLVPayload.CharacterUUID, charDesp.TLVPayload.CharHandle)
			node := allMessageTLV[serviceKey][characterKey]
			node.CCCDHanle = charDesp.TLVPayload.DescriptorHandle
			allMessageTLV[serviceKey][characterKey] = node
		}
		if config.C.General.UseRedis { //覆盖式存储
			pipe := globalredis.RedisCache.Pipeline()
			for svcKey, charMap := range allMessageTLV {
				for charKey, Node := range charMap {
					nodeByte, err1 := json.Marshal(Node)
					if err1 != nil {
						globallogger.Log.Errorf("<procBleUUIDServiceResponse> restore msg has error %v\n", err1)
						continue
					}
					pipe.HSet(ctx, svcKey, charKey, nodeByte)
				}
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> redis has error %v\n", err)
				return
			}
		} else {
			for svcKey, charMap := range allMessageTLV {
				globalmemo.MemoCacheServiceForChar.Set(svcKey, charMap)
			}
		}
		//发起特征值配置请求
		var terminalInfo globalstruct.TerminalInfo
		cacheKey := globalutils.CreateCacheKey(globalconstants.BleDevCacheConnPrefix, devEui)
		if config.C.General.UseRedis {
			strInfo, err1 := globalredis.RedisCache.Get(ctx, cacheKey).Result()
			if err1 != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> DevEui %s loss connect dev infomation\n", devEui)
				return
			}
			err1 = json.Unmarshal([]byte(strInfo), &terminalInfo)
			if err1 != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> DevEui %s connect redis infomation has error %v\n", devEui, err1)
				return
			}
		} else {
			byteInfo, err1 := globalmemo.BleFreeCacheDevConnInfo.Get([]byte(cacheKey))
			if err1 != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> DevEui %s loss connect dev infomation\n", devEui)
				return
			}
			err1 = json.Unmarshal(byteInfo, &terminalInfo)
			if err1 != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> DevEui %s connect memo infomation has error %v\n", devEui, err1)
				return
			}
		}
		//针对0001
		if strings.Contains(nusServiceKey, "0001") {
			err1 := CharacterConfig(ctx, terminalInfo, allMessageTLV[nusServiceKey], nusServiceKey)
			if err1 != nil {
				globallogger.Log.Errorf("<procBleUUIDServiceResponse> DevEui %s character config has error %v\n", devEui, err1)
			}
		}
	}
}

//处理特征配置应答
//将里面的特征值参数配置给改过来 修改某个服务中的特定特征值改变
//是否需要推送消息？
func procBleCharacterConfResponse(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleCharacterConfResponse> DevEui %s start deal ble uuid service response message\n", devEui)
	if DealWithResponseBle(ctx, jsonInfo.MessageHeader.LinkMsgFrameSN, jsonInfo.MessageAppHeader.SN, "procBleCharacterConfResponse", jsonInfo.MessageAppBody.ErrorCode, devEui) {
		serviceKeyByte, err := globalmemo.BleFreeCacheUpDown.Get([]byte(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac + jsonInfo.MessageAppHeader.SN))
		serviceKey := string(serviceKeyByte)
		if err != nil {
			globallogger.Log.Errorf("<procBleCharacterConfResponse> DevEui %s can't find right service information, please check it\n", devEui)
			return
		}
		if config.C.General.UseRedis {
			charMap, err := globalredis.RedisCache.HGetAll(ctx, serviceKey).Result()
			if err != nil {
				globallogger.Log.Errorf("<procBleCharacterConfResponse> DevEui %s can't find right character information (redis), please check it\n", devEui)
				return
			}
			for key, value := range charMap {
				var tempNode globalstruct.ServiceCharacterNode
				err = json.Unmarshal([]byte(value), &tempNode)
				if err != nil {
					globallogger.Log.Errorf("<procBleCharacterConfResponse> DevEui %s can't resolve redis data for character, please check it\n", devEui)
					return
				}
				if tempNode.CharacterHandle == jsonInfo.MessageAppBody.TLV.TLVPayload.CharHandle &&
					tempNode.CCCDHanle == jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandle {
					tempNode.CCCDHanleValue = jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandleValue
					byteNode, _ := json.Marshal(tempNode)
					_, err = globalredis.RedisCache.HSet(ctx, serviceKey, key, byteNode).Result()
					if err != nil {
						globallogger.Log.Errorf("<procBleCharacterConfResponse> DevEui %s redis has error, please check it\n", devEui)
						return
					}
					break
				}
			}
		} else {
			tempMap := make(map[string]globalstruct.ServiceCharacterNode)
			charMap, ok := globalmemo.MemoCacheServiceForChar.Get(serviceKey)
			if !ok {
				globallogger.Log.Errorf("<procBleCharacterConfResponse> DevEui %s can't find right character information (memo), please check it\n", devEui)
				return
			}
			for key, node := range charMap.(map[string]globalstruct.ServiceCharacterNode) {
				if node.CharacterHandle == jsonInfo.MessageAppBody.TLV.TLVPayload.CharHandle &&
					node.CCCDHanle == jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandle {
					node.CCCDHanleValue = jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandleValue
				} else {
					tempMap[key] = node
				}
			}
			globalmemo.MemoCacheServiceForChar.Set(serviceKey, tempMap)
		}
		//后续是否需要推送消息
		/*MQTT消息推送*/
	}
}

//处理特征通告， 查找特征值得到翠英
func procBleCharacteristicNotice(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleCharacteristicNotice> DevEui %s start deal ble character response message\n", devEui)
	//通告消息无需校对

	/*将消息封装为mqtt形式，
	数据信息计算
	上行发送*/
	jsonInfo.MessageAppBody.ErrorCode = packets.Success
	jsonInfo.MessageAppBody.TLV = packets.TLV{}
	jsonInfo.MessageAppHeader.OpType = packets.Response
	jsonInfo.MessageHeader.OpType = packets.Response
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	SendDownMessage(byteToSend, jsonInfo.MessageBody.GwMac+jsonInfo.MessageBody.ModuleID, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac)
}

//处理蓝牙终端断开连接情况
//连接只做了内容
func procBleTerminalDisConnect(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procBleTerminalDisConnect> DevEui %s start deal ble terminal disconnect event message\n", devEui)
	for _, curTLV := range jsonInfo.MessageAppBody.MultiTLV {
		connCacheKey := globalutils.CreateCacheKey(globalconstants.BleDevCacheConnPrefix, curTLV.TLVPayload.ConnHandle)
		_, err := globalmemo.BleFreeCacheDevConnInfo.Get([]byte(connCacheKey))
		if err != nil {
			globallogger.Log.Errorf("<procBleTerminalDisConnect> DevEui %s has already disconnect", curTLV.TLVPayload.DevMac)
			continue
		}
		/* 发送mqtt通知， 并清空缓存*/
		globalmemo.BleFreeCacheDevConnInfo.Del([]byte(connCacheKey))
	}
}
