// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package consumer

import (
	"context"

	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Core represents the core dependencies required for the consumer package.
type Core struct {
	Logger        *logger.Manager
	Redis         *redis.Manager
	MysqlDB       map[string]*gorm.DB
	KafkaConsumer *kafka.Manager
}

// NewAutoSubmit starts a Kafka consumer for auto-submission.
//
// Parameters:
//   - ctx: The context for managing the lifecycle of the function.
//   - core: A pointer to the Core struct containing necessary dependencies.
//
// This function continuously listens for messages on specific Kafka topics
// and processes them accordingly. It runs indefinitely until the context is cancelled.
func NewAutoSubmit(ctx context.Context, core *Core) {
	core.Logger.Info(ctx, "Kafka Auto Submit Consumer started successfully")
	for {
		select {
		// Consume a message from Kafka
		case msg := <-core.KafkaConsumer.ConsumerMessages:
			switch msg.Topic {
			case "topic1":
				// Process messages from topic1
				continue
			case "topic2":
				// Process messages from topic2
				continue
			}
		}
	}
}

// New starts a Kafka consumer that manually commits messages.
//
// Parameters:
//   - ctx: The context for managing the lifecycle of the function.
//   - core: A pointer to the Core struct containing necessary dependencies.
//
// This function continuously listens for Kafka consumers and processes
// messages from specific topics. It runs indefinitely until the context is cancelled.
func New(ctx context.Context, core *Core) {
	// Uncomment and initialize the handler if needed
	// handler := test.New(core.Logger, core.Redis, core.MysqlDB["test"])

	core.Logger.Info(ctx, "Kafka Consumer started successfully")
	for {
		select {
		// Get a consumer
		// For manual commit, pass the consumer to the processing logic
		// Call consumer.Submit() to commit the current message
		case consumer := <-core.KafkaConsumer.Consumers:
			msg := consumer.GetMsg()
			switch msg.Topic {
			case "test":
				// Process messages from the "test" topic
				// Uncomment the following line to use the handler
				// go handler.Create(consumer)
				continue
			}
		}
	}
}
