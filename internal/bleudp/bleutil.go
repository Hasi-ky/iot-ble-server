package bleudp

import (
	"encoding/hex"
	"errors"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"iot-ble-server/internal/storage"
	"strconv"
	"strings"
	"time"
)

//验证 `H3C` 报文
func checkMsgSafe(data []byte) bool {
	messageLen, _ := strconv.ParseInt(hex.EncodeToString(append(data[:0:0], data[2:4]...)), 16, 0)
	return int(messageLen) == len(data)
}

//上行 `UDP` 消息解析
func parseUdpMessage(data []byte, rinfo dgram.RInfo) (packets.JsonUdpInfo, error) {
	var err error
	offset, jsonUdpInfo, dataLen := 0, packets.JsonUdpInfo{}, len(data)
	jsonUdpInfo.Rinfo = rinfo
	jsonUdpInfo.MessageHeader = ParseMessageHeader(data, &offset)
	if globalutils.JudgePacketLenthLimit(offset, dataLen) {
		return jsonUdpInfo, nil
	}
	jsonUdpInfo.MessageBody, err = ParseMessageBody(data, jsonUdpInfo.MessageHeader.LinkMsgType, &offset)
	if err != nil || globalutils.JudgePacketLenthLimit(offset, dataLen) {
		return jsonUdpInfo, err
	}
	jsonUdpInfo.MessageAppHeader = ParseMessageAppHeader(data, &offset)
	if globalutils.JudgePacketLenthLimit(offset, dataLen) {
		return jsonUdpInfo, nil
	}
	jsonUdpInfo.MessageAppBody, err = ParseMessageAppBody(data, &offset, jsonUdpInfo.MessageBody.GwMac, jsonUdpInfo.MessageAppHeader.Type)
	if err != nil {
		return jsonUdpInfo, err
	}
	return jsonUdpInfo, nil
}

func ParseMessageHeader(data []byte, offset *int) packets.MessageHeader {
	*offset = 12
	return packets.MessageHeader{
		Version:           hex.EncodeToString(append(data[:0:0], data[0:2]...)),
		LinkMessageLength: hex.EncodeToString(append(data[:0:0], data[2:4]...)),
		LinkMsgFrameSN:    hex.EncodeToString(append(data[:0:0], data[4:8]...)),
		LinkMsgType:       hex.EncodeToString(append(data[:0:0], data[8:10]...)),
		OpType:            hex.EncodeToString(append(data[:0:0], data[10:12]...)),
	}
}

func ParseMessageAppHeader(data []byte, offset *int) packets.MessageAppHeader {
	tempOffset := *offset
	*offset += 12
	return packets.MessageAppHeader{
		TotalLen:   hex.EncodeToString(append(data[:0:0], data[tempOffset:tempOffset+2]...)),
		SN:         hex.EncodeToString(append(data[:0:0], data[tempOffset+2:tempOffset+4]...)),
		CtrlField:  hex.EncodeToString(append(data[:0:0], data[tempOffset+4:tempOffset+6]...)),
		FragOffset: hex.EncodeToString(append(data[:0:0], data[tempOffset+6:tempOffset+8]...)),
		Type:       hex.EncodeToString(append(data[:0:0], data[tempOffset+8:tempOffset+10]...)),
		OpType:     hex.EncodeToString(append(data[:0:0], data[tempOffset+10:tempOffset+12]...)),
	}
}

