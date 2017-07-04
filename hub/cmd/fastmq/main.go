package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/runner-mei/command"
	"github.com/three-plus-three/modules/hub"
	"github.com/three-plus-three/modules/hub/engine"
)

func main() {
	command.ParseAndRun()
}

type runCmd struct {
	listenAt string
}

func (cmd *runCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.listenAt, "listen_at", ":59876", "")
	return fs
}

func (cmd *runCmd) Run(args []string) error {
	opt := &engine.Options{}

	srv, err := engine.NewEngine(opt, nil)
	if err != nil {
		return err
	}

	fmt.Println("listen at -", cmd.listenAt)
	return http.ListenAndServe(cmd.listenAt, srv)
}

type sendCmd struct {
	url    string
	typ    string
	id     string
	repeat uint
	stat   bool
}

func (cmd *sendCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.url, "url", "http://127.0.0.1:59876", "")
	fs.StringVar(&cmd.typ, "type", hub.QUEUE, "send to '"+hub.TOPIC+"' or '"+hub.QUEUE+"'.")
	fs.StringVar(&cmd.id, "id", "", "the name of client.")
	fs.UintVar(&cmd.repeat, "repeat", 1, "send message count.")
	fs.BoolVar(&cmd.stat, "stat", false, "stat message rate.")
	return fs
}

func (cmd *sendCmd) Run(args []string) error {
	if len(args) != 2 {
		return errors.New("arguments error!\r\n\tUsage: fastmq send queue name messagebody")
	}

	builder := hub.Connect(cmd.url).ID(cmd.id)

	var err error
	var cli *hub.Publisher
	switch cmd.typ {
	case "topic":
		cli, err = builder.ToTopic(args[0])
	case "queue":
		cli, err = builder.ToQueue(args[0])
	default:
		return errors.New("arguments error: type must is '" + hub.TOPIC + "' or '" + hub.QUEUE + "'.")
	}

	if nil != err {
		return err
	}
	defer cli.Close()

	if cmd.repeat == 0 {
		cmd.repeat = 1
	}

	if cmd.stat {
		cli.Send([]byte("begin"))

		for i := uint(0); i < cmd.repeat; i++ {
			cli.Send([]byte(args[1] + strconv.FormatUint(uint64(i), 10)))
		}

		cli.Send([]byte("end"))
	} else {
		msg := []byte(args[1])
		for i := uint(0); i < cmd.repeat; i++ {
			cli.Send(msg)
		}
	}
	return nil
}

type subscribeCmd struct {
	url     string
	typ     string
	id      string
	forward string
	console bool
	stat    bool
	//repeat  uint
}

func (cmd *subscribeCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.url, "url", "http://127.0.0.1:59876", "the address of target mq server.")
	fs.StringVar(&cmd.typ, "type", hub.QUEUE, "send to '"+hub.TOPIC+"' or '"+hub.QUEUE+"'.")
	fs.StringVar(&cmd.id, "id", "", "the name of client.")
	fs.StringVar(&cmd.forward, "forward", "", "resend to address.")
	fs.BoolVar(&cmd.console, "console", true, "print message to console.")
	fs.BoolVar(&cmd.stat, "stat", false, "stat message rate.")
	//fs.UintVar(&self.repeat, "repeat", 1, "send message count.")
	return fs
}

func (cmd *subscribeCmd) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("arguments error!\r\n\tUsage: fastmq subscribe queue name")
	}
	var forwarder *hub.Publisher
	var subscription *hub.Subscription
	var err error

	if cmd.forward != "" {
		forwardBuilder := hub.Connect(cmd.url)
		switch cmd.typ {
		case "topic":
			forwarder, err = forwardBuilder.ToTopic(cmd.forward)
		case "queue":
			forwarder, err = forwardBuilder.ToQueue(cmd.forward)
		default:
			return errors.New("arguments error: type must is '" + hub.TOPIC + "' or '" + hub.QUEUE + "'.")
		}
		if err != nil {
			return err
		}
	}

	subBuilder := hub.Connect(cmd.url).ID(cmd.id)

	var startAt, endAt time.Time
	var messageCount uint = 0

	switch cmd.typ {
	case "topic":
		subscription, err = subBuilder.SubscribeTopic(args[0])
	case "queue":
		subscription, err = subBuilder.SubscribeQueue(args[0])
	default:
		return errors.New("arguments error: type must is '" + hub.TOPIC + "' or '" + hub.QUEUE + "'.")
	}
	if nil != err {
		return err
	}

	cb := func(sub *hub.Subscription, msg []byte) {
		if cmd.console {
			fmt.Println(string(msg))
		}

		if forwarder != nil {
			if err := forwarder.Send(msg); err != nil {
				log.Fatalln(err)
				return
			}
		}

		if cmd.stat {
			if bytes.Equal(msg, []byte("begin")) {
				//fmt.Println("recv:", message_count, ", elapsed:", time.Now().Sub(start_at))

				startAt = time.Now()
				messageCount = 0
			} else if bytes.Equal(msg, []byte("end")) {
				endAt = time.Now()
				fmt.Println("recv:", messageCount, ", elapsed:", endAt.Sub(startAt))
			} else {
				messageCount++
			}
		}
	}

	return subscription.Run(cb)
}

func init() {
	command.On("run", "run as mq server", &runCmd{}, nil)
	command.On("send", "send messages to mq server", &sendCmd{}, nil)
	command.On("subscribe", "subscribe messages from mq server", &subscribeCmd{}, nil)
}
