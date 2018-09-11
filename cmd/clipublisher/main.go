package main

import (
	"context"
	"flag"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/fresh8/domsub/logging"
)

func PublishMain(projectID, topicName, message string) error {
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

	flag.StringVar(&message, "message", "hola señor!", "message to publish into the topic")
	flag.StringVar(&topicName, "topic", "domsub", "topic to publish messages to")
	flag.Parse()

	logging.Init()

	err := PublishMain("fresh-8-staging", topicName, message)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
