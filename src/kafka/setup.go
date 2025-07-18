package kafka

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	KafkaBootstrapServers string
	KafkaClientID         string
	KafkaGroupID          string
)

type KafkaReadTopicsParams struct {
	Topic   string
	Handler func(m *kafka.Message) error
}

// Configurações padrão para os tópicos
const (
	NumPartitions     = 3
	ReplicationFactor = -1
)

func kafkaSetup(topicParams []KafkaReadTopicsParams) {
	// TopicParams = topicParams

	KafkaBootstrapServers = os.Getenv("KAFKA_BOOTSTRAP_SERVER")
	KafkaClientID = os.Getenv("KAFKA_CLIENT_ID")
	KafkaGroupID = os.Getenv("KAFKA_GROUP_ID")

	ensureTopics(KafkaBootstrapServers, topicParams)

	log.Println("Kafka configurado com sucesso")
}

func ensureTopics(broker string, topicParams []KafkaReadTopicsParams) {
	// Timeout administrativo (exemplo: 30 segundos)
	adminTimeout := 30 * time.Second

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatalf("Failed to create AdminClient: %s\n", err)
	}
	defer adminClient.Close()

	// Criação dos tópicos
	var topicSpecifications []kafka.TopicSpecification
	for _, param := range topicParams {
		topicSpecifications = append(topicSpecifications, kafka.TopicSpecification{
			Topic:             param.Topic,
			NumPartitions:     NumPartitions,
			ReplicationFactor: ReplicationFactor,
		})
	}

	// Cria os tópicos (somente os que não existem)
	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()

	results, err := adminClient.CreateTopics(ctx, topicSpecifications)
	if err != nil {
		log.Fatalf("Erro ao criar tópicos: %v", err)
	}

	// Log o status da criação
	for _, result := range results {
		if result.Error.Code() == kafka.ErrTopicAlreadyExists {
			continue
		}
		if result.Error.Code() != kafka.ErrNoError {
			log.Printf("Erro ao criar tópico %s: %v\n", result.Topic, result.Error)
		} else {
			log.Printf("Tópico criado com sucesso: %s\n", result.Topic)
		}
	}
}

func readTopics(topicParams []KafkaReadTopicsParams) {
	if len(topicParams) == 0 {
		log.Println("Nenhum tópico para consumir")
		return
	}

	topics := make([]string, len(topicParams))
	for i, topicParam := range topicParams {
		topics[i] = topicParam.Topic
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        KafkaBootstrapServers,
		"group.id":                 KafkaGroupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"session.timeout.ms":       6000,
		"reconnect.backoff.ms":     50,
		"reconnect.backoff.max.ms": 1000,
	})
	if err != nil {
		log.Fatalf("Erro ao criar o consumidor: %v", err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatalf("Erro ao subscrever-se aos tópicos: %v", err)
	}

	log.Println("Consumidor iniciado. Aguardando mensagens...")

	// Mapeia os handlers para os tópicos
	handlerMap := make(map[string]func(*kafka.Message) error)
	for _, param := range topicParams {
		handlerMap[param.Topic] = param.Handler
	}

	// Inicia leitura das mensagens
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Erro ao ler mensagem: %v\n", err)
				continue
			}

			// Recuperar o handler associado ao tópico
			handler, exists := handlerMap[*msg.TopicPartition.Topic]
			if !exists {
				log.Printf("Nenhum handler encontrado para o tópico: %s\n", *msg.TopicPartition.Topic)
				continue
			}

			// Processar a mensagem usando o handler
			if err := readMessageMiddlewareAPM(msg, handler); err != nil {
				log.Printf("Erro ao processar mensagem do tópico %s: %v\n", *msg.TopicPartition.Topic, err)
				continue
			}

			// Commit manual após sucesso
			_, commitErr := consumer.CommitMessage(msg)
			if commitErr != nil {
				log.Printf("Erro ao fazer commit do offset: %v", commitErr)
			} else {
				log.Printf("Leitura confirmada com sucesso para o tópico %s", *msg.TopicPartition.Topic)
			}
		}
	}()

	// Aguardar todas as goroutines terminarem (poderia ser um shutdown controlado)
	wg.Wait()
}

func PublishMessage(topic string, message string) error {
	return publishMessageAPM(topic, message)
}
