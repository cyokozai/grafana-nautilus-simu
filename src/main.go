package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)


type DataFrame struct {
	Fields []FrameField    `json:"fields"`
	Values [][]interface{} `json:"values"`
}

type FrameField struct {
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

const PopulationSize = 20
const Margin = 0.1
const TurnFactor = 0.001
const SpeedLimit = 0.2

var GrafanaURL = os.Getenv("GRAFANA_URL")
var GrafanaToken = os.Getenv("GRAFANA_TOKEN")

const Stream = "stream/boids.v1.positions"

func main() {
	if GrafanaURL == "" || GrafanaToken == "" {
		log.Fatal("GRAFANA_URL and GRAFANA_TOKEN must be set")
	}

	conn, err := connectGrafanaLive()
	if err != nil {
		log.Fatal("connect error:", err)
	}
	defer conn.Close()

	go readLoop(conn)
	go startPing(conn)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids := initBoids(random)

	log.Println("Starting simulation loop...")

	for t := range ticker.C {
		now := t.UnixMilli()
		for i := range boids {
			UpdateBoid(&boids[i])
			boids[i].Time = now
		}

		frame := boidsToFramePayload(boids, now)
		
		msg := map[string]interface{}{
			"action":  "publish",
			"channel": Stream,
			"data":    frame,
		}

		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("write error (connection lost?):", err)
			
			break
		}
	}
}

func readLoop(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}
		
		log.Printf("recv: %s", message)
	}
}

func startPing(conn *websocket.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := conn.WriteControl(
			websocket.PingMessage,
			[]byte{},
			time.Now().Add(5*time.Second),
		)
		if err != nil {
			log.Println("ping error:", err)
			return
		}
	}
}

func initBoids(r *rand.Rand) []Boid {
	boids := make([]Boid, PopulationSize)
	now := time.Now().UnixMilli()

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

func connectGrafanaLive() (*websocket.Conn, error) {
	u, err := url.Parse(GrafanaURL)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}

	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	u.Path = "/api/live/ws"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+GrafanaToken)

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	return conn, err
}

func UpdateBoid(b *Boid) {
	b.X, b.Y = b.X+b.Vx, b.Y+b.Vy

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

	rad := math.Atan2(b.Vy, b.Vx)
	b.Angle = rad * 180 / math.Pi
}

func boidsToFramePayload(boids []Boid, ts int64) DataFrame {
	count := len(boids)

	colTime := make([]interface{}, count)
	colID := make([]interface{}, count)
	colX := make([]interface{}, count)
	colY := make([]interface{}, count)
	colRot := make([]interface{}, count)

	for i, b := range boids {
		colTime[i] = ts
		colID[i] = b.ID
		colX[i] = b.X
		colY[i] = b.Y
		colRot[i] = b.Angle
	}

	return DataFrame{
		Fields: []FrameField{
			{Name: "time", Type: "time"},
			{Name: "id", Type: "string"},
			{Name: "x", Type: "number"},
			{Name: "y", Type: "number"},
			{Name: "rotation", Type: "number"},
		},
		Values: [][]interface{}{
			colTime,
			colID,
			colX,
			colY,
			colRot,
		},
	}
}