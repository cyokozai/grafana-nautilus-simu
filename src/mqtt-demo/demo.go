package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


var MQTTBroker string = os.Getenv("MQTT_BROKER")
var MQTTTopic  string = os.Getenv("MQTT_TOPIC")
const ClientID string = "boids-simulator"


func main() {
	fmt.Println(MQTTBroker)
	opts := mqtt.NewClientOptions().
		AddBroker(MQTTBroker).
		SetClientID(MQTTTopic)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for t := range ticker.C {
		msg := map[string]interface{}{
			"time": t.UnixMilli(),
			"id":   "boid-001",
			"x":    0.5,
			"y":    -0.3,
			"rotation": 45.0,
		}

		b, _ := json.Marshal(msg)
		client.Publish(MQTTTopic, 0, false, b)
	}
}
