package consumer

import (
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	Core struct {
		Logger        *zap.Logger
		Redis         *redis.Manager
		MysqlDB       map[string]*gorm.DB
		KafkaConsumer *kafka.Manager
	}
)

func NewAutoSubmit(core *Core) {
	core.Logger.Info("Kafka Auto Submit Consumer started successfully")
	for {
		select {
		// 取一条消费信息
		case msg := <-core.KafkaConsumer.ConsumerMessages:
			switch msg.Topic {
			case "topic1": // 对监听到的topic1进行消费
				continue
			case "topic2": // 对监听到的topic2进行消费
				continue
			}
		}
	}
}

func New(core *Core) {
	// handler := test.New(core.Logger, core.Redis, core.MysqlDB["test"])

	core.Logger.Info("Kafka Consumer started successfully")
	for {
		select {
		// 取一个消费者
		// 需要实现手动提交时，需要将消费者consumer传入处理逻辑中
		// 调用consumer.Submit()提交当前消息
		case consumer := <-core.KafkaConsumer.Consumers:
			msg := consumer.GetMsg()
			switch msg.Topic {
			case "test": // 对监听到的topic1进行消费
				// go handler.Create(consumer)
				continue
			}
		}
	}
}
