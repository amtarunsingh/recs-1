package amazon_sns

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"io"
	"os"
	"strings"
)

type SnsSubscriber struct {
	wrappedSubscriber *sns.Subscriber
	logger            platform.Logger
}

type TopicResolver struct {
	config config.Config
}

func (t TopicResolver) ResolveTopic(ctx context.Context, topic string) (snsTopic sns.TopicArn, err error) {
	return sns.TopicArn(fmt.Sprintf("arn:aws:sns:%s:%s:%s", t.config.Aws.Region, "000000000000", topic)), nil
}

func NewSnsSubscriber(
	config config.Config,
	logger platform.Logger,
) *SnsSubscriber {
	awsCfg := GetSnsAwsConfig(config, logger)

	snsCfg := sns.SubscriberConfig{
		AWSConfig: awsCfg,
		GenerateSqsQueueName: func(ctx context.Context, topicArn sns.TopicArn) (string, error) {
			isFIFO := strings.HasSuffix(string(topicArn), ".fifo")
			base := strings.TrimSuffix(string(topicArn), ".fifo")

			parts := strings.Split(base, ":")
			name := parts[len(parts)-1]
			name = name + "-queue"
			if isFIFO {
				name = name + ".fifo"
			}
			return name, nil
		},
		TopicResolver: TopicResolver{
			config: config,
		},
	}

	sqsCfg := sqs.SubscriberConfig{
		AWSConfig: awsCfg,
	}

	subscriber, err := sns.NewSubscriber(snsCfg, sqsCfg, watermill.NewCaptureLogger())
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	return &SnsSubscriber{
		wrappedSubscriber: subscriber,
		logger:            logger,
	}
}

func (p *SnsSubscriber) Subscribe(ctx context.Context, topic messaging.Topic) (<-chan messaging.BackMessage, error) {
	messages, err := p.wrappedSubscriber.Subscribe(ctx, string(topic))
	if err != nil {
		return nil, err
	}

	out := make(chan messaging.BackMessage)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-messages:
				if !ok {
					return
				}
				p.safeProcessSnsMessage(ctx, topic, m, out)
			}
		}
	}()

	return out, nil
}

func (p *SnsSubscriber) safeProcessSnsMessage(ctx context.Context, topic messaging.Topic, m *message.Message, out chan<- messaging.BackMessage) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error(fmt.Sprintf("Panic processing SNS message on topic %s (ID: %s): %v", topic, m.UUID, r))
		}
	}()

	p.logger.Debug(fmt.Sprintf("Topic `%s`: received SNS message with ID `%s`", topic, m.UUID))
	bm := newSnsBackMessage(m)
	select {
	case out <- bm:
	case <-ctx.Done():
		return
	}
}

func (p *SnsSubscriber) Close() error {
	if c, ok := any(p.wrappedSubscriber).(io.Closer); ok {
		return c.Close()
	}
	return nil
}

type SnsBackMessage struct {
	wrappedMessage *message.Message
}

func newSnsBackMessage(message *message.Message) *SnsBackMessage {
	return &SnsBackMessage{
		wrappedMessage: message,
	}
}

func (bm *SnsBackMessage) GetPayload() messaging.Payload {
	return messaging.Payload(bm.wrappedMessage.Payload)
}
func (bm *SnsBackMessage) Nack() bool {
	return bm.wrappedMessage.Nack()
}
func (bm *SnsBackMessage) Ack() bool {
	return bm.wrappedMessage.Ack()
}
