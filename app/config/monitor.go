package config

// Monitor defines monitoring configuration options.
type Monitor struct {
	PanicRobot PanicRobot `json:"panic_robot"` // Panic robot configuration
}

// PanicRobot defines configuration for panic reporting.
type PanicRobot struct {
	Enable bool        `json:"enable"` // Panic robot enable flag
	Wechat robotConfig `json:"wechat"` // WeChat's configuration for panic reporting
	Feishu robotConfig `json:"feishu"` // Feishu configuration for panic reporting
}

// robotConfig defines configuration for messaging platforms.
type robotConfig struct {
	Enable  bool   `json:"enable"`   // Robot enable flag
	PushUrl string `json:"push_url"` // URL for pushing messages
}