//目前写法为已知一个TLV，后续嵌入其余TLV
func ParseMessageBody(data []byte, msgType string, offset *int) (packets.MessageBody, error) {
	msgBody := packets.MessageBody{}
	var err error
	switch msgType {
	case packets.GatewayDevInfo:
		msgBody.GwMac = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+6]...))
		msgBody.ModuleID = hex.EncodeToString(append(data[:0:0], data[*offset+6:*offset+8]...))
		msgBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset+8:*offset+10]...))
		msgBody.TLV.TLVMsgType = hex.EncodeToString(append(data[:0:0], data[*offset+10:*offset+12]...))
		msgBody.TLV.TLVLen = hex.EncodeToString(append(data[:0:0], data[*offset+12:*offset+14]...))
		msgBody.TLV.TLVPayload.TLVReserve = make([]packets.TLV, 0)
		dfsForAllTLV(data, &msgBody.TLV.TLVPayload.TLVReserve, *offset+14)
	case packets.IotModuleRset:
		msgBody.GwMac = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+6]...))
		msgBody.ModuleID = hex.EncodeToString(append(data[:0:0], data[*offset+6:*offset+8]...))
		msgBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset+8:*offset+10]...))
	case packets.IotModuleStatusChange:
		msgBody.GwMac = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+6]...))
		msgBody.ModuleID = hex.EncodeToString(append(data[:0:0], data[*offset+6:*offset+8]...))
		msgBody.TLV = packets.TLV{}
		msgBody.TLV.TLVMsgType = hex.EncodeToString(append(data[:0:0], data[*offset+8:*offset+10]...))
		msgBody.TLV.TLVLen = hex.EncodeToString(append(data[:0:0], data[*offset+10:*offset+12]...))
		msgBody.TLV.TLVPayload = packets.TLVFeature{
			IotModuleId:           hex.EncodeToString(append(data[:0:0], data[*offset+12:*offset+14]...)),
			IotModuleStatus:       hex.EncodeToString(append(data[:0:0], data[*offset+14:*offset+15]...)),
			IotModuleChangeReason: hex.EncodeToString(append(data[:0:0], data[*offset+15:*offset+16]...)),
		}
	default:
		err = errors.New("ParseMessageBody: from: " + msgBody.GwMac + " unable to recognize message type with " + msgType)
	}
	*offset = len(data)
	return msgBody, err
}

func ParseMessageAppBody(data []byte, offset *int, devMac string, msgType string) (packets.MessageAppBody, error) {
	var err error
	parseMessageAppBody := packets.MessageAppBody{}
	switch msgType {
	case packets.BleConfirm:
		parseMessageAppBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+2]...))
		parseMessageAppBody.RespondFrame = hex.EncodeToString(append(data[:0:0], data[*offset+2:*offset+6]...))
	case packets.BleResponse:
		parseMessageAppBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+2]...))
		parseMessageAppBody.RespondFrame = hex.EncodeToString(append(data[:0:0], data[*offset+2:*offset+6]...))
		parseMessageAppBody.TLV = GetDevResponseTLV(data, *offset+6)
	case packets.BleGetConnDevList:
		parseMessageAppBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+2]...))
		parseMessageAppBody.DevSum = hex.EncodeToString(append(data[:0:0], data[*offset+2:*offset+4]...))
		parseMessageAppBody.Reserve = hex.EncodeToString(append(data[:0:0], data[*offset+4:*offset+6]...))
		parseMessageAppBody.TLV = GetDevResponseTLV(data, *offset+6)
	case packets.BleCharacteristic:
		parseMessageAppBody.ErrorCode = hex.EncodeToString(append(data[:0:0], data[*offset:*offset+2]...))
	case packets.BleBoardcast, packets.BleTerminalEvent:  //附加TLV无处理
		parseMessageAppBody.TLV = GetDevResponseTLV(data, *offset)
	default:
		err = errors.New("ParseMessageAppBody: from: " + devMac + " unable to recognize message type with " + parseMessageAppBody.TLV.TLVMsgType)
	}
	return parseMessageAppBody, err
}

