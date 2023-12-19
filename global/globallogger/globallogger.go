package globallogger

import (
	"iot-ble-server/internal/config"
	"iot-ble-server/logger"
	"log"
	"strconv"
)

var Log logger.Logger

func Init() {
	var err error
	if Log, err = logger.New(logger.Config{
		Path:  config.C.General.LogFile,
		Level: strconv.Itoa(config.C.General.LogLevel),
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
