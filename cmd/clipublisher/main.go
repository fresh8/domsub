package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/fresh8/domsub/logging"
)

func PublishMain(projectID, topicName, subName, message string) error {
	log.Println("publishing", message, "to", projectID, topicName)

	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	topic := pubsubClient.Topic(topicName)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !ok {
		topic, err = pubsubClient.CreateTopic(ctx, topicName)
		if err != nil {
			return err
		}

		cfg := pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 60 * time.Second,
		}

		pubsubClient.CreateSubscription(ctx, subName, cfg)
	}

	defer func(t *pubsub.Topic) {
		t.Stop()
	}(topic)

	result := topic.Publish(ctx, &pubsub.Message{Data: []byte(message)})

	id, err := result.Get(ctx)
	if err != nil {
		return err
	}

	log.Println("published message", id)
	return nil
}

func main() {
	var message string
	var topicName string
	var subName string

	flag.StringVar(&message, "message", "hola señor!", "message to publish into the topic")
	flag.StringVar(&subName, "sub", "scenario", "subscription name")
	flag.StringVar(&topicName, "topic", "domsub", "topic to publish messages to")
	flag.Parse()

	logging.Init()

	err := PublishMain(os.Getenv("GOOGLE_PROJECT_ID"), topicName, subName, message)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
