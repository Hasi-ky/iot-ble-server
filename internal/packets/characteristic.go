package packets

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type characteristic struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type characteristics struct {
	UUIDs []characteristic
}

var c characteristics

func SetCharacteristics() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\characteristic_uuids.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		panic(err)
	}
}

func GetCharName(uuid uint16) string {
	for _, v := range c.UUIDs {
		if v.UUID == uuid {
			return v.Name
		}
	}

	return "Unknown Characteristic"
}
