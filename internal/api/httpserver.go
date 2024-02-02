package api

import (
	"iot-ble-server/global/globallogger"
)

func Start()  error {
	globallogger.Log.Infoln("start http server")
	
	return nil
}