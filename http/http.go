package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/dchest/uniuri"

	"github.com/callumj/iot-router/message"
	"github.com/callumj/iot-router/mqtt"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

var mux map[string]func(http.ResponseWriter, *http.Request)
var chanMap map[string]chan message.Response

var c *MQTT.Client

func main() {
	chanMap = make(map[string]chan message.Response)
	c = mqtt.CreateClient()

	ref := c.Subscribe("http/responses", 0, HandleMqttMessage)
	ref.Wait()
	if err := ref.Error(); err != nil {
		log.Println(err)
		return
	}

	server := http.Server{
		Addr:         ":8000",
		Handler:      &myHandler{},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.ListenAndServe()
}

type myHandler struct{}

func HandleMqttMessage(client *MQTT.Client, msg MQTT.Message) {
	var r message.Response
	if err := json.Unmarshal(msg.Payload(), &r); err != nil {
		log.Println(err)
		return
	}
	log.Printf("Incoming: %v", r)
	if ch, ok := chanMap[r.RequestId]; ok {
		ch <- r
	}
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}

	formVals := map[string]string{}
	if err := r.ParseForm(); err == nil {
		for k, v := range r.Form {
			if len(v) >= 0 {
				formVals[k] = v[0]
			}
		}
	}

	t := time.Now().UnixNano()
	data := message.Message{
		RequestId:     fmt.Sprintf("%d-%s", t, uniuri.NewLen(5)),
		RequestPath:   r.RequestURI,
		RequestParams: formVals,
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Panic(err)
	}

	chanMap[data.RequestId] = make(chan message.Response, 1)

	log.Printf("Notifying of '%s'", data.RequestId)
	tok := c.Publish("http/requests", 0, false, string(b))
	tok.Wait()
	if err := tok.Error(); err != nil {
		log.Println(err)
	}

	select {
	case resp := <-chanMap[data.RequestId]:
		io.WriteString(w, resp.Data)
		close(chanMap[data.RequestId])
		chanMap[data.RequestId] = nil
		log.Println("Done!")
	case <-time.After(5 * time.Second):
		http.Error(w, "Handler timed out", 500)
	}
}
