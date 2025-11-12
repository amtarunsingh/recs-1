package app

import (
	"context"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"os"
	"sync"
)

type MessageProcessor struct {
	topicListener *TopicListener
	logger        platform.Logger
}

func NewMessageProcessor(
	topicListener *TopicListener,
	logger platform.Logger,
) *MessageProcessor {

	return &MessageProcessor{
		topicListener: topicListener,
		logger:        logger,
	}
}

func (s *MessageProcessor) Start(ctx context.Context) {
	wg := sync.WaitGroup{}
	for _, h := range []messaging.Topic{
		operation.DeleteRomancesTopic,
		operation.DeleteRomancesGroupTopic,
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.topicListener.Listen(ctx, h)
			if err != nil {
				s.logger.Error(err.Error())
				os.Exit(1)
			}
		}()
	}

	wg.Wait()
}

type TopicListener struct {
	subscriber   messaging.Subscriber
	topicHandler *messaging.TopicHandler
	logger       platform.Logger
}

func NewTopicListener(
	subscriber messaging.Subscriber,
	topicHandler *messaging.TopicHandler,
	logger platform.Logger,
) *TopicListener {
	return &TopicListener{
		subscriber:   subscriber,
		topicHandler: topicHandler,
		logger:       logger,
	}
}

func (t TopicListener) Listen(
	ctx context.Context,
	topic messaging.Topic,
) error {
	t.logger.Debug(
		fmt.Sprintf(
			"Starting listening to topic `%v` with handlers: %v",
			topic,
			t.topicHandler.GetRegisteredHandlers(topic),
		),
	)

	messages, err := t.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	defer func() { _ = t.subscriber.Close() }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case m, ok := <-messages:
			if !ok {
				return nil
			}
			t.safeProcessMessage(ctx, topic, m)
		}
	}
}

func (t TopicListener) safeProcessMessage(ctx context.Context, topic messaging.Topic, m messaging.BackMessage) {
	defer func() {
		if r := recover(); r != nil {
			t.logger.Error(fmt.Sprintf("Panic processing message on topic %s: %v", topic, r))
			m.Nack()
		}
	}()

	err := t.topicHandler.Dispatch(ctx, topic, m)
	if err != nil {
		t.logger.Error(err.Error())
		m.Nack()
		return
	}
	m.Ack()
}
