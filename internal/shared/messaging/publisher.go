package messaging

import (
	"reflect"
)

//go:generate mockgen -destination=../../testlib/mocks/publisher_mock.go -package=mocks github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging Publisher

type Publisher interface {
	Publish(topic Topic, message Message) error
}

type Payload []byte

type Message interface {
	GetDeduplicationId() string
	GetPayload() Payload
	Load(payload Payload) error
}

func MessageFromPayload[T Message](payload Payload) (*T, error) {
	var t T

	rv := reflect.ValueOf(&t).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	err := t.Load(payload)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
