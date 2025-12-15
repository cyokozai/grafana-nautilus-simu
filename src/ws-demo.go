package main

import (
	"encoding/json"
	"time"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


const MQTTBroker string = "tcp://192.168.139.2:1883"
const Topic      string = "boids/v1.positions"


func main() {
	opts := mqtt.NewClientOptions().
		AddBroker(MQTTBroker).
		SetClientID(Topic)

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
		client.Publish(Topic, 0, false, b)
	}
}
