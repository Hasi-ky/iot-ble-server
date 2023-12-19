package bleudp

import "iot-ble-server/internal/packets"

func dispatchHelloAck(jsoninfo packets.JsonUdpInfo) {
	var newJsonInfo = packets.JsonUdpInfo{}
	var reMsgHeader = packets.MessageHeader {
		Version: jsoninfo.MessageHeader.Version,
		LinkMsgFrameSN: jsoninfo.MessageHeader.LinkMsgFrameSN,
		LinkMsgType: jsoninfo.MessageAppHeader.Type,
		OpType: packets.Response,
	}
	
}