// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/seakee/go-api/app/consumer"
	"github.com/sk-pkg/kafka"
)

func (a *App) startKafkaConsumer(ctx context.Context) {
	if a.Config.Kafka.ConsumerEnable {
		if a.Config.Kafka.ConsumerAutoSubmit {
			// 自动提交场景
			consumer.NewAutoSubmit(ctx, &consumer.Core{
				Logger:        a.Logger,
				Redis:         a.Redis["go-api"],
				MysqlDB:       a.MysqlDB,
				KafkaConsumer: a.KafkaConsumer,
			})
		} else {
			// 手动提交场景
			consumer.New(ctx, &consumer.Core{
				Logger:        a.Logger,
				Redis:         a.Redis["go-api"],
				MysqlDB:       a.MysqlDB,
				KafkaConsumer: a.KafkaConsumer,
			})
		}
	}
}

// loadKafka 加载kafka
func (a *App) loadKafka(ctx context.Context) error {
	var err error
	// 初始化Producer
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

	// 初始化Consumer
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
