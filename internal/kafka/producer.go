package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
)

// ErrBufferFull is returned when the async producer's internal buffer is full,
// meaning Kafka cannot keep up with the incoming rate.
var ErrBufferFull = errors.New("kafka producer input buffer full")

type Producer struct {
	producer sarama.AsyncProducer
	topic    string
	wg       sync.WaitGroup
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.ChannelBufferSize = 10000

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	p := &Producer{producer: producer, topic: topic}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for kerr := range producer.Errors() {
			slog.Error("kafka async publish error", slog.Any("error", kerr.Err))
		}
	}()

	return p, nil
}

// SendMessage enqueues a message for async delivery to Kafka.
// Returns ErrBufferFull if the internal buffer is saturated.
func (p *Producer) SendMessage(key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(valueBytes),
	}

	select {
	case p.producer.Input() <- msg:
		return nil
	default:
		return ErrBufferFull
	}
}

func (p *Producer) Close(ctx context.Context) error {
	p.producer.AsyncClose()
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