/// =============================解析=====================================
// 封装的链路消息编码字节数字
// `ctrl` 控制四部分编码生成
func EnCodeForDownUdpMessage(jsonInfo packets.JsonUdpInfo, ctrl int) []byte {
	var (
		enCodeResHeader, enCodeResBody, enCodeResAppMsg, enCodeResAppMsgBody, encodeResAppMsgBodyTLV strings.Builder
		encodeStr, tempStr, tempAppStr, tempAppTLVStr                                                string
	)
	if ctrl&1 == 1 {
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.Version)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.LinkMsgFrameSN)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.LinkMsgType)
		enCodeResHeader.WriteString(jsonInfo.MessageHeader.OpType)
	}
	if (ctrl>>1)&1 == 1 {
		enCodeResBody.WriteString(jsonInfo.MessageBody.GwMac)
		enCodeResBody.WriteString(jsonInfo.MessageBody.ModuleID)
	}
	if (ctrl>>2)&1 == 1 {
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.SN)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.CtrlField)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.FragOffset)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.Type)
		enCodeResAppMsg.WriteString(jsonInfo.MessageAppHeader.OpType)
	}
	if (ctrl>>3)&1 == 1 {
		enCodeResAppMsgBody.WriteString(jsonInfo.MessageAppBody.ErrorCode)
		enCodeResAppMsgBody.WriteString(jsonInfo.MessageAppBody.RespondFrame)

		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVMsgType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVLen)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanAble)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ReserveOne)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.AddrType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.UUID)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.StartHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.EndHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.OpType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ReserveTwo)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ParaLength)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ScanType)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.Handle) //char handle ？
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.CCCDHandle)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.FeatureCfg)
		encodeResAppMsgBodyTLV.WriteString(jsonInfo.MessageAppBody.TLV.TLVPayload.ValueHandle)
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
	switch ctrl {
	case globalconstants.CtrlLinkedMsgHeader:
		tempStr = enCodeResHeader.String()
		encodeStr = InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
	case globalconstants.CtrlLinkedMsgHeadWithBoy:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempStr = InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + enCodeResBody.String()
	case globalconstants.CtrlLinkedMsgWithMsgAppHeader:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String()
		tempAppStr = InsertString(tempAppStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempStr = InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppStr)+len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + tempAppStr
	default:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String() + enCodeResAppMsgBody.String()
		tempAppTLVStr = encodeResAppMsgBodyTLV.String()
		tempAppTLVStr = InsertString(tempAppTLVStr,
			globalutils.ConvertDecimalToHexStr(len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempAppStr = globalutils.ConvertDecimalToHexStr(len(tempAppTLVStr)+len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen) + tempAppStr
		tempStr = InsertString(tempStr,
			globalutils.ConvertDecimalToHexStr(len(tempStr)+len(tempAppStr)+len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + tempAppStr + tempAppTLVStr
	}
	enCodeBytes, _ := hex.DecodeString(encodeStr)
	return enCodeBytes
}

// `str` 原串 | `inset` 插入串 | `index` 插入点 |
func InsertString(str, insert string, index int) string {
	if index < 0 || index >= len(str) {
		globallogger.Log.Errorln("<InsertString>: please check param with |index| or |str|")
		return ""
	}
	pre, tail := str[:index], str[index:]
	return pre + insert + tail
}

// `frame` 帧序列 | `reSendTime` 发送次数 | `devEui` 唯一标识 标识消息队列 设备mac or 网关mac + module id | `curQueue` 消息队列 |
func ResendMessage(frameSN, reSendTimes int, curQueue *storage.CQueue, devEui string) {
	if reSendTimes >= 3 {
		globallogger.Log.Errorf("<ResendMessage> : devEui: %s has issue with the current retransmission message device", devEui)
		return
	}
	timer := time.NewTimer((time.Second * time.Duration(config.C.General.KeepAliveTime)))
	defer timer.Stop()
	<-timer.C
	if curQueue.Len() > 0 && frameSN == curQueue.Peek().PendFrame {
		globallogger.Log.Warnf("<ResendMessage> : devEui: %s resend msg, the msg type is %s", devEui, curQueue.Peek().MessageHeader.LinkMsgType)
		jsonInfo := curQueue.Peek()
		sendBytes := EnCodeForDownUdpMessage(jsonInfo, jsonInfo.PendCtrl)
		SendDownMessage(sendBytes, devEui, jsonInfo.MessageBody.GwMac)
		ResendMessage(frameSN, reSendTimes+1, curQueue, devEui)
	}
}

//异步方法， 自旋等待,在下发消息前，需要从currentMap中取到对应消息队列
func SendMsgBeforeDown(sendBytes []byte, frameSN int, curQueue *storage.CQueue, devEui string) {
	for curQueue.Peek().PendFrame != frameSN {
		globallogger.Log.Warnf("<SendMsgBeforeDown>: devEui: %s wait %d message to send down", devEui, frameSN)
		time.Sleep(time.Microsecond * 500)
	}
	SendDownMessage(sendBytes, devEui, curQueue.Peek().MessageBody.GwMac)
	go ResendMessage(frameSN, 0, curQueue, devEui)
}

//TLV分解过程
//`搜TLV`
func dfsForAllTLV(data []byte, restore *[]packets.TLV, index int) {
	for i := index; i < len(data); {
		tempTLV := packets.TLV{
			TLVMsgType: hex.EncodeToString(append(data[:0:0], data[i:i+2]...)),
			TLVLen:     hex.EncodeToString(append(data[:0:0], data[i+2:i+4]...)),
		}
		tempLen, _ := strconv.ParseInt(tempTLV.TLVLen, 16, 32)
		switch tempTLV.TLVMsgType {
		case packets.TLVGatewayDescribeMsg:
			tempTLV.TLVPayload.TLVReserve = make([]packets.TLV, 0)
			dfsForAllTLV(data, &tempTLV.TLVPayload.TLVReserve, i+4)
		case packets.TLVIotModuleDescribeMsg:
			tempTLV.TLVPayload.Port = hex.EncodeToString(append(data[:0:0], data[i+4:i+5]...))
			tempTLV.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[i+5:i+6]...))
			tempTLV.TLVPayload.TLVReserve = make([]packets.TLV, 0)
			dfsForAllTLV(data, &tempTLV.TLVPayload.TLVReserve, index+6)
		case packets.TLVIotModuleEventMsg:
			tempTLV.TLVPayload.Event = hex.EncodeToString(append(data[:0:0], data[i+4:i+5]...))
			tempTLV.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[i+5:i+6]...))
			tempTLV.TLVPayload.IotModuleId = hex.EncodeToString(append(data[:0:0], data[i+6:i+8]...))
			dfsForAllTLV(data, &tempTLV.TLVPayload.TLVReserve, index+8)
		case packets.TLVServiceMsg:
			tempTLV.TLVPayload.Primary = hex.EncodeToString(append(data[:0:0], data[i+4:i+5]...))
			tempTLV.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[i+5:i+6]...))
			tempTLV.TLVPayload.Handle = hex.EncodeToString(append(data[:0:0], data[i+6:i+8]...))
			tempTLV.TLVPayload.UUID = hex.EncodeToString(append(data[:0:0], data[i+8:]...))
		case packets.TLVCharReqMsg:
			tempTLV.TLVPayload.Properties = hex.EncodeToString(append(data[:0:0], data[i+4:i+5]...))
			tempTLV.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[i+5:i+6]...))
			tempTLV.TLVPayload.ServiceHandle = hex.EncodeToString(append(data[:0:0], data[i+6:i+8]...))
			tempTLV.TLVPayload.CharHandle = hex.EncodeToString(append(data[:0:0], data[i+8:i+10]...))
			tempTLV.TLVPayload.UUID = hex.EncodeToString(append(data[:0:0], data[i+10:]...))
		case packets.TLVCharDescribeMsg:
			tempTLV.TLVPayload.ServiceHandle = hex.EncodeToString(append(data[:0:0], data[i+4:i+6]...))
			tempTLV.TLVPayload.CharHandle = hex.EncodeToString(append(data[:0:0], data[i+6:i+8]...))
			tempTLV.TLVPayload.DescriptorHandle = hex.EncodeToString(append(data[:0:0], data[i+8:i+10]...))
			tempTLV.TLVPayload.UUID = hex.EncodeToString(append(data[:0:0], data[i+10:]...))
		case packets.TLVGatewayTypeMsg:
			tempTLV.TLVPayload.GwType = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		case packets.TLVGatewaySNMsg:
			tempTLV.TLVPayload.GwSN = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		case packets.TLVGatewayMACMsg:
			tempTLV.TLVPayload.GwMac = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		case packets.TLVIotModuleMsg:
			tempTLV.TLVPayload.IotModuleType = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		case packets.TLVIotModuleSNMsg:
			tempTLV.TLVPayload.IotModuleSN = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		case packets.TLVIotModuleMACMsg:
			tempTLV.TLVPayload.IotModuleMac = hex.EncodeToString(append(data[:0:0], data[i+4:i+int(tempLen)]...))
		}
		*restore = append(*restore, tempTLV)
		i += int(tempLen)
	}
}

