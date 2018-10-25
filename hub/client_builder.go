package hub

import (
	"net/url"
	"strings"

	"golang.org/x/net/websocket"
)

const (
	QUEUE = "queue"
	TOPIC = "topic"
)

type ClientBuilder struct {
	baseURL string
	//capacity int
	//bufSize  int
	id string
}

func (builder *ClientBuilder) Clone() *ClientBuilder {
	return &ClientBuilder{
		baseURL: builder.baseURL,
		//capacity: builder.capacity,
		//bufSize:  builder.bufSize,
		//id:       builder.id,
	}
}

func (builder *ClientBuilder) ID(name string) *ClientBuilder {
	builder.id = name
	return builder
}

/*
func (builder *ClientBuilder) SetBufSize(size int) *ClientBuilder {
	builder.bufSize = size
	return builder
}

func (builder *ClientBuilder) SetQueueCapacity(capacity int) *ClientBuilder {
	builder.capacity = capacity
	return builder
}
*/

func (builder *ClientBuilder) ToQueue(name string) (*Publisher, error) {
	u := joinURL(builder.baseURL, "/sendQueue?name="+url.QueryEscape(name)+
		"&client="+url.QueryEscape(builder.id))
	return builder.to(u)
}

func (builder *ClientBuilder) ToTopic(name string) (*Publisher, error) {
	u := joinURL(builder.baseURL, "/sendTopic?name="+url.QueryEscape(name)+
		"&client="+url.QueryEscape(builder.id))
	return builder.to(u)
}

func (builder *ClientBuilder) to(uri string) (*Publisher, error) {
	conn, err := builder.connect(uri)
	if err != nil {
		return nil, err
	}
	return (*Publisher)(conn), nil
}

func (builder *ClientBuilder) connect(uri string) (*websocket.Conn, error) {
	origin := uri
	if strings.HasPrefix(uri, "http://") {
		uri = "ws://" + strings.TrimPrefix(uri, "http://")
	} else if strings.HasPrefix(uri, "https://") {
		uri = "wss://" + strings.TrimPrefix(uri, "https://")
	}

	config, err := websocket.NewConfig(uri, origin)
	if err != nil {
		return nil, err
	}
	//config.Protocol = []string{protocol}
	return websocket.DialConfig(config)
}

func (builder *ClientBuilder) SubscribeQueue(name string) (*Subscription, error) {
	u := joinURL(builder.baseURL, "/subscribeQueue?name="+url.QueryEscape(name)+
		"&client="+url.QueryEscape(builder.id))
	return builder.subscribe(u)
}

func (builder *ClientBuilder) SubscribeTopic(name string) (*Subscription, error) {
	u := joinURL(builder.baseURL, "/subscribeTopic?name="+url.QueryEscape(name)+
		"&client="+url.QueryEscape(builder.id))
	return builder.subscribe(u)
}

func (builder *ClientBuilder) subscribe(uri string) (*Subscription, error) {
	conn, err := builder.connect(uri)
	if err != nil {
		return nil, err
	}
	return &Subscription{Conn: conn}, nil
}

func Connect(uri string) *ClientBuilder {
	return &ClientBuilder{baseURL: uri}
}

type Publisher websocket.Conn

func (pub *Publisher) Send(bs Message) error {
	return websocket.Message.Send((*websocket.Conn)(pub), bs.Bytes())
}

func (pub *Publisher) Close() error {
	return closeConn((*websocket.Conn)(pub))
}

func closeConn(conn *websocket.Conn) error {
	return conn.Close()
}

func joinURL(a, b string) string {
	if strings.HasSuffix(a, "/") {
		if strings.HasPrefix(b, "/") {
			return strings.TrimSuffix(a, "/") + b
		}
		return a + b
	}

	if strings.HasPrefix(b, "/") {
		return a + b
	}
	return a + "/" + b
}
