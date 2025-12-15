package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Boid struct {
	Time  int64   `json:"time"`
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"rotation"`
	Speed float64 `json:"-"`
	Vx    float64 `json:"-"`
	Vy    float64 `json:"-"`
}


const PopulationSize = 20
const Margin         = 0.1
const TurnFactor     = 0.001
const SpeedLimit     = 0.2

var MQTTBroker = os.Getenv("MQTT_BROKER")
var MQTTTopic  = os.Getenv("MQTT_TOPIC")


func main() {
	if MQTTBroker == "" || MQTTTopic == "" {
		log.Fatal("MQTT_BROKER or MQTT_TOPIC is not set")
	}

	opts := mqtt.NewClientOptions().
		AddBroker(MQTTBroker).
		SetClientID("boids-simulation").
		SetAutoReconnect(true).
		SetCleanSession(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	defer client.Disconnect(250)

	log.Println("Connected to MQTT broker")

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids  := initBoids(random)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for t := range ticker.C {
		ts := t.UnixMilli()

		for i := range boids {
			UpdateBoid(&boids[i])
			boids[i].Time = ts

			line := boidToLineProtocol(&boids[i])
			client.Publish(MQTTTopic, 0, false, line)
		}
	}
}

// InfluxDB Line Protocol 変換
func boidToLineProtocol(b *Boid) string {
	return fmt.Sprintf(
		"boids,id=%s x=%f,y=%f,rotation=%f,vx=%f,vy=%f,speed=%f %d",
		b.ID,
		b.X,
		b.Y,
		b.Angle,
		b.Vx,
		b.Vy,
		b.Speed,
		b.Time*1_000_000, // ms → ns
	)
}

// Boidsの初期集団を生成
func initBoids(r *rand.Rand) []Boid {
	boids := make([]Boid, PopulationSize)
	now   := time.Now().UnixMilli()

	for i := range boids {
		angle := r.Float64() * 2 * math.Pi
		speed := 0.005 + r.Float64()*0.005

		boids[i] = Boid{
			Time:  now,
			ID:    fmt.Sprintf("boid-%03d", i),
			X:     r.Float64()*2.0 - 1.0,
			Y:     r.Float64()*2.0 - 1.0,
			Angle: angle * 180 / math.Pi,
			Speed: speed,
			Vx:    math.Cos(angle) * speed,
			Vy:    math.Sin(angle) * speed,
		}
	}

	return boids
}

// Boids Algorithm
func UpdateBoid(b *Boid) {
	b.X += b.Vx
	b.Y += b.Vy

	if b.X < -1.0+Margin {
		b.Vx += TurnFactor
	}
	if b.X > 1.0-Margin {
		b.Vx -= TurnFactor
	}
	if b.Y < -1.0+Margin {
		b.Vy += TurnFactor
	}
	if b.Y > 1.0-Margin {
		b.Vy -= TurnFactor
	}

	speed := math.Sqrt(b.Vx*b.Vx + b.Vy*b.Vy)
	if speed > SpeedLimit {
		ratio := SpeedLimit / speed
		b.Vx *= ratio
		b.Vy *= ratio
	}

	b.Angle = math.Atan2(b.Vy, b.Vx) * 180 / math.Pi
}
