package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)


type Boid struct {
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"rotation"`
	Speed float64 `json:"-"`
	Vx    float64 `json:"-"`
	Vy    float64 `json:"-"`
}

type Payload struct {
	Timestamp int64  `json:"timestamp"`
	Boids     []Boid `json:"boids"`
}

const PopulationSize = 100
const Margin         = 0.1
const TurnFactor     = 0.001
const SpeedLimit     = 0.2

var	grafanaURL   = os.Getenv("GRAFANA_URL")
var grafanaToken = os.Getenv("GRAFANA_TOKEN")


func main() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids := make([]Boid, PopulationSize)

	for i := range boids {
		angle := random.Float64() * 2.0 * math.Pi
		speed := 0.005 + random.Float64()*0.005
		boids[i] = Boid{
			ID:    fmt.Sprintf("boid-%03d", i),
			X:     random.Float64()*2.0 - 1.0,
			Y:     random.Float64()*2.0 - 1.0,
			Angle: angle * 180.0 / math.Pi,
			Speed: speed,
			Vx:    math.Cos(angle) * speed,
			Vy:    math.Sin(angle) * speed,
		}
	}

	log.Printf("Starting Simulation with %d boids...", PopulationSize)

	for t := range ticker.C {
		for i := range boids {
			UpdateBoid(&boids[i])
		}

		payloadData := Payload{
			Timestamp: t.UnixMilli(),
			Boids:     boids,
		}

		go postAnnotation(payloadData)
	}
}


func UpdateBoid(b *Boid) {
	b.X, b.Y = b.X+b.Vx, b.Y+b.Vy

	if b.X < -1.0 + Margin {
		b.Vx += TurnFactor
	}
	if b.X > 1.0 - Margin {
		b.Vx -= TurnFactor
	}
	if b.Y < -1.0 + Margin {
		b.Vy += TurnFactor
	}
	if b.Y > 1.0 - Margin {
		b.Vy -= TurnFactor
	}

	speed := math.Sqrt(b.Vx*b.Vx + b.Vy*b.Vy)
	if speed > SpeedLimit {
		ratio := SpeedLimit / speed
		b.Vx *= ratio
		b.Vy *= ratio
	}

	rad := math.Atan2(b.Vy, b.Vx)
	b.Angle = rad * 180 / math.Pi
}


func postAnnotation(p Payload) {
	if grafanaURL == "" {
		log.Println("GRAFANA_URL is not set")

		return
	}
	wrapper := map[string]interface{}{"data": p}
	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Error marshaling JSON:", err)

		return
	}

	req, err := http.NewRequest("POST", grafanaURL+"/api/live/push/nautilus_stream", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+grafanaToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)

		return
	}
	defer resp.Body.Close()
}
