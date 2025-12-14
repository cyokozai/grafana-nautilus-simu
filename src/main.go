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

type Frame struct {
	Fields []Field        `json:"fields"`
	Values [][]interface{} `json:"values"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

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

type Payload struct {
	Timestamp int64  `json:"timestamp"`
	Boids     []Boid `json:"boids"`
}

const PopulationSize = 100
const Margin         = 0.1
const TurnFactor     = 0.001
const SpeedLimit     = 0.2

const StreamName = "boids.v1.positions"
var	grafanaURL   = os.Getenv("GRAFANA_URL")
var grafanaToken = os.Getenv("GRAFANA_TOKEN")
var httpClient   = &http.Client{Timeout: 5 * time.Second}

func main() {
	if grafanaURL == "" || grafanaToken == "" {
		log.Fatal("GRAFANA_URL or GRAFANA_TOKEN is not set")
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids  := make([]Boid, PopulationSize)

	for i := range boids {
		angle := random.Float64() * 2.0 * math.Pi
		speed := 0.005 + random.Float64()*0.005
		boids[i] = Boid{
			Time:	 time.Now().UnixMilli(),
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
			boids[i].Time = t.UnixMilli()
		}

		flame := buildFrame(t, boids)
		postFrame(flame)
	}
}


func buildFrame(t time.Time, boids []Boid) Frame {
	values := make([][]interface{}, 0, len(boids))
	for _, b := range boids {
		values = append(values, []interface{}{
			t.UnixMilli(),
			b.ID,
			b.X,
			b.Y,
			b.Angle,
		})
	}

	return Frame{
		Fields: []Field{
			{Name: "time", Type: "time"},
			{Name: "id", Type: "string"},
			{Name: "x", Type: "number"},
			{Name: "y", Type: "number"},
			{Name: "rotation", Type: "number"},
		},
		Values: values,
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


func postFrame(frame Frame) {
	jsonData, err := json.Marshal(frame)
	if err != nil {
		log.Println("Error marshaling JSON:", err)

		return
	}

	req, err := http.NewRequest(
		"POST",
		grafanaURL+"/api/live/push/boids.v1.positions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Println("Error creating request:", err)

		return
	}
	req.Header.Set("Authorization", "Bearer "+grafanaToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", "1")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)

		return
	}
	resp.Body.Close()
}

