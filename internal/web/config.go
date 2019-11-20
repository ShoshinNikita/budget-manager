package web

type Config struct {
	Port int `env:"SERVER_PORT" envDefault:"8080"`
}
