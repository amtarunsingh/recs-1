package messaging

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock message and handler for testing
type testMessage struct {
	Data string
}

func (m *testMessage) GetDeduplicationId() string {
	return "test-dedup-id"
}

func (m *testMessage) GetPayload() Payload {
	return Payload(m.Data)
}

func (m *testMessage) Load(payload Payload) error {
	m.Data = string(payload)
	return nil
}

type testHandler struct {
	name string
}

func (h *testHandler) GetName() string {
	return h.name
}

func (h *testHandler) Handle(ctx context.Context, message *testMessage) error {
	return nil
}

func TestRegisterTopicHandler_PanicsOnDuplicateName(t *testing.T) {
	logger := slog.Default()
	topicHandler := NewTopicHandler(logger)

	topic := Topic("test-topic")
	handler1 := &testHandler{name: "duplicate_handler"}
	handler2 := &testHandler{name: "duplicate_handler"}

	// Register first handler - should succeed
	RegisterTopicHandler(topicHandler, topic, handler1)

	// Register second handler with same name - should panic
	assert.Panics(t, func() {
		RegisterTopicHandler(topicHandler, topic, handler2)
	}, "Expected panic when registering handler with duplicate name")
}

func TestRegisterTopicHandler_AllowsSameNameOnDifferentTopics(t *testing.T) {
	logger := slog.Default()
	topicHandler := NewTopicHandler(logger)

	topic1 := Topic("test-topic-1")
	topic2 := Topic("test-topic-2")
	handler1 := &testHandler{name: "shared_handler"}
	handler2 := &testHandler{name: "shared_handler"}

	// Register same handler name on different topics - should succeed
	assert.NotPanics(t, func() {
		RegisterTopicHandler(topicHandler, topic1, handler1)
		RegisterTopicHandler(topicHandler, topic2, handler2)
	}, "Should allow same handler name on different topics")
}

func TestRegisterTopicHandler_AllowsDifferentNamesOnSameTopic(t *testing.T) {
	logger := slog.Default()
	topicHandler := NewTopicHandler(logger)

	topic := Topic("test-topic")
	handler1 := &testHandler{name: "handler_1"}
	handler2 := &testHandler{name: "handler_2"}

	// Register different handlers on same topic - should succeed
	assert.NotPanics(t, func() {
		RegisterTopicHandler(topicHandler, topic, handler1)
		RegisterTopicHandler(topicHandler, topic, handler2)
	}, "Should allow different handler names on same topic")
}
