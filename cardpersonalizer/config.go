package cardpersonalizer

type Config struct {
	HTTPAddr   string
	CardReader string
}

func DefaultConfig() *Config {
	return &Config{
		HTTPAddr:   "127.0.0.1:7070",
		CardReader: "ACS ACR1252 Dual Reader PICC",
		//CardReader: "ACS ACR122U PICC Interface",
	}
}
