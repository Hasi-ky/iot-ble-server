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
	"iot-ble-server/global/globaltransfersn"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"strconv"

	"github.com/redis/go-redis/v9"
)

//生成主服务发现,单设备
func GenerateMainServiceFindJsonInfo(ctx context.Context, terminalInfo globalstruct.TerminalInfo) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, terminalInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+globalutils.ConvertDecimalToHexStr(int(terminalInfo.IotModuleId), globalconstants.BYTE_STR_FOUR))
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVMainServiceReqMsg, terminalInfo.TerminalMac)
	if err != nil {
		return
	}
	jsonInfo.PendCtrl = globalconstants.CtrlAllMsg
	return
}

//生成特征发现
func GenerateCharacterFind(ctx context.Context, terminalInfo globalstruct.TerminalInfo, charInfo globalstruct.CharacterInfo) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, terminalInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+globalutils.ConvertDecimalToHexStr(int(terminalInfo.IotModuleId), globalconstants.BYTE_STR_FOUR))
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVCharReqMsg, charInfo)
	if err != nil {
		return
	}
	jsonInfo.PendCtrl = globalconstants.CtrlAllMsg
	return
}

//生成设备连接信息，
//生成的状态是全部消息，all
func GenerateDevConnJsonInfo(ctx context.Context, terminalInfo globalstruct.TerminalInfo,
	devConn globalstruct.DevConnection) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, terminalInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+globalutils.ConvertDecimalToHexStr(int(terminalInfo.IotModuleId), globalconstants.BYTE_STR_FOUR))
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVConnectMsg, devConn)
	if err != nil {
		return
	}
	jsonInfo.PendCtrl = globalconstants.CtrlAllMsg
	return
}

//扫描状态
func GenerateScanJsonInfo(ctx context.Context, scanInfo globalstruct.ScanInfo) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, scanInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, scanInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, scanInfo.GwMac+globalutils.ConvertDecimalToHexStr(int(scanInfo.IotModuleId), globalconstants.BYTE_STR_FOUR))
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVScanMsg, scanInfo)
	if err != nil {
		return
	}
	jsonInfo.PendCtrl = globalconstants.CtrlAllMsg
	return
}

//信息头生成
//此时长度未生成
func GenerateMessageHeader(ctx context.Context, gwMac, msgType string) (resultJson packets.MessageHeader, err error) {
	transferLinkedKey := globalconstants.GwRedisSNTransfer + gwMac
	var curFrameNum int64
	resultJson.Version = packets.Version3
	switch msgType {
	case packets.BleRequest:
		resultJson.LinkMsgType = packets.BleRequest
		resultJson.OpType = packets.RequireWithResp
	default:
		return resultJson, errors.New("unrecognized LinkMessage")
	}
	if config.C.General.UseRedis {
		curFrameNum, err = globalredis.RedisCache.Incr(ctx, transferLinkedKey).Result()
		if err != nil {
			return packets.MessageHeader{}, err
		}
		if curFrameNum > int64(globalconstants.MaxMessageFourLimit) {
			globalredis.RedisCache.Del(ctx, transferLinkedKey)
			curFrameNum, _ = globalredis.RedisCache.Incr(ctx, transferLinkedKey).Result()
		}
	} else {
		globaltransfersn.TransferSN.Lock()
		if globaltransfersn.TransferSN.SN[transferLinkedKey] < 1 {
			globaltransfersn.TransferSN.SN[transferLinkedKey] = 1
		} else {
			globaltransfersn.TransferSN.SN[transferLinkedKey] = globaltransfersn.TransferSN.SN[transferLinkedKey] + 1
			if globaltransfersn.TransferSN.SN[transferLinkedKey] > globalconstants.MaxMessageFourLimit {
				globaltransfersn.TransferSN.SN[transferLinkedKey] = 1
			}
		}
		curFrameNum = int64(globaltransfersn.TransferSN.SN[transferLinkedKey])
		globaltransfersn.TransferSN.Unlock()
	}
	resultJson.LinkMsgFrameSN = globalutils.ConvertDecimalToHexStr(int(curFrameNum), globalconstants.BYTE_STR_EIGHT)
	return
}

//消息体消息生成
func GenerateMessageBody(msgType string, bodyInfo interface{}) (resultJson packets.MessageBody, err error) {
	switch msgType {
	case packets.BleRequest:
		if curInfo, ok := bodyInfo.(globalstruct.TerminalInfo); ok {
			resultJson.GwMac = curInfo.GwMac
			resultJson.ModuleID = globalutils.ConvertDecimalToHexStr(int(curInfo.IotModuleId), globalconstants.BYTE_STR_FOUR)
		} else if curInfo, ok := bodyInfo.(globalstruct.ScanInfo); ok {
			resultJson.GwMac = curInfo.GwMac
			resultJson.ModuleID = globalutils.ConvertDecimalToHexStr(int(curInfo.IotModuleId), globalconstants.BYTE_STR_FOUR)
		}
	default:
		return resultJson, errors.New("unrecognized LinkMessage")
	}
	return
}

