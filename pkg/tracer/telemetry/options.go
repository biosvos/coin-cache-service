package telemetry

type Options struct {
	URL string
}

func NewOptions() *Options {
	return &Options{
		URL: "localhost:4318",
	}
}

type Option func(*Options)

func WithURL(url string) Option {
	return func(o *Options) {
		o.URL = url
	}
}
