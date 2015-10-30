package main

import (
	"encoding/json"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/callumj/iot-router/mqtt"
	"io"
	"log"
	"net/http"
	"time"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

var mux map[string]func(http.ResponseWriter, *http.Request)

var c *MQTT.Client

func main() {
	c = mqtt.CreateClient()

	server := http.Server{
		Addr:         ":8000",
		Handler:      &myHandler{},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.ListenAndServe()
}

type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}

	t := time.Now().Unix()
	data := struct {
		RequestId string `"json:request_id"`
	}{
		fmt.Sprintf("%d", t),
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Panic(err)
	}

	done := make(chan bool, 1)

	sub := "/http/response" + data.RequestId
	f := func(client *MQTT.Client, msg MQTT.Message) {
		log.Printf("Response from MQTT: %q", msg)
		io.WriteString(w, string(msg.Payload()))
		c.Unsubscribe(sub)
		done <- true
	}

	ref := c.Subscribe(sub, 0, f)
	ref.Wait()
	if ref.Error(); err != nil {
		log.Println(err)
		return
	}
	ch := "/http" + r.RequestURI
	log.Printf("Notifying of '%s', waiting on %s", ch, sub)
	tok := c.Publish(ch, 0, false, string(b))
	tok.Wait()
	if err := tok.Error(); err != nil {
		log.Println(err)
	}

	select {
	case <-done:
		log.Println("Done!")
	case <-time.After(5 * time.Second):
		c.Unsubscribe(sub)
		http.Error(w, "Handler timed out", 500)
	}
}
