package bleudp

import (
	"encoding/hex"
	"errors"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
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
	jsonUdpInfo := packets.JsonUdpInfo{}
	messageHeader := ParseMessageHeader(data)
	messageBody, offset, err := ParseMessageBody(data, messageHeader.LinkMsgType)
	if err != nil {
		return packets.JsonUdpInfo{}, err
	}
	messageAppHeader, offset := ParseMessageAppHeader(data, offset)
	messageAppBody, err := ParseMessageAppBody(data, offset+10, messageBody.GwMac)
	if err != nil {
		return packets.JsonUdpInfo{}, err
	}
	jsonUdpInfo.MessageHeader = messageHeader
	jsonUdpInfo.MessageBody = messageBody
	jsonUdpInfo.MessageAppHeader = messageAppHeader
	jsonUdpInfo.MessageAppBody = messageAppBody
	jsonUdpInfo.Rinfo = rinfo
	return jsonUdpInfo, nil
}

func ParseMessageHeader(data []byte) packets.MessageHeader {
	return packets.MessageHeader{
		Version:           hex.EncodeToString(append(data[:0:0], data[0:2]...)),
		LinkMessageLength: hex.EncodeToString(append(data[:0:0], data[2:4]...)),
		LinkMsgFrameSN:    hex.EncodeToString(append(data[:0:0], data[4:8]...)),
		LinkMsgType:       hex.EncodeToString(append(data[:0:0], data[8:10]...)),
		OpType:            hex.EncodeToString(append(data[:0:0], data[10:12]...)),
	}
}

func ParseMessageAppHeader(data []byte, offset int) (packets.MessageAppHeader, int) {
	return packets.MessageAppHeader{
		TotalLen:   hex.EncodeToString(append(data[:0:0], data[offset:offset+2]...)),
		SN:         hex.EncodeToString(append(data[:0:0], data[offset+2:offset+4]...)),
		CtrlField:  hex.EncodeToString(append(data[:0:0], data[offset+4:offset+6]...)),
		FragOffset: hex.EncodeToString(append(data[:0:0], data[offset+6:offset+8]...)),
		Type:       hex.EncodeToString(append(data[:0:0], data[offset+8:offset+10]...)),
		OpType:     hex.EncodeToString(append(data[:0:0], data[offset+10:offset+12]...)),
	}, offset + 12
}

func ParseMessageBody(data []byte, msgType string) (packets.MessageBody, int, error) {
	msgBody, offset := packets.MessageBody{}, 0
	var err error
	switch msgType {
	case packets.IBeaconMsg:
		msgBody.GwMac = hex.EncodeToString(append(data[:0:0], data[10:18]...))
		msgBody.ModuleID = hex.EncodeToString(append(data[:0:0], data[18:20]...))
		offset = 20
	default:
		err = errors.New("ParseMessageBody: from: " + msgBody.GwMac + " unable to recognize message type with " + msgType)
	}
	return msgBody, offset, err
}

