package stream

import (
	"context"
	"github.com/applike/gosoline/pkg/mon"
)

type ConsumerAcknowledge struct {
	logger mon.Logger
	input  Input
}

func NewConsumerAcknowledgeWithInterfaces(logger mon.Logger, input Input) ConsumerAcknowledge {
	return ConsumerAcknowledge{
		logger: logger,
		input:  input,
	}
}

func (c *ConsumerAcknowledge) Acknowledge(ctx context.Context, msg *Message) {
	var ok bool
	var ackInput AcknowledgeableInput

	if ackInput, ok = c.input.(AcknowledgeableInput); !ok {
		return
	}

	err := ackInput.Ack(msg)

	if err != nil {
		c.logger.WithContext(ctx).Error(err, "could not acknowledge the message")
	}
}

func (c *ConsumerAcknowledge) AcknowledgeBatch(ctx context.Context, msg []*Message) {
	var ok bool
	var ackInput AcknowledgeableInput

	if ackInput, ok = c.input.(AcknowledgeableInput); !ok {
		return
	}

	err := ackInput.AckBatch(msg)

	if err != nil {
		c.logger.WithContext(ctx).Error(err, "could not acknowledge the messages")
	}
}
