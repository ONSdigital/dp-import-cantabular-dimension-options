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
	kafkaProducer, err := kafka.NewProducer(ctx, config.KafkaAddr, config.CategoryDimensionImportTopic, pChannels, &kafka.ProducerConfig{
		KafkaVersion: &config.KafkaVersion,
	})
	if err != nil {
		log.Fatal(ctx, "fatal error trying to create kafka producer", err, log.Data{"topic": config.CategoryDimensionImportTopic})
		os.Exit(1)
	}

	// kafka error logging go-routines
	kafkaProducer.Channels().LogErrors(ctx, "kafka producer")

	time.Sleep(500 * time.Millisecond)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		e := scanEvent(scanner)
		log.Info(ctx, "sending category-dimension-import event", log.Data{"CategoryDimensionImportEvent": e})

		bytes, err := schema.CategoryDimensionImport.Marshal(e)
		if err != nil {
			log.Fatal(ctx, "category-dimension-import event error", err)
			os.Exit(1)
		}

		// Wait for producer to be initialised
		<-kafkaProducer.Channels().Ready

		// Send bytes to output channel
		kafkaProducer.Channels().Output <- bytes
	}
}

// scanEvent creates a CategoryDimensionImport event according to the user input
func scanEvent(scanner *bufio.Scanner) *event.CategoryDimensionImport {
	fmt.Println("--- [Send Kafka CategoryDimensionImport] ---")

	e := &event.CategoryDimensionImport{}

	fmt.Println("Please type the Job ID")
	fmt.Printf("$ ")
	scanner.Scan()
	e.JobID = scanner.Text()

	fmt.Println("Please type the Dimension ID")
	fmt.Printf("$ ")
	scanner.Scan()
	e.DimensionID = scanner.Text()

	fmt.Println("Please type the Instance ID")
	fmt.Printf("$ ")
	scanner.Scan()
	e.InstanceID = scanner.Text()

	return e
}
