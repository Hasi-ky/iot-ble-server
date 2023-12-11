package packets

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type service struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type services struct {
	UUIDs []service
}

var s services

func SetServices() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\service_uuids.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
}

func GetSvcName(uuid uint16) string {
	for _, v := range s.UUIDs {
		if v.UUID == uuid {
			return v.Name
		}
	}

	return "Unknown Service"
}
