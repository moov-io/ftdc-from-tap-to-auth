package cardpersonalizer

type Config struct {
	HTTPAddr string
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr: "127.0.0.1:7070",
	}
}
