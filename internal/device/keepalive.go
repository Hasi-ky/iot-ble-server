package device

import (
	"iot-ble-server/global/globallogger"

)

func KeepAlive() error {
	globallogger.Log.Infoln("start device keep alive")

	return nil
}