func ParseMessageAppBody(data []byte, offset int, devMac string) (packets.MessageAppBody, error) {
	var err error
	parseMessageAppBody := packets.MessageAppBody{
		ErrorCode:    hex.EncodeToString(append(data[:0:0], data[offset:offset+2]...)),
		RespondFrame: hex.EncodeToString(append(data[:0:0], data[offset:offset+6]...)),
		TLV: packets.TLV{
			TLVMsgType: hex.EncodeToString(append(data[:0:0], data[offset+6:offset+8]...)),
			TLVLen:     hex.EncodeToString(append(data[:0:0], data[offset+8:offset+10]...)),
		},
	}
	switch parseMessageAppBody.TLV.TLVMsgType {
	case packets.ScanRespMsg:
		parseMessageAppBody.TLV.TLVPayload = packets.TLVFeature{
			ErrorCode:    hex.EncodeToString(append(data[:0:0], data[offset+6:offset+8]...)),
			ScanStatus:   hex.EncodeToString(append(data[:0:0], data[offset+8:offset+9]...)),
			ReserveOne:   hex.EncodeToString(append(data[:0:0], data[offset+9:offset+10]...)),
			ScanType:     hex.EncodeToString(append(data[:0:0], data[offset+10:offset+11]...)),
			ScanPhys:     hex.EncodeToString(append(data[:0:0], data[offset+11:offset+12]...)),
			ScanInterval: hex.EncodeToString(append(data[:0:0], data[offset+12:offset+14]...)),
			ScanWindow:   hex.EncodeToString(append(data[:0:0], data[offset+14:offset+16]...)),
			ScanTimeout:  hex.EncodeToString(append(data[:0:0], data[offset+16:offset+28]...)),
		}
	case packets.ConnectRespMsg:
		parseMessageAppBody.TLV.TLVPayload = packets.TLVFeature{
			DevMac:       hex.EncodeToString(append(data[:0:0], data[offset+6:offset+12]...)),
			ErrorCode:    hex.EncodeToString(append(data[:0:0], data[offset+12:offset+14]...)),
			ConnStatus:   hex.EncodeToString(append(data[:0:0], data[offset+14:offset+15]...)),
			PHY:          hex.EncodeToString(append(data[:0:0], data[offset+15:offset+16]...)),
			Handle:       hex.EncodeToString(append(data[:0:0], data[offset+16:offset+18]...)),
			ConnInterval: hex.EncodeToString(append(data[:0:0], data[offset+18:offset+20]...)),
			ConnLatency:  hex.EncodeToString(append(data[:0:0], data[offset+20:offset+22]...)),
			ScanTimeout:  hex.EncodeToString(append(data[:0:0], data[offset+22:offset+24]...)),
			MTUSize:      hex.EncodeToString(append(data[:0:0], data[offset+24:offset+26]...)),
		}
	default:
		err = errors.New("ParseMessageAppBody: from: " + devMac + " unable to recognize message type with " + parseMessageAppBody.TLV.TLVMsgType)
	}
	return parseMessageAppBody, err
}

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
			ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
	case globalconstants.CtrlLinkedMsgHeadWithBoy:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempStr = InsertString(tempStr,
			ConvertDecimalToHexStr(len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + enCodeResBody.String()
	case globalconstants.CtrlLinkedMsgWithMsgAppHeader:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String()
		tempAppStr = InsertString(tempAppStr,
			ConvertDecimalToHexStr(len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempStr = InsertString(tempStr,
			ConvertDecimalToHexStr(len(tempAppStr)+len(tempStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		encodeStr = tempStr + tempAppStr
	default:
		tempStr = enCodeResHeader.String() + enCodeResBody.String()
		tempAppStr = enCodeResAppMsg.String() + enCodeResAppMsgBody.String()
		tempAppTLVStr = encodeResAppMsgBodyTLV.String()
		tempAppTLVStr = InsertString(tempAppTLVStr,
			ConvertDecimalToHexStr(len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
			globalconstants.EncodeInsertIndex)
		tempAppStr = ConvertDecimalToHexStr(len(tempAppTLVStr)+len(tempAppStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen) + tempAppStr
		tempStr = InsertString(tempStr,
			ConvertDecimalToHexStr(len(tempStr)+len(tempAppStr)+len(tempAppTLVStr)+globalconstants.EncodeInsertLen, globalconstants.EncodeInsertLen),
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

// `param` 待生成数据 | `expectLen` 期望长 |
func ConvertDecimalToHexStr(param, expectLen int) string {
	hexStr := strconv.FormatInt(int64(param), 16)
	if len(hexStr) > expectLen {
		globallogger.Log.Errorln("<ConvertDecimalToHexStr>: please check param with |param| or |expectLen|")
		return ""
	}
	for len(hexStr) < expectLen {
		hexStr = "0" + hexStr
	}
	return hexStr
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
