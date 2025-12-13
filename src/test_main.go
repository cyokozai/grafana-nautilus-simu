package test_main

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

	"testing"
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

var grafanaURL   string = os.Getenv("GRAFANA_URL")
var grafanaToken string = os.Getenv("GRAFANA_TOKEN")

const windowWidth  float64 = 896			/ 2.0
const windowHeight float64 = 597			/ 2.0

const populationSize int = 100
const margin 		 float64 = 0.1
const turnFactor float64 = 0.001
const speedLimit float64 = 0.2


func main() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids := make([]Boid, populationSize)
	for i := range boids {
		angle := random.Float64() * 2.0 * math.Pi
		speed := 0.005 + random.Float64() * 0.005
		boids[i] = Boid{
			ID:    "boid-" + fmt.Sprintf("%03d", i),
			X:     (random.Float64()*2.0 - 1.0) * windowWidth,
			Y:     (random.Float64()*2.0 - 1.0) * windowHeight,
			Angle: angle * 180.0 / math.Pi,
			Speed: speed,
			Vx:    math.Cos(angle) * speed,
			Vy:    math.Sin(angle) * speed,
		}
	}
	
	log.Printf("Starting Grafana Nautiluses Simulation with %d boids...", populationSize)

	for t := range ticker.C {
		for i := range boids {
			b := &boids[i]

			b.X, b.Y = b.X + b.Vx, b.Y + b.Vy

			if b.X < -1.0 * windowWidth + margin * windowWidth {
				b.Vx += turnFactor
			}
			if b.X > 1.0 * windowWidth - margin * windowWidth {
				b.Vx -= turnFactor
			}
			if b.Y < -1.0 * windowHeight + margin * windowHeight {
				b.Vy += turnFactor
			}
			if b.Y > 1.0 * windowHeight - margin * windowHeight {
				b.Vy -= turnFactor
			}

			speed := math.Sqrt(b.Vx * b.Vx + b.Vy * b.Vy)
			if speed > speedLimit {
				b.Vx, b.Vy = (b.Vx / speed) * speedLimit, (b.Vy / speed) * speedLimit
			}

			rad 		:= math.Atan2(b.Vy, b.Vx)
			degrees := rad * 180 / math.Pi
			b.Angle	 = degrees
		}

		payloadData := Payload{
			Timestamp: t.UnixMilli(),
			Boids:     boids,
		}

		go postAnnotation(payloadData)
	}
}


func postAnnotation(p Payload) {
	wrapper := map[string]interface{}{
		"data": p,
	}
	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
	}

	req, err := http.NewRequest(
		"POST",
		grafanaURL + "/api/live/push/nautilus_stream",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Println("Error creating request:", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + grafanaToken)

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
	}
	defer resp.Body.Close()
}
