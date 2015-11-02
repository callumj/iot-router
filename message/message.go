package message

type Message struct {
	RequestId     string
	RequestPath   string
	RequestParams map[string]string
}

type Response struct {
	RequestId string
	Data      string
}
