package feishu

// Option 客户端选项函数
type Option func(*Config)

// WithAppID 设置 App ID
func WithAppID(appID string) Option {
	return func(c *Config) {
		c.AppID = appID
	}
}

// WithAppSecret 设置 App Secret
func WithAppSecret(appSecret string) Option {
	return func(c *Config) {
		c.AppSecret = appSecret
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}
