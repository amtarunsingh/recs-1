package amazon_sns

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"github.com/google/uuid"
	"os"
	"reflect"
)

type SnsPublisher struct {
	pub    *sns.Publisher
	logger platform.Logger
}

func NewSnsPublisher(config config.Config, logger platform.Logger) *SnsPublisher {
	pub, err := sns.NewPublisher(
		sns.PublisherConfig{
			AWSConfig: GetSnsAwsConfig(config, logger),
			TopicResolver: TopicResolver{
				config: config,
			},
		},
		watermill.NewCaptureLogger(),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	return &SnsPublisher{
		pub:    pub,
		logger: logger,
	}
}

func (p SnsPublisher) Publish(topic messaging.Topic, m messaging.Message) error {
	wm := watermillMessage.NewMessage(uuid.NewString(), watermillMessage.Payload(m.GetPayload()))

	if topic.IsFifo() {
		t := reflect.Indirect(reflect.ValueOf(m)).Type()
		wm.Metadata.Set(sns.MessageGroupIdMetadataField, t.String())
		wm.Metadata.Set(sns.MessageDeduplicationIdMetadataField, m.GetDeduplicationId())
	}

	err := p.pub.Publish(string(topic), wm)
	if err != nil {
		return err
	}
	p.logger.Debug(fmt.Sprintf("Topic `%s`: published new SNS message with ID `%s`", topic, wm.UUID))
	return nil
}
