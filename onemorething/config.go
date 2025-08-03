package onemorething

type Config struct {
	ServerAddr string `yaml:"server_addr"`
	PrinterURL string `yaml:"printer_url"` // URL of the printer service
}
