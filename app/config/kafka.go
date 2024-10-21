package config

// Kafka defines Kafka configuration options.
type Kafka struct {
	Brokers            []string `json:"brokers"`              // Kafka broker addresses
	MaxRetry           int      `json:"max_retry"`            // Maximum number of retries
	ClientID           string   `json:"client_id"`            // Kafka client ID
	ProducerEnable     bool     `json:"producer_enable"`      // Producer enable flag
	ConsumerEnable     bool     `json:"consumer_enable"`      // Consumer enable flag
	ConsumerGroup      string   `json:"consumer_group"`       // Consumer group name
	ConsumerTopics     []string `json:"consumer_topics"`      // Topics to consume
	ConsumerAutoSubmit bool     `json:"consumer_auto_submit"` // Auto-submit consumer offsets flag
}
