package messaging

import (
	"context"
	"errors"
	"fmt"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

type Handler[T Message] interface {
	GetName() string
	Handle(ctx context.Context, message T) error
}

type TopicHandler struct {
	handlers map[Topic]map[string]untypedHandler
	logger   platform.Logger
}

func NewTopicHandler(logger platform.Logger) *TopicHandler {
	return &TopicHandler{handlers: make(map[Topic]map[string]untypedHandler), logger: logger}
}

func RegisterTopicHandler[T Message](r *TopicHandler, topic Topic, h Handler[T]) {
	r.register(topic, handlerAdapter[T]{name: h.GetName(), h: h})
}

func (r *TopicHandler) Dispatch(ctx context.Context, topic Topic, msg BackMessage) error {
	hs := r.handlers[topic]

	if len(hs) == 0 {
		return fmt.Errorf("no handlers registered for topic %q", topic)
	}

	var errs []error
	for _, h := range hs {
		if !h.canHandle(msg) {
			r.logger.Info(fmt.Sprintf("Message `%s` not handled for topic %q", msg.GetPayload(), topic))
			continue
		}
		if err := h.handleUntyped(ctx, msg); err != nil {
			errs = append(errs, fmt.Errorf("%w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *TopicHandler) GetRegisteredHandlers(topic Topic) []string {
	var result []string

	for name := range r.handlers[topic] {
		result = append(result, name)
	}

	return result
}

func (r *TopicHandler) register(topic Topic, h untypedHandler) {
	if _, ok := r.handlers[topic]; !ok {
		r.handlers[topic] = make(map[string]untypedHandler)
	}

	handlerName := h.getName()
	if _, exists := r.handlers[topic][handlerName]; exists {
		panic(fmt.Sprintf("handler with name %q already registered for topic %q", handlerName, topic))
	}

	r.handlers[topic][handlerName] = h
}

type untypedHandler interface {
	getName() string
	canHandle(msg BackMessage) bool
	handleUntyped(ctx context.Context, msg BackMessage) error
}

type handlerAdapter[T Message] struct {
	name string
	h    Handler[T]
}

func (a handlerAdapter[T]) getName() string {
	return a.name
}

func (a handlerAdapter[T]) canHandle(backMsg BackMessage) bool {
	_, err := MessageFromPayload[T](backMsg.GetPayload())
	return err == nil
}

func (a handlerAdapter[T]) handleUntyped(ctx context.Context, backMsg BackMessage) error {
	msg, err := MessageFromPayload[T](backMsg.GetPayload())
	if err != nil {
		return fmt.Errorf("wrong message type for %q: have %T", a.name, msg)
	}
	return a.h.Handle(ctx, *msg)
}
