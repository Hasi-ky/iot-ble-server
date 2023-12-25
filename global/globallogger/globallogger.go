package globallogger

import (
	"iot-ble-server/internal/config"
	"iot-ble-server/logger"
	"log"
)

var Log logger.Logger

func Init() {
	var err error
	numberToLog := map[int]string{
		5 :"debug", 
		4 :"info", 
		3 :"warning", 
		2 :"error", 
		1:"fatal", 
		0:"panic", 
	}
	if Log, err = logger.New(logger.Config{
		Path:  config.C.General.LogFile,
		Level: numberToLog[config.C.General.LogLevel],
	}, "service", "iot-ble-server"); err != nil {
		log.Panic(err)
	}
}

func SetLogLevel(level string) {
	var err error
	if Log, err = logger.New(logger.Config{
		Level: level,
	}, "service", "iot-ble-server"); err != nil {
		log.Panic(err)
	}
}
