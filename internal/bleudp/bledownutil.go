package bleudp

import (
	"context"
	"errors"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/global/globaltransfersn"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
)

//生成设备连接信息，
//生成的状态是全部消息，all
func GenerateJsonInfo(ctx context.Context, terminalInfo globalstruct.TerminalInfo, ctrl int, linkMsgType, tlvMsgType string,
	devConn globalstruct.DevConnection) (jsonInfo packets.JsonUdpInfo, err error) {
	if ctrl&1 == 1 {
		jsonInfo.MessageHeader, err = GenerateMessageHeader(ctx, terminalInfo.GwMac, linkMsgType)
		if err != nil {
			return
		}
	}
	if (ctrl>>1)&1 == 1 {
		jsonInfo.MessageBody, err = GenerateMessageBody(linkMsgType, terminalInfo)
		if err != nil {
			return
		}
	}
	if (ctrl>>2)&1 == 1 {
		jsonInfo.MessageAppHeader, err = GenerateMessageAppHeader(ctx, linkMsgType, terminalInfo)
		if err != nil {
			return
		}
	}
	if (ctrl>>3)&1 == 1 {
		jsonInfo.MessageAppBody, err = GenerateMessageAppBody(linkMsgType, tlvMsgType, devConn)
		if err != nil {
			return
		}
	}
	jsonInfo.PendCtrl = ctrl
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
	resultJson.LinkMsgFrameSN = globalutils.ConvertDecimalToHexStr(int(curFrameNum), 8)
	return
}

//消息体消息生成
func GenerateMessageBody(msgType string, terminal globalstruct.TerminalInfo) (resultJson packets.MessageBody, err error) {
	switch msgType {
	case packets.BleRequest:
		resultJson.GwMac = terminal.GwMac
		resultJson.ModuleID = globalutils.ConvertDecimalToHexStr(int(terminal.IotModuleId), 4)
	default:
		return resultJson, errors.New("unrecognized LinkMessage")
	}
	return
}

//应用消息头生成
//缺少总长度
func GenerateMessageAppHeader(ctx context.Context, appMsgType string,
	terminal globalstruct.TerminalInfo) (resultJson packets.MessageAppHeader, err error) {
	var (
		curFrameNum    int64
		transferDevKey = globalconstants.BleRedisSNTransfer + terminal.TerminalMac
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
func GenerateMessageAppBody(appMsgType, TLVType string, args interface{}) (resultJson packets.MessageAppBody, err error) {
	switch appMsgType {
	case packets.BleRequest: //28字节
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
	default:
		return resultJson, errors.New("unrecognized BLEMessage")
	}
	return
}
