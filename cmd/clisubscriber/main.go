package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/fresh8/domsub/logging"
)

type msgchan chan *pubsub.Message

type waitTime struct {
	ack   int64
	start int64
	stop  int64
}

func (w waitTime) Ack() time.Duration   { return time.Duration(w.ack) * time.Second }
func (w waitTime) Start() time.Duration { return time.Duration(w.start) * time.Second }
func (w waitTime) Stop() time.Duration  { return time.Duration(w.stop) * time.Second }

func SubscribeMain(projectID, topicName, subName string, wait waitTime, dieHard bool) error {
	uid, err := uuid()
	if err != nil {
		return err
	}
	log.Printf("starting subscriber %v\n", uid)
	log.SetPrefix("[" + uid + "] ")

	log.Println("subscribing to", subName, projectID, topicName)

	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := pubsubClient.Subscription(subName)
	sub.ReceiveSettings.MaxExtension = 2 * time.Minute

	log.Printf("sleep before starting for %v\n", wait.Start())
	time.Sleep(wait.Start())
	cctx, cancel := context.WithCancel(ctx)

	ch := make(msgchan)
	var wg sync.WaitGroup
	wg.Add(2)
	go receiveMessages(cctx, sub, &wg, ch)
	go ackMessages(&wg, ch, false, wait.Ack())

	log.Printf("sleep before stop for %v\n", wait.Stop())
	time.Sleep(wait.Stop())
	if dieHard {
		log.Fatal("Yippee ki-yay")
	}

	// cancel causes Receive to exit once all received messages are Ack'ed or Nack'ed.
	log.Println("canceling context")
	cancel()
	wg.Wait()

	return nil
}

func ackMessages(wg *sync.WaitGroup, ch msgchan, isNack bool, waitToAck time.Duration) {
	log.Printf("sleep before ack for %v\n", waitToAck)
	time.Sleep(waitToAck)
	log.Println("entering ack loop")
	for {
		m, ok := <-ch
		if !ok {
			break
		}

		if isNack {
			log.Println("nacking", m.ID, string(m.Data))
			m.Nack()
			continue
		}

		log.Println("acking", m.ID, string(m.Data))
		m.Ack()

	}
	log.Println("left ack loop")
	wg.Done()
}

func receiveMessages(ctx context.Context, sub *pubsub.Subscription, wg *sync.WaitGroup, ch msgchan) {
	log.Println("entering receive loop")
	err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Println("received", m.ID, string(m.Data))
		ch <- m
	})
	log.Println("left receive loop")
	close(ch)

	if err != nil {
		log.Println(err)
	}

	wg.Done()
}

func main() {
	var dieHard bool
	var subName string
	var topicName string
	var waitTimes waitTime

	flag.BoolVar(&dieHard, "fail", false, "hard fail without a clean shutdown")
	flag.Int64Var(&waitTimes.ack, "ack", 0, "wait to ack in seconds")
	flag.Int64Var(&waitTimes.start, "start", 0, "wait to ack in seconds")
	flag.Int64Var(&waitTimes.stop, "stop", 0, "wait to ack in seconds")
	flag.StringVar(&subName, "sub", "scenario1", "subscription to read messages from")
	flag.StringVar(&topicName, "topic", "domsub", "topic to read messages from")
	flag.Parse()

	logging.Init()

	err := SubscribeMain("fresh-8-staging", topicName, subName, waitTimes, dieHard)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func uuid() (string, error) {
	var u [16]byte
	_, err := rand.Read(u[:])
	if err != nil {
		return "", err
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return hex.EncodeToString(u[:]), nil
}
