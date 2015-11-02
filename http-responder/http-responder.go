package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/callumj/iot-router/message"
	"github.com/callumj/iot-router/mqtt"
)

var f = func(client *MQTT.Client, msg MQTT.Message) {
	log.Printf("Message: %s", msg)
	var m message.Message
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Println(err)
		return
	}

	data := message.Response{
		RequestId: m.RequestId,
		Data:      "The time is" + time.Now().String(),
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Publishing: %v", data)
	t := client.Publish("http/responses", 0, false, string(b))
	t.Wait()
	if err := t.Error(); err != nil {
		log.Println(err)
	}
}

func main() {
	c := mqtt.CreateClient()

	c.Subscribe("http/requests", 0, f)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println("awaiting signal")
	<-sigs
	log.Println("exiting")

	c.Disconnect(250)
}
