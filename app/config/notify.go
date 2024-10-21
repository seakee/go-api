package config

// Notify defines notification configuration options.
type Notify struct {
	DefaultChannel string `json:"default_channel"`
	DefaultLevel   string `json:"default_level"`
	Lark           Lark   `json:"lark"`
}

// Lark defines Lark configuration options.
type Lark struct {
	Enable                 bool               `json:"enable"`
	DefaultSendChannelName string             `json:"default_send_channel_name"`
	ChannelSize            int                `json:"channel_size"`
	PoolSize               int                `json:"pool_size"`
	BotWebhooks            map[string]string  `json:"bot_webhooks"`
	Larks                  map[string]LarkApp `json:"larks"`
}

// LarkApp defines Lark application configuration options.
type LarkApp struct {
	AppType   string `json:"app_type"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}
