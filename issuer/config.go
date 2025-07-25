package issuer

// Config is a configuration for the issuer application
type Config struct {
	HTTPAddr             string
	ISO8583Addr          string
	CardPersonalizerAddr string
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr:             "localhost:9090",
		ISO8583Addr:          "localhost:8583",
		CardPersonalizerAddr: "http://localhost:7070",
	}
}
