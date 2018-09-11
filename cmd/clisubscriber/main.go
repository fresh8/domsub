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

func SubscribeMain(projectID, topicName, subName string, waitToStart, waitToAck, waitToStop time.Duration, dieHard bool) error {
	uid, err := uuid()
	if err != nil {
		return err
	}
	log.SetPrefix("[" + uid + "] ")

	log.Println("subscribing to", subName, projectID, topicName)

	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := pubsubClient.Subscription(subName)

	ch := make(msgchan)
	var wg sync.WaitGroup

	time.Sleep(waitToStart)
	cctx, cancel := context.WithCancel(ctx)
	wg.Add(1)
	go receiveMessages(cctx, sub, &wg, ch)
	wg.Add(1)
	go ackMessages(&wg, ch, false, waitToAck)

	time.Sleep(waitToStop)
	// cancel causes Receive to exit once all received messages are Ack'ed or Nack'ed.
	log.Println("canceling context")
	cancel()
	wg.Wait()

	return nil
}

func ackMessages(wg *sync.WaitGroup, ch msgchan, isNack bool, waitToAck time.Duration) {
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
	var subName string
	var topicName string
	var hardFail bool

	flag.StringVar(&subName, "sub", "scenario1", "subscription to read messages from")
	flag.StringVar(&topicName, "topic", "domsub", "topic to read messages from")
	flag.BoolVar(&hardFail, "fail", false, "hard fail without a clean shutdown")
	flag.Parse()

	logging.Init()

	err := SubscribeMain("fresh-8-staging", topicName, subName, 1*time.Second, 30*time.Second, 30*time.Second, false)
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
