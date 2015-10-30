package main

import (
	"encoding/json"
	"github.com/callumj/iot-router/mqtt"
	"log"
	"os"
	"time"
)

func main() {
	c := mqtt.CreateClient()

	if os.Getenv("MESSAGE") != "" {
		data := struct {
			Message string `json:"message"`
		}{
			os.Getenv("MESSAGE"),
		}
		b, err := json.Marshal(data)
		if err != nil {
			log.Panic(err)
		}
		t := c.Publish("/go-mqtt/sample", 0, false, string(b))
		t.Wait()
		if err := t.Error(); err != nil {
			log.Panic(err)
		}
		return
	}

	c.Subscribe("/go-mqtt/sample", 0, nil)

	time.Sleep(1 * time.Hour)

	c.Disconnect(250)
}