//应用消息头生成
//缺少总长度
func GenerateMessageAppHeader(ctx context.Context, appMsgType string,
	macWithModule string) (resultJson packets.MessageAppHeader, err error) {
	var (
		curFrameNum    int64
		transferDevKey = globalconstants.BleRedisSNTransfer + macWithModule
	)
	resultJson.CtrlField = globalconstants.CtrlField
	resultJson.FragOffset = globalconstants.FragOffset
	switch appMsgType {
	case packets.BleRequest:
		resultJson.OpType = packets.RequireWithResp
	default:
		return resultJson, errors.New("unrecognized BLEMessage")
	}
	if config.C.General.UseRedis {
		curFrameNum, err = globalredis.RedisCache.Incr(ctx, transferDevKey).Result()
		if err != nil {
			return
		}
		if curFrameNum > int64(globalconstants.MaxMessageTwoLimit) {
			globalredis.RedisCache.Del(ctx, transferDevKey)
			curFrameNum, _ = globalredis.RedisCache.Incr(ctx, transferDevKey).Result()
		}
	} else {
		globaltransfersn.TransferSN.Lock()
		if globaltransfersn.TransferSN.SN[transferDevKey] < 1 {
			globaltransfersn.TransferSN.SN[transferDevKey] = 1
		} else {
			globaltransfersn.TransferSN.SN[transferDevKey] = globaltransfersn.TransferSN.SN[transferDevKey] + 1
			if globaltransfersn.TransferSN.SN[transferDevKey] > globalconstants.MaxMessageTwoLimit {
				globaltransfersn.TransferSN.SN[transferDevKey] = 1
			}
		}
		curFrameNum = int64(globaltransfersn.TransferSN.SN[transferDevKey])
		globaltransfersn.TransferSN.Unlock()
	}
	resultJson.Type = globalutils.ConvertDecimalToHexStr(int(curFrameNum), 4)
	return
}

//应用消息生成
func GenerateMessageAppBody(TLVMsgType string, args interface{}) (resultJson packets.MessageAppBody, err error) {
	switch TLVMsgType {
	case packets.TLVConnectMsg: //28字节
		devConn := args.(globalstruct.DevConnection)
		resultJson.TLV.TLVMsgType = packets.TLVConnectMsg
		resultJson.TLV.TLVLen = globalutils.ConvertDecimalToHexStr(28, globalconstants.BYTE_STR_FOUR) //固定长度
		resultJson.TLV.TLVPayload.ReserveOne = globalutils.ConvertDecimalToHexStr(0, globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.AddrType = globalutils.ConvertDecimalToHexStr(int(devConn.AddrType), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.DevMac = devConn.DevMac
		resultJson.TLV.TLVPayload.OpType = globalutils.ConvertDecimalToHexStr(int(devConn.OpType), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ReserveTwo = globalutils.ConvertDecimalToHexStr(0, globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanType = globalutils.ConvertDecimalToHexStr(int(devConn.ScanType), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ReserveThree = globalutils.ConvertDecimalToHexStr(0, globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanInterval = globalutils.ConvertDecimalToHexStr(int(devConn.ScanInterval), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanWindow = globalutils.ConvertDecimalToHexStr(int(devConn.ScanWindow), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanTimeout = globalutils.ConvertDecimalToHexStr(int(devConn.ScanTimeout), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ConnInterval = globalutils.ConvertDecimalToHexStr(int(devConn.ConnInterval), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ConnLatency = globalutils.ConvertDecimalToHexStr(int(devConn.ConnLatency), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.SupervisionWindow = globalutils.ConvertDecimalToHexStr(int(devConn.SupervisionWindow), globalconstants.BYTE_STR_FOUR)
	case packets.TLVScanMsg: //12字节
		scanInfo := args.(globalstruct.ScanInfo)
		resultJson.TLV.TLVMsgType = packets.TLVScanMsg
		resultJson.TLV.TLVLen = globalutils.ConvertDecimalToHexStr(12, globalconstants.BYTE_STR_FOUR) //固定长度
		resultJson.TLV.TLVPayload.ScanAble = globalutils.ConvertDecimalToHexStr(int(scanInfo.EnableScan), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ReserveOne = globalutils.ConvertDecimalToHexStr(0, globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanType = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanType), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanInterval = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanInterval), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanWindow = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanWindow), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanTimeout = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanTimeout), globalconstants.BYTE_STR_FOUR)
	case packets.TLVMainServiceReqMsg:
		devMac := args.(string)
		resultJson.TLV.TLVPayload.DevMac = devMac
	case packets.TLVCharReqMsg:
		charInfo := args.(globalstruct.CharacterInfo)
		resultJson.TLV.TLVPayload.DevMac = charInfo.DevMac
		resultJson.TLV.TLVPayload.StartHandle = charInfo.StartHandle
		resultJson.TLV.TLVPayload.EndHandle = charInfo.EndHandle
	default:
		return resultJson, errors.New("unrecognized BLEMessage")
	}
	return
}

//特征发现
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
		DevMac: devMac,
		StartHandle: startHandle,
		EndHandle:  endHandle,
	})
	if err1 != nil {
		return err1
	}
	byteToSend := EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	SendMsgBeforeDown(ctx, byteToSend, int(curSN), jsonInfo.MessageBody.GwMac + jsonInfo.MessageBody.ModuleID, terminal.GwMac, packets.TLVConnectMsg)
	return err
}
