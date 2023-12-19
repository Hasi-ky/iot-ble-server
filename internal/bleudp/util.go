package bleudp

import (
	"encoding/hex"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/internal/packets"
	"strconv"
)

func checkMsgSafe(data []byte) bool {
	messageLen, _ := strconv.ParseInt(hex.EncodeToString(append(data[:0:0], data[2:4]...)), 16, 0)
	return int(messageLen) == len(data)
}

//udp版本号可能会有所不同，所以此处错误捕捉
//此处截留网关信息
func procMessage(data []byte, rinfo dgram.RInfo) {
	jsonInfo, err := parseUdpMessage(data, rinfo)
	if err != nil {
		globallogger.Log.Errorln("procMessage: UDP packets from:", rinfo.Address, "with an error:", err)
		return
	}
	if jsonInfo.MessageHeader.Version != packets.Version3 {
		globallogger.Log.Errorln("procMessage: DevEui:", jsonInfo.MessageBody.GwMac, "UDP version is illegal!")
		return
	}
	
}

func parseUdpMessage(data []byte, rinfo dgram.RInfo) (packets.JsonUdpInfo, error) {
	jsonUdpInfo := packets.JsonUdpInfo{}
	messageHeader := packets.ParseMessageHeader(data)
	messageBody, offset, err := packets.ParseMessageBody(data, messageHeader.LinkMsgType)
	if err != nil {
		return packets.JsonUdpInfo{}, err
	}
	messageAppHeader, offset := packets.ParseMessageAppHeader(data, offset)
	messageAppBody := &packets.MessageAppBody{}
	errorCodeOut, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset:offset+2]...)), 16, 16)
	respFrame, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+2:offset+6]...)), 16, 32)
	messageAppBody.ErrorCode = packets.ErrorCode(errorCodeOut)
	messageAppBody.RespondFrame = uint32(respFrame)
	messageAppBody.TLV.TLVMsgType = hex.EncodeToString(append(data[:0:0], data[offset+6:offset+8]...))
	tlvLen, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+8:offset+10]...)), 16, 16)
	messageAppBody.TLV.TLVLen = uint16(tlvLen)
	err = packets.ParseMessageAppBody(data, offset+10, messageAppBody, messageBody.GwMac)
	if err != nil {
		return packets.JsonUdpInfo{}, err
	}
	jsonUdpInfo.MessageHeader = messageHeader
	jsonUdpInfo.MessageBody = messageBody
	jsonUdpInfo.MessageAppHeader = messageAppHeader
	jsonUdpInfo.MessageAppBody = *messageAppBody
	jsonUdpInfo.Rinfo = rinfo
	return jsonUdpInfo, nil
}
