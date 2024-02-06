package bleudp

import (
	"context"
	"iot-ble-server/dgram"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalsocket"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
)

// Start ble udp server
func Start(ctx context.Context) error {
	globallogger.Log.Infoln("Start ble udp server")
	globalsocket.ServiceSocket = dgram.CreateUDPSocket(config.C.General.BindPort)
	go func() {
		defer globalsocket.ServiceSocket.Close()
		data := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, rinfo, err := globalsocket.ServiceSocket.Receive(data)
				if err != nil {
					globallogger.Log.Errorln(err)
					return
				}
				go UDPMsgProc(ctx, msg, rinfo)
			}
		}
	}()
	return nil
}

func UDPMsgProc(ctx context.Context, msg []byte, rinfo dgram.RInfo) {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorf("<UDPMsgProc> err :\n", err)
		}
	}()
	if checkMsgSafe(msg) {
		globallogger.Log.Warnln("<UDPMsgProc>: start parsing UDP packets ... from: IP", rinfo.Address, "the port is:", rinfo.Port)
		procMessage(ctx, msg, rinfo)
	}
}

func procMessage(ctx context.Context, data []byte, rinfo dgram.RInfo) {
	jsonInfo, err := parseUdpMessage(data, rinfo)
	if err != nil {
		globallogger.Log.Errorln("<procMessage>: UDP packets from:", rinfo.Address, "with an error:", err)
		return
	}
	if jsonInfo.MessageHeader.Version != packets.Version3 {
		globallogger.Log.Errorln("<procMessage>: DevEui:", jsonInfo.MessageBody.GwMac, "UDP version is illegal!")
		return
	}
	globallogger.Log.Infoln("<procMessage>: DevEui:", jsonInfo.MessageBody.GwMac, "current proc msg is", jsonInfo.MessageHeader.LinkMsgType)
	switch jsonInfo.MessageHeader.LinkMsgType[0:2] {
	case packets.ChannelControl:
		procChannelMsg(ctx, jsonInfo, jsonInfo.MessageBody.GwMac)
	case packets.GatewayManager:
		procGatewayMsg(ctx, jsonInfo, jsonInfo.MessageBody.GwMac+jsonInfo.MessageBody.ModuleID)
	case packets.TerminalManager:
		procTerminalMsg(ctx, jsonInfo, jsonInfo.MessageAppBody.TLV.TLVPayload.DevMac)
	default:
		globallogger.Log.Errorln("procMessage: DevEui:", jsonInfo.MessageBody.GwMac, "received unrecognized link message type!")
	}
}

func procChannelMsg(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procChannelMsg> : channel %s start proc msg\n", devEui)
	switch jsonInfo.MessageHeader.LinkMsgType {
	case packets.Hello:
		procHelloAck(ctx, jsonInfo, devEui)
	default:
		globallogger.Log.Errorf("<procChannelMsg>: DevEui:%s received unrecognized link message type:%sn", devEui, jsonInfo.MessageHeader.LinkMsgType)
	}
}

func procGatewayMsg(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procGatewayMsg> : gateway %s start proc msg\n", devEui)
	switch jsonInfo.MessageHeader.LinkMsgType {
	case packets.IotModuleStatusChange:
		procIotModuleStatus(ctx, jsonInfo, devEui)
	default:
		globallogger.Log.Errorf("<procGatewayMsg>: DevEui:%s received unrecognized message type:%s\n", devEui, jsonInfo.MessageHeader.LinkMsgType)
	}
}

func procTerminalMsg(ctx context.Context, jsonInfo packets.JsonUdpInfo, devEui string) {
	globallogger.Log.Infof("<procGatewayMsg> : terminal %s start proc msg\n", devEui)
	switch jsonInfo.MessageHeader.LinkMsgType {
	case packets.BleResponse:
		procBleResponse(ctx, jsonInfo, devEui)
	case packets.BleBoardcast:
		procBleBoardCast(ctx, jsonInfo, devEui)
	default:
		globallogger.Log.Errorf("<procTerminalMsg>: DevEui:%s received unrecognized  message type%s:\n", devEui, jsonInfo.MessageHeader.LinkMsgType)
	}
}
