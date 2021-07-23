package gru

type Gru struct {
	Name       string `env:"GRU_NAME" envDefault:"Gru"`
	ManagePort string
	LogLevel   string `env:"GRU_LOG_LEVEL" envDefault:"debug"`
	Mode       string `env:"GRU_MODE" envDefault:"all"`
}