//`针对终端响应BLE 应答TLV建立`
//三大消息中可直接使用，响应消息调tlv内容
func GetDevResponseTLV(data []byte, index int) (res packets.TLV) {
	res.TLVMsgType = hex.EncodeToString(append(data[:0:0], data[index:index+2]...))
	res.TLVLen = hex.EncodeToString(append(data[:0:0], data[index+2:index+4]...))
	switch res.TLVMsgType {
	case packets.TLVScanRespMsg:
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+4:index+6]...))
		res.TLVPayload.ScanStatus = hex.EncodeToString(append(data[:0:0], data[index+6:index+7]...))
		res.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[index+7:index+8]...))
		res.TLVPayload.ScanType = hex.EncodeToString(append(data[:0:0], data[index+8:index+9]...))
		res.TLVPayload.ScanPhys = hex.EncodeToString(append(data[:0:0], data[index+9:index+10]...))
		res.TLVPayload.ScanInterval = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		res.TLVPayload.ScanWindow = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...))
		res.TLVPayload.ScanTimeout = hex.EncodeToString(append(data[:0:0], data[index+14:index+16]...))
	case packets.TLVConnectRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		res.TLVPayload.ConnStatus = hex.EncodeToString(append(data[:0:0], data[index+12:index+13]...))
		res.TLVPayload.PHY = hex.EncodeToString(append(data[:0:0], data[index+13:index+14]...))
		res.TLVPayload.Handle = hex.EncodeToString(append(data[:0:0], data[index+14:index+16]...))
		res.TLVPayload.ConnInterval = hex.EncodeToString(append(data[:0:0], data[index+16:index+18]...))
		res.TLVPayload.ConnLatency = hex.EncodeToString(append(data[:0:0], data[index+18:index+20]...))
		res.TLVPayload.ConnTimeout = hex.EncodeToString(append(data[:0:0], data[index+20:index+22]...))
		res.TLVPayload.MTUSize = hex.EncodeToString(append(data[:0:0], data[index+22:index+24]...))
	case packets.TLVMainServiceRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		res.TLVPayload.ServiceSum = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...))
		dfsForAllTLV(data, &res.TLVPayload.TLVReserve, index+14)
	case packets.TLVCharRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ServiceHandle = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...)) //服务handle
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...))
		res.TLVPayload.FeatureSum = hex.EncodeToString(append(data[:0:0], data[index+14:index+16]...))
		dfsForAllTLV(data, &res.TLVPayload.TLVReserve, index+16)
	case packets.TLVCharConfRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		res.TLVPayload.CharHandle = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...))
		res.TLVPayload.CCCDHandle = hex.EncodeToString(append(data[:0:0], data[index+14:index+16]...))
		res.TLVPayload.FeatureCfg = hex.EncodeToString(append(data[:0:0], data[index+16:index+17]...)) //特征配置
	case packets.TLVCharOptRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		res.TLVPayload.ParaLength = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...))
		res.TLVPayload.CharHandle = hex.EncodeToString(append(data[:0:0], data[index+14:index+16]...))
		res.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[index+16:index+18]...))
		res.TLVPayload.ParaValue = hex.EncodeToString(append(data[:0:0], data[index+18:]...))
	case packets.TLVMainServiceByUUIDRespMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.ErrorCode = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
		dfsForAllTLV(data, &res.TLVPayload.TLVReserve, index+12)
	case packets.TLVDeviceListMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.Handle = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))
	case packets.TLVBroadcastMsg:
		res.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[index+4:index+5]...))
		res.TLVPayload.AddrType = hex.EncodeToString(append(data[:0:0], data[index+5:index+6]...))
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+6:index+12]...))
		res.TLVPayload.RSSI = hex.EncodeToString(append(data[:0:0], data[index+12:index+13]...))
		res.TLVPayload.ADType = hex.EncodeToString(append(data[:0:0], data[index+13:index+14]...))
		res.TLVPayload.NoticeContent = hex.EncodeToString(append(data[:0:0], data[index+14:]...))
	case packets.TLVDisconnectMsg:
		res.TLVPayload.DevMac = hex.EncodeToString(append(data[:0:0], data[index+4:index+10]...))
		res.TLVPayload.Handle = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...)) //连接句柄
		res.TLVPayload.DisConnReason = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...)) //断开事件原因
	default:
		globallogger.Log.Errorf("<GetDevResponseTLV>: the corresponding message type cannot be recognized. please troubleshoot the error!")
		return packets.TLV{}
	}
	return
}
