package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ONSdigital/dp-import-cantabular-dimension-options/config"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/event"
	"github.com/ONSdigital/dp-import-cantabular-dimension-options/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "dp-import-cantabular-dimension-options"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	// Get Config
	config, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting config", err)
		os.Exit(1)
	}

	// Create Kafka Producer
	pChannels := kafka.CreateProducerChannels()
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.HelloCalledTopic, pChannels, &kafka.ProducerConfig{
		KafkaVersion: &config.KafkaVersion,
	})
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": config.HelloCalledTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Event(ctx, "sending hello-called event", log.INFO, log.Data{"helloCalledEvent": e})

		bytes, err := schema.HelloCalledEvent.Marshal(e)
		if err != nil {
			log.Fatal(ctx, "hello-called event error", err)
			os.Exit(1)
		}

		// Send bytes to Output channel, after calling Initialise just in case it is not initialised.
		kafkaProducer.Initialise(ctx)
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a HelloCalled event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.HelloCalled {
	fmt.Println("--- [Send Kafka HelloCalled] ---")

	fmt.Println("Please type the recipient name")
	fmt.Printf("$ ")
	scanner.Scan()
	name := scanner.Text()

	return &event.HelloCalled{
		RecipientName: name,
	}
}
