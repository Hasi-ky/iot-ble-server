package bleudp

import (
	"context"
	"encoding/hex"
	"errors"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/global/globaltransfersn"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"strings"
)
//重构标记

//生成特征配置请求  暂行
func GenerateCharacterConfJsonInfo(ctx context.Context, terminalInfo globalstruct.TerminalInfo, nodeChar globalstruct.ServiceCharacterNode) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, terminalInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+terminalInfo.IotModuleId)
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVCharConfReqMsg, terminalInfo.TerminalMac, nodeChar)
	if err != nil {
		return
	}
	jsonInfo.PendCtrl = globalconstants.CtrlAllMsg
	return
}

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
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+terminalInfo.IotModuleId)
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

func GenerateMainServiceFindUUIDJsonInfo(ctx context.Context, terminalInfo globalstruct.TerminalInfo, uuid string) (jsonInfo packets.JsonUdpInfo, err error) {
	jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, packets.BleRequest)
	if err != nil {
		return
	}
	jsonInfo.MessageBody, err = GenerateMessageBody(packets.BleRequest, terminalInfo)
	if err != nil {
		return
	}
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+terminalInfo.IotModuleId)
	if err != nil {
		return
	}
	jsonInfo.MessageAppBody, err = GenerateMessageAppBody(packets.TLVMainServiceByUUIDReqMsg, terminalInfo.TerminalMac)
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
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+terminalInfo.IotModuleId)
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
	jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, packets.BleRequest, terminalInfo.GwMac+terminalInfo.IotModuleId)
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
			resultJson.ModuleID = curInfo.IotModuleId
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
func GenerateMessageAppBody(TLVMsgType string, args ...interface{}) (resultJson packets.MessageAppBody, err error) {
	switch TLVMsgType {
	case packets.TLVConnectMsg: //28字节
		devConn := args[0].(globalstruct.DevConnection)
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
		scanInfo := args[0].(globalstruct.ScanInfo)
		resultJson.TLV.TLVMsgType = packets.TLVScanMsg
		resultJson.TLV.TLVLen = globalutils.ConvertDecimalToHexStr(12, globalconstants.BYTE_STR_FOUR) //固定长度
		resultJson.TLV.TLVPayload.ScanAble = globalutils.ConvertDecimalToHexStr(int(scanInfo.EnableScan), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ReserveOne = globalutils.ConvertDecimalToHexStr(0, globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanType = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanType), globalconstants.BYTE_STR_TWO)
		resultJson.TLV.TLVPayload.ScanInterval = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanInterval), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanWindow = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanWindow), globalconstants.BYTE_STR_FOUR)
		resultJson.TLV.TLVPayload.ScanTimeout = globalutils.ConvertDecimalToHexStr(int(scanInfo.ScanTimeout), globalconstants.BYTE_STR_FOUR)
	case packets.TLVMainServiceReqMsg:
		devMac := args[0].(string)
		resultJson.TLV.TLVPayload.DevMac = devMac
	case packets.TLVCharReqMsg:
		charInfo := args[0].(globalstruct.CharacterInfo)
		resultJson.TLV.TLVPayload.DevMac = charInfo.DevMac
		resultJson.TLV.TLVPayload.StartHandle = charInfo.StartHandle
		resultJson.TLV.TLVPayload.EndHandle = charInfo.EndHandle
	case packets.TLVCharConfReqMsg: //特殊处理
		devMac := args[0].(globalstruct.TerminalInfo).TerminalMac
		node := args[1].(globalstruct.ServiceCharacterNode)
		resultJson.TLV.TLVMsgType = packets.TLVCharConfReqMsg
		resultJson.TLV.TLVLen = globalutils.ConvertDecimalToHexStr(16, globalconstants.BYTE_STR_FOUR) //固定长度
		resultJson.TLV.TLVPayload.DevMac = devMac
		resultJson.TLV.TLVPayload.CCCDHandle = node.CCCDHanle
		resultJson.TLV.TLVPayload.CharHandle = node.CharacterHandle
		resultJson.TLV.TLVPayload.CCCDHandleValue = "01" //写死为notify  针对配置
	default:
		return resultJson, errors.New("unrecognized BLEMessage")
	}
	return
}

/// =============================解析=====================================
// 封装的链路消息编码字节数字
// `ctrl` 控制四部分编码生成
func EnCodeForDownUdpMessage(jsonInfo packets.JsonUdpInfo) []byte {
	var (
		enCodeResHeader, enCodeResBody, enCodeResAppMsg, enCodeResAppMsgBody, encodeResAppMsgBodyTLV strings.Builder
		encodeStr, tempStr, tempAppStr, tempAppTLVStr                                                string
	)
	if jsonInfo.PendCtrl&1 == 1 {
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.Version)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.LinkMsgFrameSN)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.LinkMsgType)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.OpType)
	}
	if (jsonInfo.PendCtrl>>1)&1 == 1 {
		enCodeResBody.WriteString(jsonInfo.MessageBody.GwMac)
		enCodeResBody.WriteString(jsonInfo.MessageBody.ModuleID)
	}
	if (jsonInfo.PendCtrl>>2)&1 == 1 {
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.SN)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.CtrlField)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.FragOffset)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.Type)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.OpType)
	}
	if (jsonInfo.PendCtrl>>3)&1 == 1 {
		enCodeResAppMsgBody.WriteString(jsonInfo.MessageAppBody.ErrorCode)
		enCodeResAppMsgBody.WriteString(jsonInfo.MessageAppBody.RespondFrame)

		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVMsgType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVLen)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanAble)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ReserveOne)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.AddrType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ServiceUUID)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.StartHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.EndHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.OpType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ReserveTwo)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ParaLength)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ServiceHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CharHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CharacterUUID)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.FeatureCfg)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CharHandleValue)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandleValue)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanPhys)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ReserveThree)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ParaValue)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanInterval)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanWindow)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanTimeout)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ConnInterval)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ConnLatency)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ConnTimeout)
	}
	switch jsonInfo.PendCtrl {
	case globalconstants.CtrlLinkedMsgHeader:
		tempStr = enCodeResHeader.String()
		encodeStr = globalutils.InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
	case globalconstants.CtrlLinkedMsgHeadWithBoy:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempStr = globalutils.InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + enCodeResBody.String()
	case globalconstants.CtrlLinkedMsgWithMsgAppHeader:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String()
		tempAppStr = globalutils.InsertString(tempAppStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempStr = globalutils.InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppStr)+len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + tempAppStr
	default:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String() + enCodeResAppMsgBody.String()
		tempAppTLVStr = encodeResAppMsgBodyTLV.String()
		tempAppTLVStr = globalutils.InsertString(tempAppTLVStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempAppStr = globalutils.ConvertDecimalToHexStr(len(tempAppTLVStr)+len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen) + tempAppStr
		tempStr = globalutils.InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+len(tempAppStr)+len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + tempAppStr + tempAppTLVStr
	}
	enCodeBytes, _ := hex.DecodeString(encodeStr)
	return enCodeBytes
}
