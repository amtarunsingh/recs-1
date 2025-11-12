package message

import (
	"encoding/json"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
)

type Envelope[T messaging.Message] struct {
	Name    string `json:"name"`
	Message T      `json:"message"`
}

func MarshalMessage[T messaging.Message](name string, message T) (messaging.Payload, error) {
	env := Envelope[T]{
		Name:    name,
		Message: message,
	}
	return json.Marshal(env)
}

func UnmarshalMessage[T messaging.Message](
	p messaging.Payload,
	expectName string,
) (T, error) {
	var env Envelope[T]
	var zero T

	if err := json.Unmarshal(p, &env); err != nil {
		return zero, err
	}

	gotName := env.Name
	if gotName != expectName {
		return zero, fmt.Errorf("wrong message name: have %q, want %q", gotName, expectName)
	}

	return env.Message, nil
}
