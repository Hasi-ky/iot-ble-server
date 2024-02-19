package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/internal/config"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

//创建全局mqtt publish消息处理 handler
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	globallogger.Log.Warnf("Sub Client Unknow Topic: %s, msg : %s", msg.Topic(), msg.Payload())
}

var ClientOptions *mqtt.ClientOptions
var client mqtt.Client
var taskID = "h3c-iot-ble-server"
var connectting = false
var connectFlag = false

func connect(clientOptions *mqtt.ClientOptions) error {
	//创建客户端连接
	client = mqtt.NewClient(clientOptions)
	//客户端连接判断
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		globallogger.Log.Errorf("[MQTT] mqtt connect error, taskID: %s, error: %s", taskID, token.Error())
		return token.Error()
	} else {
		globallogger.Log.Errorf("[MQTT] connect success taskID:%s, token:%+v, clientOptions:%+v", taskID, token, clientOptions)
		return nil
	}
}
func connectLoop(clientOptions *mqtt.ClientOptions) {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorln("MQTT connectLoop err :", err)
		}
	}()
	for {
		if err := connect(clientOptions); err != nil {
			client.Disconnect(250)
			time.Sleep(time.Second * 2)
		} else {
			connectFlag = true
			connectting = false
			break
		}
	}
}

// ConnectMQTT ConnectMQTT
func ConnectMQTT(mqttHost string, mqttPort string, mqttUserName string, mqttPassword string) {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorln("ConnectMQTT err :", err)
		}
	}()
	//设置连接参数
	ClientOptions = mqtt.NewClientOptions().AddBroker("tcp://" + mqttHost + ":" + mqttPort).SetUsername(mqttUserName).SetPassword(mqttPassword)
	//设置客户端ID
	ClientOptions.SetClientID(fmt.Sprintf("%s-%d", taskID, rand.Int()))
	//设置handler
	ClientOptions.SetDefaultPublishHandler(messagePubHandler)
	//设置保活时长
	ClientOptions.SetKeepAlive(30 * time.Second)
	ClientOptions.SetAutoReconnect(true)
	ClientOptions.SetMaxReconnectInterval(time.Minute)
	connectting = true
	connectLoop(ClientOptions)
	go func() {
		for {
			if !connectting && !client.IsConnectionOpen() {
				connectting = true
				client.Disconnect(250)
				connectLoop(ClientOptions)
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

// Publish Publish
func Publish(topic string, msg string) {
	if client.IsConnectionOpen() {
		//发布消息
		token := client.Publish(topic, 0, false, msg)
		if !token.WaitTimeout(60 * time.Second) {
			globallogger.Log.Errorf("[MQTT] Publish timeout: topic: %s, mqttMsg %s", topic, msg)
			client.Disconnect(250)
			time.Sleep(time.Second)
			return
		}
		if token.Error() != nil {
			globallogger.Log.Errorf("[MQTT] Publish error: topic: %s, mqttMsg %s, error %s", topic, msg, token.Error().Error())
			client.Disconnect(250)
			time.Sleep(time.Second)
		} else {
			globallogger.Log.Warnf("[MQTT] Publish success: topic: %s, mqttMsg %s", topic, msg)
		}
	} else {
		globallogger.Log.Errorf("[MQTT] Publish client is not connect: topic: %s, mqttMsg %s", topic, msg)
	}
}

// Subscribe Subscribe
func Subscribe(topic string, cb func(topic string, msg []byte)) {
	globallogger.Log.Errorf("Sub Topic : %s start", topic)
	client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		globallogger.Log.Warnf("Sub Client Topic : %s, msg : %s", msg.Topic(), msg.Payload())
		go cb(msg.Topic(), msg.Payload())
	})
}

func GetConnectFlag() bool {
	return connectFlag
}

func SetConnectFlag(flag bool) {
	connectFlag = flag
}

func Start(ctx context.Context) error {
	globallogger.Log.Infoln("Start ble mqtt server")
	go ConnectMQTT(config.C.MQTTConfig.MqttHost, config.C.MQTTConfig.MqttPort, config.C.MQTTConfig.Username, config.C.MQTTConfig.Password)
	go func() {
		for {
			if GetConnectFlag() {
				globallogger.Log.Warnln("start mqtt subscribe")
				subscribeFromMQTT()
				SetConnectFlag(false)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func subscribeFromMQTT() {
	subTopics := make([]string, 0)
	subTopics = append(subTopics, globalconstants.TopicV3GatewayNetworkIn)
	subTopics = append(subTopics, globalconstants.TopicV3GatewayNetworkInAck)
	subTopics = append(subTopics, globalconstants.TopicV3GatewayRPC)
	subTopics = append(subTopics, globalconstants.TopicV3GatewayTelemetry)
	for _, topic := range subTopics {
		Subscribe(topic, func(topic string, msg []byte) {
			ProcSubMsg(topic, msg)
		})
	}
}

func ProcSubMsg(topic string, jsonMsg []byte) {
	defer func() {
		err := recover()
		if err != nil {
			globallogger.Log.Errorln("ProcSubMsg err :", err)
		}
	}()
	switch topic {
	case globalconstants.TopicV3GatewayRPC:
		var mqttMsg globalstruct.RPCIotware
		err := json.Unmarshal(jsonMsg, &mqttMsg)
		if err != nil {
			globallogger.Log.Errorln("[Subscribe]: JSON marshaling failed:", err)
			return
		}
		globallogger.Log.Warnf("[Subscribe]: mqtt msg subscribe: topic: %s, mqttMsg %+v", topic, mqttMsg)
		procRPCIotware(mqttMsg)
	case globalconstants.TopicV3GatewayNetworkInAck:
		var networkInAckMsg globalstruct.NetworkInAckIotware
		err := json.Unmarshal(jsonMsg, &networkInAckMsg)
		if err != nil {
			globallogger.Log.Errorln("[Subscribe]: JSON marshaling failed:", err)
			return
		}
		procAckIotware(networkInAckMsg)
	case globalconstants.TopicV3GatewayTelemetry:
		var deviceUpMsg globalstruct.TelemetryIotware
		err := json.Unmarshal(jsonMsg, &deviceUpMsg)
		if err != nil {
			globallogger.Log.Errorln("[Subscribe]: JSON marshaling failed:", err)
			return
		}
		globallogger.Log.Errorf("[Subscribe]: mqtt msg subscribe: topic: %s, deviceDeleteMsg %+v", topic, deviceUpMsg)
		procTelemetryIotware(deviceUpMsg)
	case globalconstants.TopicV3GatewayNetworkIn:
		var deviceJoin globalstruct.NetworkInIotware
		err := json.Unmarshal(jsonMsg, &deviceJoin)
		if err != nil {
			globallogger.Log.Errorln("[Subscribe]: JSON marshaling failed:", err)
			return
		}
		globallogger.Log.Errorf("[Subscribe]: mqtt msg subscribe: topic: %s, deviceUpdateMsg %+v", topic, deviceJoin)
		procNetworkInIotware(deviceJoin)
	default:
		globallogger.Log.Warnf("[Subscribe]: invalid topic: %s", topic)
	}

}

func procNetworkInIotware(mqttMsg globalstruct.NetworkInIotware) {

}

func procAckIotware(mqttMsg globalstruct.NetworkInAckIotware) {

}

func procTelemetryIotware(mqttMsg globalstruct.TelemetryIotware) {

}

func procRPCIotware(mqttMsg globalstruct.RPCIotware) {

}
