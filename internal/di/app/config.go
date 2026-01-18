package app


type Config struct {
	ServerAddress string
	BaseURL       string
}


func NewConfig(serverAddress, baseURL string) *Config {
	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}
}
