package bleudp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"iot-ble-server/internal/storage"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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
	case packets.BleBoardcast, packets.BleTerminalEvent: //附加TLV无处理
		parseMessageAppBody.TLV = GetDevResponseTLV(data, *offset)
	default:
		err = errors.New("ParseMessageAppBody: from: " + devMac + " unable to recognize message type with " + parseMessageAppBody.TLV.TLVMsgType)
	}
	return parseMessageAppBody, err
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

// `frame` 帧序列 | `reSendTime` 发送次数 | `devEui` 唯一标识 标识消息队列 设备mac or 网关mac + module id | `curQueue` 消息队列 |
func ResendMessage(ctx context.Context, frameSN, reSendTimes int, devMac, gwMac string, waitDown []byte) {
	if reSendTimes >= 3 {
		globallogger.Log.Errorf("<ResendMessage> : devEui: %s has issue with the current retransmission message device\n", devMac)
		return
	}
	timer := time.NewTimer((time.Second * time.Duration(config.C.General.KeepAliveTime)))
	defer timer.Stop()
	<-timer.C
	if config.C.General.UseRedis {
		byteNode, err := globalredis.RedisCache.LIndex(ctx, devMac, 0).Result()
		if err != nil {
			if err != redis.Nil {
				globallogger.Log.Errorf("<ResendMessage> :devEui:%s redis has error, %v\n", devMac, err)
			}
			return //空值就直接放弃
		}
		var curNode storage.NodeCache
		err = json.Unmarshal([]byte(byteNode), &curNode)
		if err != nil {
			globallogger.Log.Errorf("<ResendMessage> :devEui:%s decode has error, %v\n", devMac, err)
			return
		}
		globallogger.Log.Infof("<ResendMessage> : devEui: %s resend msg, the msg type is %s\n", devMac, curNode.MsgType)
		if frameSN == curNode.FrameSN {
			SendDownMessage(waitDown, devMac, gwMac)
			go ResendMessage(ctx, frameSN, reSendTimes+1, devMac, gwMac, waitDown)
		}
	} else {
		tempQueue, ok := globalmemo.MemoCacheDev.Get(devMac)
		if ok {
			curQueue := tempQueue.(*storage.CQueue)
			if curQueue.Len() > 0 && frameSN == curQueue.Peek().FrameSN {
				globallogger.Log.Warnf("<ResendMessage> : devEui: %s resend msg, the msg type is %s", devMac, curQueue.Peek().MsgType)
				SendDownMessage(waitDown, devMac, gwMac)
				go ResendMessage(ctx, frameSN, reSendTimes+1, devMac, gwMac, waitDown)
			}
		} //若不存在待下发的，就放弃重传操作
	}
}

//异步方法， 自旋等待,在下发消息前，需要从currentMap中取到对应消息队列
//redis就传key下来 curQueue
//方法仅针对终端，因为网关应当不是一发一收 针对插卡
func SendMsgBeforeDown(ctx context.Context, sendBytes []byte, frameSN int, devEui, gwMac, msgType string) {
	curMemo := storage.NodeCache{
		FrameSN:   frameSN,
		TimeStamp: time.Now(),
		MsgType:   msgType,
	}
	cache, err := json.Marshal(curMemo)
	if err != nil {
		globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s data generate failed %v\n", devEui, err)
		return
	}
	if config.C.General.UseRedis {
		_, errPre := globalredis.RedisCache.RPush(ctx, devEui, cache).Result()
		if errPre != nil {
			globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s redis can't work %v\n", devEui, errPre)
			return
		}
		for {
			queueByte, err := globalredis.RedisCache.LIndex(ctx, devEui, 0).Result()
			if err != nil {
				globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s redis has error %v, frameSN sendDown failed\n", devEui, err)
				return
			}
			var headNode storage.NodeCache
			err = json.Unmarshal([]byte(queueByte), &headNode)
			if err != nil {
				globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s data has error %v, frameSN sendDown failed\n", devEui, err)
				return
			}
			if headNode.FrameSN == frameSN {
				if globalutils.CompareTimeIsExpire(time.Now(), headNode.TimeStamp, globalconstants.LimitMessageTime) {
					globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s queue has timeOut, frameSN sendDown failed\n", devEui)
					globalredis.RedisCache.Del(ctx, devEui)
				} else {
					SendDownMessage(sendBytes, devEui, gwMac)
					go ResendMessage(ctx, frameSN, 0, devEui, gwMac, sendBytes)
				}
				break
			}
			globallogger.Log.Warnf("<SendMsgBeforeDown>: devEui: %s wait to send down message, current frame is %v\n", devEui, frameSN)
			time.Sleep(time.Microsecond * 500) //自旋等待下发
		}
	} else {
		var curQueue *storage.CQueue
		if tempQueue, ok := globalmemo.MemoCacheDev.Get(devEui); ok {
			curQueue = tempQueue.(*storage.CQueue)
		}
		curQueue.Enqueue(curMemo)
		for {
			curQueue, ok := globalmemo.MemoCacheDev.Get(devEui)
			if !ok { //缺少下发数据
				globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s memo has error, frameSN sendDown failed\n", devEui)
				return
			}
			if curQueue.(*storage.CQueue).Peek().FrameSN == frameSN {
				if globalutils.CompareTimeIsExpire(time.Now(), curQueue.(*storage.CQueue).Peek().TimeStamp, globalconstants.LimitMessageTime) {
					globallogger.Log.Errorf("<SendMsgBeforeDown>: devEui: %s queue has timeOu, frameSN sendDown failed\n", devEui)
					globalmemo.MemoCacheDev.Remove(devEui)
				} else {
					SendDownMessage(sendBytes, devEui, gwMac)
					go ResendMessage(ctx, frameSN, 0, devEui, gwMac, sendBytes)
				}
				break
			}
			globallogger.Log.Warnf("<SendMsgBeforeDown>: devEui: %s wait %d message to send down", devEui, frameSN)
			time.Sleep(time.Microsecond * 500)
		}
	}
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
			tempTLV.TLVPayload.UUID = hex.EncodeToString(append(data[:0:0], data[i+8:i+int(tempLen)]...))
		case packets.TLVCharacteristicMsg:
			tempTLV.TLVPayload.Properties = hex.EncodeToString(append(data[:0:0], data[i+4:i+5]...))
			tempTLV.TLVPayload.ReserveOne = hex.EncodeToString(append(data[:0:0], data[i+5:i+6]...))
			tempTLV.TLVPayload.ServiceHandle = hex.EncodeToString(append(data[:0:0], data[i+6:i+8]...))
			tempTLV.TLVPayload.CharHandle = hex.EncodeToString(append(data[:0:0], data[i+8:i+10]...))
			tempTLV.TLVPayload.UUID = hex.EncodeToString(append(data[:0:0], data[i+10:i+int(tempLen)]...))
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
		res.TLVPayload.Handle = hex.EncodeToString(append(data[:0:0], data[index+10:index+12]...))        //连接句柄
		res.TLVPayload.DisConnReason = hex.EncodeToString(append(data[:0:0], data[index+12:index+14]...)) //断开事件原因
	default:
		globallogger.Log.Errorf("<GetDevResponseTLV>: the corresponding message type cannot be recognized. please troubleshoot the error!")
		return packets.TLV{}
	}
	return
}

