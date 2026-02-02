// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/seakee/go-api/app/consumer"
	"github.com/sk-pkg/kafka"
)

// startKafkaConsumer initializes and starts the Kafka consumer based on the application configuration.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//
// This function checks the Kafka consumer configuration and initializes either an auto-submit
// or manual-submit consumer based on the settings.
func (a *App) startKafkaConsumer(ctx context.Context) {
	// Check if Kafka consumer is enabled in the configuration
	if a.Config.Kafka.ConsumerEnable {
		// Create a new consumer.Core with common dependencies
		core := &consumer.Core{
			Logger:        a.Logger,
			Redis:         a.Redis["go-api"],
			SqlDB:         a.SqlDB,
			KafkaConsumer: a.KafkaConsumer,
		}

		if a.Config.Kafka.ConsumerAutoSubmit {
			// Initialize auto-submit consumer
			consumer.NewAutoSubmit(ctx, core)
		} else {
			// Initialize manual-submit consumer
			consumer.New(ctx, core)
		}
	}
}

// loadKafka initializes Kafka producer and consumer based on the application configuration.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//
// Returns:
//   - error: An error if any occurred during the initialization process, nil otherwise.
//
// This function sets up both Kafka producer and consumer if they are enabled in the configuration.
// It uses the kafka package to create new instances with the specified options.
func (a *App) loadKafka(ctx context.Context) error {
	var err error

	// Initialize Kafka Producer if enabled
	if a.Config.Kafka.ProducerEnable {
		a.KafkaProducer, err = kafka.New(
			kafka.WithClientID(a.Config.Kafka.ClientID),
			kafka.WithProducerBrokers(a.Config.Kafka.Brokers),
			kafka.WithProducerRetryMax(a.Config.Kafka.MaxRetry),
			kafka.WithLogger(a.Logger.Zap),
		)

		if err != nil {
			return err
		}

		a.Logger.Info(ctx, "Kafka Producer loaded successfully")
	}

	// Initialize Kafka Consumer if enabled
	if a.Config.Kafka.ConsumerEnable {
		a.KafkaConsumer, err = kafka.New(
			kafka.WithClientID(a.Config.Kafka.ClientID),
			kafka.WithConsumerBrokers(a.Config.Kafka.Brokers),
			kafka.WithConsumerTopics(a.Config.Kafka.ConsumerTopics),
			kafka.WithConsumerGroup(a.Config.Kafka.ConsumerGroup),
			kafka.WithProducerRetryMax(a.Config.Kafka.MaxRetry),
			kafka.WithAutoSubmit(a.Config.Kafka.ConsumerAutoSubmit),
			kafka.WithLogger(a.Logger.Zap),
		)

		if err != nil {
			return err
		}

		a.Logger.Info(ctx, "Kafka Consumer loaded successfully")
	}

	return err
}
