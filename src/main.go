package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"math"
	"math/rand/v2"

	"github.com/gorilla/websocket"
)


type Frame struct {
	Type string       `json:"type"`
	Data FramePayload `json:"data"`
}

type FramePayload struct {
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
const Margin         = 0.1
const TurnFactor     = 0.001
const SpeedLimit     = 0.2

var	GrafanaURL   = os.Getenv("GRAFANA_URL")
var GrafanaToken = os.Getenv("GRAFANA_TOKEN")


func main() {
	conn, err := connectGrafanaLive()
	if err != nil {
		log.Fatal("connect error:", err)
	}
	defer conn.Close()
	go startPing(conn)

	stream := "stream/boids.v1.positions"
	log.Println("Subscribed to", stream)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	random := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	boids  := initBoids(random)

	for t := range ticker.C {
		for i := range boids {
			UpdateBoid(&boids[i])
			boids[i].Time = t.UnixMilli()
		}

		frame := boidsToFramePayload(boids, t.UnixMilli())

		msg := map[string]interface{}{
			"type":    "publish",
			"channel": stream,
			"data":    frame,
		}

		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("write error:", err)

			break
		}
	}
}


func startPing(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		conn.WriteControl(
			websocket.PingMessage,
			[]byte{},
			time.Now().Add(5*time.Second),
		)
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
		log.Println("Error parsing Grafana URL:", err)

		return nil, err
	}

	u.Scheme = map[string]string{
		"http":  "ws",
		"https": "wss",
	}[u.Scheme]
	u.Path = "/api/live/ws/"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+GrafanaToken)
	header.Set("X-Grafana-Org-Id", "1")

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)

	return conn, err
}


func subscribe(conn *websocket.Conn, stream string) error {
	msg := map[string]interface{}{
		"type": "subscribe",
		"channel": stream,
	}

	return conn.WriteJSON(msg)
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


func boidsToFramePayload(boids []Boid, ts int64) Frame {
	values := make([][]interface{}, 0, len(boids))

	for _, b := range boids {
		values = append(values, []interface{}{
			ts, 
			b.ID,
			b.X,
			b.Y,
			b.Angle,
		})
	}

	return Frame{
		Type: "dataframe",
		Data: FramePayload{
			Fields: []FrameField{
				{Name: "time", Type: "time"},
				{Name: "id", Type: "string"},
				{Name: "x", Type: "number"},
				{Name: "y", Type: "number"},
				{Name: "rotation", Type: "number"},
			},
			Values: values,
		},
	}
}