//处理上行消息时消息弹栈操作
//校验消息是否与当前消息对应
//正对消息、落后消息，超前不存在
// 网关不需要
// -1 出现错误 0 正对  1 落后 2超前不存在
func CheckUpMessageFrame(ctx context.Context, frameKey string, frameSN int) (resCode int) {
	resCode = globalconstants.TERMINAL_CORRECT
	var (
		frameInfo storage.NodeCache
		cacheStr  string
		cache     []byte
		queue     interface{}
		err       error
		ok        bool
	)
	if config.C.General.UseRedis {
		cacheStr, err = globalredis.RedisCache.LIndex(ctx, frameKey, 0).Result() //解读完就弹出
		if err != nil && err != redis.Nil {
			globallogger.Log.Errorf("<CheckUpMessageFrame> DevEui %s get redis error %v\n", frameKey, err)
			resCode = globalconstants.TERMINAL_EXCEPTION
			return
		}
		cache = []byte(cacheStr)
		err = json.Unmarshal([]byte(cache), &frameInfo)
		if err != nil {
			globallogger.Log.Errorf("<CheckUpMessageFrame> DevEui %s data change error %v\n", frameKey, err)
			resCode = globalconstants.TERMINAL_EXCEPTION
			return
		}
	} else {
		queue, ok = globalmemo.MemoCacheDev.Get(frameKey)
		if !ok {
			globallogger.Log.Errorf("<CheckUpMessageFrame> DevEui %s get memo cache error %v\n", frameKey, err)
			resCode = globalconstants.TERMINAL_EXCEPTION
			return
		}
		frameInfo = queue.(*storage.CQueue).Peek()
	}
	if frameInfo.FrameSN < frameSN { //缓存落后，当前数据超前
		globallogger.Log.Errorf("<CheckUpMessageFrame> DevEui %s cache is out\n", frameKey)
		resCode = globalconstants.TERMINAL_ADVANCE
	} else if frameInfo.FrameSN > frameSN {
		globallogger.Log.Errorf("<CheckUpMessageFrame> DevEui %s message is out\n", frameKey)
		resCode = globalconstants.TERMINAL_DELAY //终端消息来的滞后
	} else {
		if config.C.General.UseRedis {
			globalredis.RedisCache.LPop(ctx, frameKey)
		} else {
			queue.(*storage.CQueue).Dequeue()
		}
	}
	return
}

//通用处理应答For Ble
//bool表示是否继续执行
func DealWithResponseBle(ctx context.Context, linkSN, appSN, methodName, errorCode, devEui string) bool {
	globallogger.Log.Infof("<%s> DevEui %s start deal ble mainservice response message\n", methodName, devEui)
	if errorCode != packets.Success {
		globallogger.Log.Errorf("<%s> DevEui %s has an error %v\n", devEui, packets.GetResult(errorCode, packets.English).String())
		return false
	}
	curLinkSN, _ := strconv.ParseInt(linkSN, 16, 32)
	curAppSN, _ := strconv.ParseInt(appSN, 16, 16)
	verifyLinkFrame := CheckUpMessageFrame(ctx, devEui, int(curLinkSN))
	verifyAppFrame := CheckUpMessageFrame(ctx, devEui, int(curAppSN))
	if verifyLinkFrame != globalconstants.TERMINAL_CORRECT || verifyAppFrame != globalconstants.TERMINAL_CORRECT { //目前仅后面
		globallogger.Log.Errorf("<%s> DevEui %s has an frame is error verifyLinkFrame = %v and verifyAppFrame = %v\n", devEui, verifyLinkFrame, verifyAppFrame)
		return false
	}
	return true
}
