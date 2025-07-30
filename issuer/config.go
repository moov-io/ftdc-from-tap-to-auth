package issuer

// Config is a configuration for the issuer application
type Config struct {
	HTTPAddr            string `yaml:"http_addr"`
	ISO8583Addr         string `yaml:"iso8583_addr"`
	CardPersonalizerURL string `yaml:"card_personalizer_url"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr:            "localhost:9090",
		ISO8583Addr:         "localhost:8583",
		CardPersonalizerURL: "http://localhost:7070",
	}
}
