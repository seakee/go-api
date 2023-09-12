package bootstrap

import (
	"github.com/seakee/go-api/app/consumer"
	"github.com/sk-pkg/kafka"
)

func (a *App) startKafkaConsumer() {
	if a.Config.Kafka.ConsumerEnable {
		if a.Config.Kafka.ConsumerAutoSubmit {
			// 自动提交场景
			consumer.NewAutoSubmit(&consumer.Core{
				Logger:        a.Logger,
				Redis:         a.Redis["go-api"],
				MysqlDB:       a.MysqlDB,
				KafkaConsumer: a.KafkaConsumer,
			})
		} else {
			// 手动提交场景
			consumer.New(&consumer.Core{
				Logger:        a.Logger,
				Redis:         a.Redis["go-api"],
				MysqlDB:       a.MysqlDB,
				KafkaConsumer: a.KafkaConsumer,
			})
		}
	}
}

// loadKafka 加载kafka
func (a *App) loadKafka() error {
	var err error
	// 初始化Producer
	if a.Config.Kafka.ProducerEnable {
		a.KafkaProducer, err = kafka.New(
			kafka.WithClientID(a.Config.Kafka.ClientID),
			kafka.WithProducerBrokers(a.Config.Kafka.Brokers),
			kafka.WithProducerRetryMax(a.Config.Kafka.MaxRetry),
			kafka.WithLogger(a.Logger),
		)

		if err != nil {
			return err
		}

		a.Logger.Info("Kafka Producer loaded successfully")
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
			kafka.WithLogger(a.Logger),
		)

		if err != nil {
			return err
		}

		a.Logger.Info("Kafka Consumer loaded successfully")
	}

	return err
}
