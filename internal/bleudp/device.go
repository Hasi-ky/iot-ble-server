package bleudp

import (
	"context"
	"encoding/json"
	"errors"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"iot-ble-server/internal/storage"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//特征发现, 但是在有了指定UUID发现以后，这个函数的功能可以暂时搁置
func ResumeCharacterFind(ctx context.Context, devMac, startHandle, endHandle string) error {
	var (
		err      error
		terminal globalstruct.TerminalInfo
		byteInfo []byte
		strInfo  string
	)
	if config.C.General.UseRedis {
		strInfo, err = globalredis.RedisCache.HGet(ctx, globalconstants.BleDevInfoCachePrefix, devMac).Result()
		if err != nil {
			if err != redis.Nil {
				globallogger.Log.Errorf("<ResumeCharacterFind> DevEui %s has redis error %v\n", devMac, err)
			} else {
				globallogger.Log.Errorf("<ResumeCharacterFind> DevEui %s absence terminal information %v\n", devMac, err)
			}
			return errors.New("cache has error")
		}
		byteInfo = []byte(strInfo)

	} else {
		byteInfo, err = globalmemo.BleFreeCacheDevInfo.Get([]byte(devMac))
		if err != nil {
			globallogger.Log.Errorf("<ResumeCharacterFind> DevEui %s has cache error %v\n", devMac, err)
			return errors.New("cache has error")
		}
	}
	err = json.Unmarshal(byteInfo, &terminal)
	if err != nil {
		return err
	}
	jsonInfo, err1 := GenerateCharacterFind(ctx, terminal, globalstruct.CharacterInfo{
		DevMac:      devMac,
		StartHandle: startHandle,
		EndHandle:   endHandle,
	})
	if err1 != nil {
		return err1
	}
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	SendMsgBeforeDown(ctx, byteToSend, int(curSN), jsonInfo.MessageBody.GwMac+jsonInfo.MessageBody.ModuleID, terminal.GwMac, packets.TLVConnectMsg)
	return err
}

//主服务发现
//在此之前已经建立连接了
func SearchMainService(ctx context.Context, terminal globalstruct.TerminalInfo) error {
	jsonInfo, err := GenerateMainServiceFindJsonInfo(ctx, terminal)
	if err != nil {
		return err
	}
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	SendMsgBeforeDown(ctx, byteToSend, int(curSN), terminal.TerminalMac, terminal.GwMac, packets.TLVConnectMsg)
	return err
}

//针对特定uuid的主服务发现
func SearchMainServiceUUID(ctx context.Context, terminal globalstruct.TerminalInfo, UUID string) error {
	jsonInfo, err := GenerateMainServiceFindUUIDJsonInfo(ctx, terminal, UUID)
	if err != nil {
		return err
	}
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	SendMsgBeforeDown(ctx, byteToSend, int(curSN), terminal.TerminalMac, terminal.GwMac, packets.TLVConnectMsg)
	return err
}

//针对特征配置下发
func CharacterConfig(ctx context.Context, terminal globalstruct.TerminalInfo, curUUIDForChar map[string]globalstruct.ServiceCharacterNode, serviceKey string) error {
	var err error
	for _, node := range curUUIDForChar {
		if node.CharUUID == "0003" {
			jsonInfo, err := GenerateCharacterConfJsonInfo(ctx, terminal, node)
			if err != nil {
				return err
			}
			byteToSend := EnCodeForDownUdpMessage(jsonInfo)
			curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
			globalmemo.BleFreeCacheUpDown.Set([]byte(terminal.TerminalMac+jsonInfo.MessageAppHeader.SN), []byte(serviceKey), 0)
			SendMsgBeforeDown(ctx, byteToSend, int(curSN), terminal.TerminalMac, terminal.GwMac, packets.TLVConnectMsg)
		}
	}
	return err
}

//在解析到特定的广播报文后，进行连接处理

func ConnectBleDev(ctx context.Context, terminal globalstruct.TerminalInfo) error {
	jsonInfo, err := GenerateDevConnJsonInfo(ctx, terminal, globalstruct.DevConnection{
		DevMac:            terminal.TerminalMac,
		AddrType:          globalconstants.BLE_CONN_DEV_ADDR_TYPE,
		OpType:            globalconstants.BLE_CONN_OPTYPE,
		ScanType:          globalconstants.BLE_CONN_SCAN_TYPE,
		ScanInterval:      globalconstants.BLE_CONN_SCAN_INTERVAL,
		ScanWindow:        globalconstants.BLE_CONN_SCAN_WINDOWS,
		ScanTimeout:       globalconstants.BLE_CONN_SCAN_TIMEOUT,
		ConnInterval:      globalconstants.BLE_CONN_INTERVAL,
		ConnLatency:       globalconstants.BLE_CONN_LATENCY,
		SupervisionWindow: globalconstants.BLE_CONN_TIMEOUT,
	})
	if err != nil {
		globallogger.Log.Errorf("<connectDev> generate down data failed : %v\n", err)
		return err
	}
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	SendMsgBeforeDown(ctx, byteToSend, int(curSN), terminal.TerminalMac, terminal.GwMac, packets.TLVConnectMsg)
	//开启异步线程检测连接是否超时
	go func() {
		timeOutMark := terminal.TerminalMac + strconv.Itoa(int(curSN))
		judgeTimeOut := false
		listCacheKey := globalutils.CreateCacheKey(globalconstants.BleDevCacheMessagePrefix, terminal.TerminalMac)
		<-time.After(time.Duration(globalconstants.BLE_CONN_TIMEOUT * 6 * time.Millisecond))
		if config.C.General.UseRedis {
			queueByte, err := globalredis.RedisCache.LIndex(ctx, listCacheKey, 0).Result()
			if err != nil {
				globallogger.Log.Errorf("<ConnectBleDev> devEui: %s redis has error %v, frameSN sendDown failed\n", terminal.TerminalMac, err)
				return
			}
			var headNode storage.NodeCache
			err = json.Unmarshal([]byte(queueByte), &headNode)
			if err != nil {
				globallogger.Log.Errorf("<ConnectBleDev> devEui: %s data has error %v, frameSN sendDown failed\n", terminal.TerminalMac, err)
				judgeTimeOut = true
				goto Final
			}
			if headNode.FirstMark == timeOutMark { //不相等则为等待下发或者已经过了
				judgeTimeOut = true
				globallogger.Log.Errorf("<ConnectBleDev> devEui: %s connect timout\n", terminal.TerminalMac)
				globalredis.RedisCache.LPop(ctx, listCacheKey) //超时消息
			}
		} else {
			curQueue, ok := globalmemo.MemoCacheDev.Get(listCacheKey)
			if !ok {
				judgeTimeOut = true
				globallogger.Log.Errorf("<ConnectBleDev>: devEui: %s memo has error\n", terminal.TerminalMac)
				goto Final
			}
			if curQueue.(*storage.CQueue).Peek().FirstMark == timeOutMark {
				judgeTimeOut = true
				globallogger.Log.Errorf("<ConnectBleDev> devEui: %s connect timout\n", terminal.TerminalMac)
				curQueue.(*storage.CQueue).Dequeue() //超时处理
			}
		}
		/*处理超时*/
	Final:
		if judgeTimeOut {
			//todo 上送超时mqtt
		}

	}()
	return nil
}
