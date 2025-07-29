package cardpersonalizer

type Config struct {
	HTTPAddr   string
	CardReader string
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr:   "0.0.0.0:7070",
		CardReader: "ACR1252 Dual Reader PICC",
		//CardReader: "ACR122U PICC Interface",
	}
}
