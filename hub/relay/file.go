package engine

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/three-plus-three/modules/hub"
	"github.com/three-plus-three/modules/tid"

	"github.com/three-plus-three/modules/hub/engine"
)

type FileRelay struct {
	core     *engine.Core
	basePath string
	Logger   *log.Logger
	stop     chan struct{}

	mu       sync.Mutex
	sendList map[string]struct{}
	recvList map[string]struct{}
}

func (fr *FileRelay) Run() error {
	for {
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-fr.stop:
			timer.Stop()
			return nil
		case <-timer.C:
		}

	}
}

func (fr *FileRelay) SubscribeTopic(name string) error {
	topic := fr.core.CreateTopicIfNotExists(name)
	consumer := topic.ListenOn()
	defer consumer.Close()

	basePath := filepath.Join(fr.basePath, "outgoing", "topic", name)
	return fr.subscribe(basePath, consumer)
}

func (fr *FileRelay) SubscribeQueue(name string) error {
	queue := fr.core.CreateQueueIfNotExists(name)
	consumer := queue.ListenOn()
	defer consumer.Close()

	basePath := filepath.Join(fr.basePath, "outgoing", "queue", name)
	return fr.subscribe(basePath, consumer)
}

func (fr *FileRelay) subscribe(basePath string, consumer *engine.Consumer) error {
	if err := os.MkdirAll(basePath, 0777); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	for msg := range consumer.C {
		err := ioutil.WriteFile(filepath.Join(basePath, tid.GenerateID()), msg.Data(), 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fr *FileRelay) ToTopic(name string) error {
	topic := fr.core.CreateTopicIfNotExists(name)
	basePath := filepath.Join(fr.basePath, "incoming", "topic", name)
	return fr.to(basePath, topic)
}

func (fr *FileRelay) ToQueue(name string) error {
	queue := fr.core.CreateQueueIfNotExists(name)
	basePath := filepath.Join(fr.basePath, "incoming", "queue", name)
	return fr.to(basePath, queue)
}

func (fr *FileRelay) to(basePath string, producer engine.Producer) error {
	for {
		fis, err := ioutil.ReadDir(basePath)
		if err != nil {
			if os.IsNotExist(err) {
				timer := time.NewTimer(1 * time.Second)
				select {
				case <-fr.stop:
					timer.Stop()
					return nil
				case <-timer.C:
				}
				continue
			}
			return err
		}

		for _, fi := range fis {
			filename := filepath.Join(basePath, fi.Name())
			bs, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			err = producer.Send(hub.CreateDataMessage(bs))
			if err != nil {
				return err
			}

			if err := os.Remove(filename); err != nil {
				return err
			}
		}
	}
	return nil
}
