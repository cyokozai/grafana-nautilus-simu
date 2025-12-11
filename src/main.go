package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)


type GrafanaAnnotation struct {
	Time    int64    `json:"time"`
	Tags    []string `json:"tags"`
	Text    string   `json:"text"`
	TimeEnd int64    `json:"timeEnd,omitempty"`
}

type Payload struct {
	Annotations []GrafanaAnnotation `json:"annotations"`
	Pi          float64						  `json:"pi"`
	X					  int64               `json:"x"`
	Y        	  int64               `json:"y"`
	Inside		 	bool                `json:"inside"`
}

const 