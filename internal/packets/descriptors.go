package packets

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Descriptor struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type Descriptors struct {
	UUIDs []Descriptor
}

var d Descriptors

func SetDescriptors() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\descriptors.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &d)
	if err != nil {
		panic(err)
	}
}

func GetDescName(uuid uint16) string {
	for _, v := range d.UUIDs {
		if v.UUID == uuid {
			return v.Name
		}
	}

	return "Unknown Descriptor"
}
