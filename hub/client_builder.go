package hub

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/three-plus-three/modules/netutil"
	"github.com/three-plus-three/modules/websocket2"
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

func (builder *ClientBuilder) connect(uri string) (*websocket2.Conn, error) {
	origin := uri
	if strings.HasPrefix(uri, "http://") {
		uri = "ws://" + strings.TrimPrefix(uri, "http://")
	} else if strings.HasPrefix(uri, "https://") {
		uri = "wss://" + strings.TrimPrefix(uri, "https://")
	}

	config, err := websocket2.NewConfig(uri, origin)
	if err != nil {
		return nil, err
	}
	//config.Protocol = []string{protocol}

	var dialer net.Dialer
	dialer.KeepAlive = 30 * time.Second
	config.Dialer = (&netutil.HttpDialer{DialWithContext: dialer.DialContext}).Dial
	return websocket2.DialConfig(config)
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

type Publisher websocket2.Conn

func (pub *Publisher) Send(bs Message) error {
	return websocket2.Message.Send((*websocket2.Conn)(pub), bs.Bytes())
}

func (pub *Publisher) Close() error {
	return closeConn((*websocket2.Conn)(pub))
}

func closeConn(conn *websocket2.Conn) error {
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
