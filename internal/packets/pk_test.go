package packets

import (
	"fmt"
	"iot-ble-server/global/globalstruct"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)
//var m1 map[string]string
func init() {
	//m1 = map[string]string{}
}

func Test1(t *testing.T) {
	var m1 map[string][]int
	fmt.Println(m1["1"][2])
}

func TestXxx(t *testing.T) {
	db, err := sqlx.Open("postgres", "postgres://iotware:iotware@33.33.33.244:5432/iotware?sslmode=disable")
	if err != nil {
		panic(err)
	}

	//_, err := db.Exec(`CREATE TABLE IF NOT EXISTS iot_ble_socketinfo (id SERIAL PRIMARY KEY, gwmac char(12), family varchar(), ipaddr varchar(), ipport INT, updatetime TIMESTAMP)`)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS iot_ble_socketinfo (id SERIAL PRIMARY KEY, gwmac CHAR(12), family VARCHAR(50), ipaddr VARCHAR(50), ipport VARCHAR(50), updatetime TIMESTAMP)`)
	if err != nil {
		panic(err)
	}
	socketInfo := globalstruct.SocketInfo{
		ID:     2,
		Mac:    "aaaaaaaaaaaa",
		Family: "family",
		IPAddr: "10.29.02.11",
		IPPort: 65225,
	}
	_, err = db.NamedExec("INSERT INTO iot_ble_socketinfo (id, gwmac, family,ipaddr,ipport) VALUES (:id,:gwmac,:family,:ipaddr,:ipport)", socketInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("连接成功")
	db.Close()
}

func TestInsert(t *testing.T) {
	db, err := sqlx.Open("postgres", "postgres://iotware:iotware@33.33.33.244:5432/iotware?sslmode=disable")
	if err != nil {
		panic(err)
	}

	m1 := map[string]interface{}{
		"gwmac":      "aaaaaaaaaaaa",
		"family":     "family",
		"ipaddr":     "10.29.02.11",
		"ipport":     65225,
		"updatetime": time.Now(),
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS iot_ble_socketinfo (id SERIAL PRIMARY KEY, gwmac TEXT  NOT NULL UNIQUE, family TEXT, ipaddr TEXT, ipport INTEGER, updatetime TIMESTAMP)`)
	_, err = db.Exec(`INSERT INTO iot_ble_socketinfo (gwmac, family, ipaddr, ipport, updatetime)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (gwmac) DO UPDATE
	SET
		gwmac = EXCLUDED.gwmac,
		family = EXCLUDED.family,
		ipaddr = EXCLUDED.ipaddr,
		ipport = EXCLUDED.ipport,
		updatetime = EXCLUDED.updatetime
	RETURNING *`, m1["gwmac"], m1["family"], m1["ipaddr"], m1["ipport"], m1["updatetime"])
	fmt.Println(err)
}

func TestMain(t *testing.T) {
	var s []string
	fmt.Println(len(s))

}
