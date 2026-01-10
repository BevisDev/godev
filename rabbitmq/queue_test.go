package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeclare_DirectExchange(t *testing.T) {
	mq, err := NewMQTest()
	require.NoError(t, err)
	defer mq.Close()

	queues := []string{
		"it.direct.queue",
		"it.direct.queue1",
	}

	err = mq.Queue.DeclareSimple(queues...)
	require.NoError(t, err)
}

func TestDeclare_TopicExchange(t *testing.T) {
	mq, err := NewMQTest()
	require.NoError(t, err)
	defer mq.Close()

	queue := "it.topic.queue"
	exchange := "it.topic.exchange"

	err = mq.Queue.Declare(Spec{
		Queues: []QueueSpec{
			{Queue: queue},
		},
		Exchanges: []ExchangeSpec{
			{
				Name: exchange,
				Type: Topic,
				Bindings: []BindingSpec{
					{
						Queue:      queue,
						RoutingKey: "order.*",
					},
				},
			},
		},
	})
	require.NoError(t, err)
}

func TestDeclare_FanoutExchange(t *testing.T) {
	mq, err := NewMQTest()
	require.NoError(t, err)
	defer mq.Close()

	queue1 := "it.fanout.q1"
	queue2 := "it.fanout.q2"
	exchange := "it.fanout.exchange"

	err = mq.Queue.Declare(Spec{
		Queues: []QueueSpec{
			{Queue: queue1},
			{Queue: queue2},
		},
		Exchanges: []ExchangeSpec{
			{
				Name: exchange,
				Type: Topic,
				Bindings: []BindingSpec{
					{Queue: queue1},
					{Queue: queue2},
				},
			},
		},
	})
	require.NoError(t, err)
}
