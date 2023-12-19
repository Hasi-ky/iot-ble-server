package packets

import (
	"encoding/hex"
	"errors"
	"strconv"
)

func ParseMessageHeader(data []byte) MessageHeader {
	return MessageHeader{
		Version:           hex.EncodeToString(append(data[:0:0], data[0:2]...)),
		LinkMessageLength: hex.EncodeToString(append(data[:0:0], data[2:4]...)),
		LinkMsgFrameSN:    hex.EncodeToString(append(data[:0:0], data[4:8]...)),
		LinkMsgType:       hex.EncodeToString(append(data[:0:0], data[8:10]...)),
		OpType:            OpType(hex.EncodeToString(append(data[:0:0], data[10:12]...))),
	}
}

func ParseMessageAppHeader(data []byte, offset int) (MessageAppHeader, int) {
	return MessageAppHeader{
		TotalLen:   hex.EncodeToString(append(data[:0:0], data[offset:offset+2]...)),
		SN:         hex.EncodeToString(append(data[:0:0], data[offset+2:offset+4]...)),
		CtrlField:  hex.EncodeToString(append(data[:0:0], data[offset+4:offset+6]...)),
		FragOffset: hex.EncodeToString(append(data[:0:0], data[offset+6:offset+8]...)),
		Type:       hex.EncodeToString(append(data[:0:0], data[offset+8:offset+10]...)),
		OpType:     OpType(hex.EncodeToString(append(data[:0:0], data[offset+10:offset+12]...))),
	}, offset + 12
}

func ParseMessageBody(data []byte, msgType string) (MessageBody, int, error) {
	msgBody, offset := MessageBody{}, 0
	var err error
	switch msgType {
	case IBeaconMsg:
		msgBody.GwMac = hex.EncodeToString(append(data[:0:0], data[10:18]...))
		msgBody.ModuleID = hex.EncodeToString(append(data[:0:0], data[18:20]...))
		offset = 20
	default:
		err = errors.New("ParseMessageBody: from: " + msgBody.GwMac + " unable to recognize message type with " + msgType)
	}
	return msgBody, offset, err
}

//一条消息仅附带单TLV, 在入参前就需要获取外部错误吗，响应帧序列号，TLV-type, length
func ParseMessageAppBody(data []byte, offset int, parseMessageAppBody *MessageAppBody, devMac string) error {
	var err error

	switch parseMessageAppBody.TLV.TLVMsgType {
	case ScanRespMsg:
		errCode, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+4:offset+6]...)), 16, 8)
		scanStatus, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+6:offset+7]...)), 16, 8)
		reserveOne, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+7:offset+8]...)), 16, 8)
		scanType, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+8:offset+9]...)), 16, 8)
		scanPhys, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+9:offset+10]...)), 16, 8)
		scanInterval, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+10:offset+12]...)), 16, 16)
		scanWindow, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+12:offset+14]...)), 16, 16)
		scanTimeout, _ := strconv.ParseUint(hex.EncodeToString(append(data[:0:0], data[offset+14:offset+16]...)), 16, 16)
		parseMessageAppBody.TLV.TLVPayload = TLVFeature{
			ErrorCode:    ErrorCode(errCode),
			ScanStatus:   uint8(scanStatus),
			ReserveOne:   uint8(reserveOne),
			ScanType:     uint8(scanType),
			ScanPhys:     uint8(scanPhys),
			ScanInterval: uint16(scanInterval),
			ScanWindow:   uint16(scanWindow),
			ScanTimeout:  uint16(scanTimeout),
		}
	case ConnectRespMsg:

	default:
		err = errors.New("ParseMessageAppBody: from: " + hex.EncodeToString(devMac) + " unable to recognize message type with " + parseMessageAppBody.TLV.TLVMsgType)
	}
	return err
}
