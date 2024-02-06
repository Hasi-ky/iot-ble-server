package packets

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type characteristic struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type characteristics struct {
	UUIDs []characteristic
}

var (
	c    characteristics
	cMap map[uint16]string
)

func SetCharacteristics() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\characteristic_uuids.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		panic(err)
	}
	cMap = make(map[uint16]string)
	for _, v := range c.UUIDs {
		cMap[v.UUID] = v.Name
	}
}

func GetCharName(uuid uint16) string {
	if curName, ok := cMap[uuid]; ok {
		return curName
	}
	return "Unknown Characteristic"
}
