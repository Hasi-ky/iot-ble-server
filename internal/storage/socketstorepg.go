package storage

import (
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalstruct"
)

func FindSocketAndUpdatePG(pgInfo map[string]interface{}) error {
	_, err := DB().Exec(`INSERT INTO iot_ble_socketinfo (gwmac, family, ipaddr, ipport, updatetime)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (gwmac) DO UPDATE
	SET
		gwmac = EXCLUDED.gwmac,
		family = EXCLUDED.family,
		ipaddr = EXCLUDED.ipaddr,
		ipport = EXCLUDED.ipport,
		updatetime = EXCLUDED.updatetime
	RETURNING *`, pgInfo["gwmac"], pgInfo["family"], pgInfo["ipaddr"], pgInfo["ipport"], pgInfo["updatetime"])
	return err
}

func createTables() {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorln("<createTables>: postgres createModels ", err)
		}
	}()
	DB().Exec(`CREATE TABLE IF NOT EXISTS iot_ble_socketinfo (id SERIAL PRIMARY KEY, gwmac TEXT  NOT NULL UNIQUE, family TEXT, ipaddr TEXT, ipport INTEGER, updatetime TIMESTAMP)`)
	//db.Exec(`CREATE TABLE IF NOT EXISTS iot_ble_socketinfo (id SERIAL PRIMARY KEY, gwmac TEXT(12), family TEXT, ipaddr TEXT, ipport INTEGER, updatetime TIMESTAMP)`)
}

//`len(args) 1` gwmac | `2` gwmac module id |
func FindSocketByGwMac(gwMac string) (*globalstruct.SocketInfo, error) {
	var socketInfo *globalstruct.SocketInfo
	err := db.Select(&socketInfo, `SELECT * FROM iot_ble_socketinfo WHERE gwmac = `, gwMac)
	return socketInfo, err
}
