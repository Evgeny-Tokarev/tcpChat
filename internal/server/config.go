package server

type Config struct {
	Port int `env:"SERVER_PORT" envDefault:"13002"`
}
