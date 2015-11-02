package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/callumj/iot-router/message"
	"github.com/callumj/iot-router/mqtt"
)

var TWIML = `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Message>Hello! The time is {MSG}</Message>
</Response>
`

var TWIML_CALL = `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say>Hello! The time is {MSG}</Say>
</Response>
`

var f = func(client *MQTT.Client, msg MQTT.Message) {
	log.Printf("Message: %s", msg)
	var m message.Message
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Println(err)
		return
	}

	twiml := TWIML
	if strings.HasPrefix(m.RequestPath, "/twilio_call_callback") {
		twiml = TWIML_CALL
	} else if !strings.HasPrefix(m.RequestPath, "/twilio_sms_callback") {
		return
	}

	data := message.Response{
		RequestId: m.RequestId,
		Data:      strings.Replace(twiml, "{MSG}", time.Now().String(), 1),
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
