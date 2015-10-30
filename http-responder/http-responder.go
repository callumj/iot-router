package main

import (
	"encoding/json"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/callumj/iot-router/mqtt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var f = func(client *MQTT.Client, msg MQTT.Message) {
	log.Printf("Message: %s", msg)
	var m struct {
		RequestId string `"json:request_id"`
	}
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Println(err)
		return
	}

	top := "/http/response" + m.RequestId
	log.Printf("Delivering to %s", top)
	t := client.Publish(top, 0, false, "The time is"+time.Now().String())
	t.Wait()
	if err := t.Error(); err != nil {
		log.Println(err)
	}
}

func main() {
	c := mqtt.CreateClient()

	c.Subscribe("/http/hello/joe", 0, f)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	log.Println("awaiting signal")
	<-sigs
	log.Println("exiting")

	c.Disconnect(250)
}
