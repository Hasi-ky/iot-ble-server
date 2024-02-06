package packets

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type service struct {
	UUID uint16 `yaml:"uuid"`
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type services struct {
	UUIDs []service
}

var (
	sMap map[uint16]string
	s    services
)

func SetServices() {
	data, err := ioutil.ReadFile("D:\\Code\\iot-ble-server\\packaging\\files\\service_uuids.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	sMap = make(map[uint16]string)
	for _, v := range s.UUIDs {
		sMap[v.UUID] = v.Name
	}

}

func GetSvcName(uuid uint16) string {
	if curSvName, ok := sMap[uuid]; ok {
		return curSvName
	}
	return "Unknown Service"
}
