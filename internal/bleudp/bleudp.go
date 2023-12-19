package bleudp

import (
	"iot-ble-server/dgram"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalsocket"
	"iot-ble-server/internal/config"
)

// Start ble udp server
func Start() error {
	globallogger.Log.Infoln("Start ble udp server")
	globalsocket.ServiceSocket = dgram.CreateUDPSocket(config.C.General.BindPort)
	var err error
	go func() {
		defer globalsocket.ServiceSocket.Close()
		data := make([]byte, 1024)
		for {
			msg, rinfo, err1 := globalsocket.ServiceSocket.Receive(data)
			if err != nil {
				err = err1
				return
			}
			go UDPMsgProc(msg, rinfo)
		}
	}()
	return err
}


func UDPMsgProc(msg []byte, rinfo dgram.RInfo) {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorln("UDPMsgProc err :", err)
		}
	}()
	if checkMsgSafe(msg) {
		globallogger.Log.Warnln("UDPMsgProc: start parsing UDP packets ... from: IP", rinfo.Address, "the port is:", rinfo.Port)
		procMessage(msg, rinfo)
	}
}
