package acquirer

type Config struct {
	HTTPAddr    string `yaml:"http_addr"`
	ISO8583Addr string `yaml:"iso8583_addr"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr:    "127.0.0.1:8080",
		ISO8583Addr: "127.0.0.1:8583",
	}
}
