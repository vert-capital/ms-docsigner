package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

// publishMessageAPM envia mensagens usando a biblioteca Confluent Kafka com APM
func publishMessageAPM(topic string, message string) error {
	ctx := context.Background()

	// Configurar o produtor
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": KafkaBootstrapServers,
	})
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	defer producer.Close()

	// Iniciar a transação APM
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		tx = apm.DefaultTracer.StartTransaction("Produce "+topic, "kafka-producer")
		defer tx.End()
	}
	ctx = apm.ContextWithTransaction(ctx, tx)
	span, _ := apm.StartSpan(ctx, "Produce "+topic, "WriteMessage")
	span.Context.SetLabel("topic", topic)
	traceParent := apmhttp.FormatTraceparentHeader(span.TraceContext())
	span.End()

	// Criar a mensagem com o traceparent
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
		Headers: []kafka.Header{
			{Key: apmhttp.W3CTraceparentHeader, Value: []byte(traceParent)},
		},
	}

	// Enviar mensagem e capturar possíveis erros
	deliveryChan := make(chan kafka.Event)
	err = producer.Produce(msg, deliveryChan)
	if err != nil {
		apm.CaptureError(ctx, err).Send()
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Confirmar entrega
	e := <-deliveryChan
	m := e.(*kafka.Message)
	close(deliveryChan)

	if m.TopicPartition.Error != nil {
		apm.CaptureError(ctx, m.TopicPartition.Error).Send()
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}
	return nil
}

// readMessageMiddlewareAPM processa mensagens usando o consumidor Confluent Kafka com APM
func readMessageMiddlewareAPM(msg *kafka.Message, handleFunc func(m *kafka.Message) error) error {
	ctx := context.Background()

	// Obter traceparent do cabeçalho
	traceTransparent := getTraceparentHeader(msg)
	traceContext, _ := apmhttp.ParseTraceparentHeader(traceTransparent)
	opts := apm.TransactionOptions{
		TraceContext: traceContext,
	}

	// Iniciar a transação APM
	transaction := apm.DefaultTracer.StartTransactionOptions("Consume "+*msg.TopicPartition.Topic, "kafka-consumer", opts)
	ctx = apm.ContextWithTransaction(ctx, transaction)
	span, _ := apm.StartSpan(ctx, "Consume "+string(msg.Key), "ReadMessage")
	span.Context.SetLabel("topic", *msg.TopicPartition.Topic)
	span.Context.SetLabel("key", string(msg.Key))
	span.Context.SetLabel("offset", msg.TopicPartition.Offset.String())
	defer span.End()

	// Processar a mensagem com a função de callback
	err := handleFunc(msg)
	if err != nil {
		apm.CaptureError(ctx, err).Send()
		transaction.End()
		return err
	}
	transaction.End()
	return nil
}

// getTraceparentHeader retorna o cabeçalho traceparent da mensagem
func getTraceparentHeader(msg *kafka.Message) string {
	for _, header := range msg.Headers {
		if header.Key == apmhttp.W3CTraceparentHeader {
			return string(header.Value)
		}
	}
	return ""
}
