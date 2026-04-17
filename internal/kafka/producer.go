package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

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

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	slog.Info(
		"Message sent to partition",
		slog.Int("partition", int(partition)),
		slog.Int64("offset", offset),
	)
	return nil
}

func (p *Producer) Close(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- p.producer.Close()
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
