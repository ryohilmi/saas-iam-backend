package providers

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"
)

type PubSub struct {
	client        *pubsub.Client
	context       context.Context
	subscriptions map[string]*pubsub.Subscription
}

type Callback func(ctx context.Context, payload map[string]interface{})

func NewPubSub(projectId string) (*PubSub, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}

	return &PubSub{
		client:        client,
		context:       ctx,
		subscriptions: make(map[string]*pubsub.Subscription),
	}, nil
}

func (p *PubSub) CloseConnection() {
	p.client.Close()
}

func (p *PubSub) Subscribe(subscriptionId string, callbacks []Callback) error {

	err := p.client.Subscription(subscriptionId).Receive(p.context, func(ctx context.Context, msg *pubsub.Message) {

		var payload map[string]interface{}
		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
		}

		for _, callback := range callbacks {
			callback(ctx, payload)
		}

		msg.Ack()
	})

	if err != nil {
		log.Fatalf("Failed to receive message: %v", err)
	}

	log.Printf("subscribed to %s", subscriptionId)

	return nil
}
