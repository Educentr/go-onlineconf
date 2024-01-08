package onlineconf

type Option interface {
	apply(*OnlineconfInstance)
}

type options func(*OnlineconfInstance)

func (o options) apply(oi *OnlineconfInstance) {
	o(oi)
}

func WithLogger(logger OnlineconfLogger) Option {
	return options(func(oi *OnlineconfInstance) {
		oi.logger = logger
	})
}

func WithConfigDir(path string) Option {
	return options(func(oi *OnlineconfInstance) {
		oi.configDir = path
	})
}
