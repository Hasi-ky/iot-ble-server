package packets

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Descriptor struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type Descriptors struct {
	UUIDs []Descriptor
}

var (
	d    Descriptors
	dMap map[uint16]string
)

func SetDescriptors() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\descriptors.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &d)
	if err != nil {
		panic(err)
	}
	dMap = make(map[uint16]string)
	for _, v := range d.UUIDs {
		dMap[v.UUID] = v.Name
	}
}

func GetDescName(uuid uint16) string {
	if curDescriptor, ok := dMap[uuid]; ok {
		return curDescriptor
	}
	return "Unknown Descriptor"
}
