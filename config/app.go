package config

type AppConfig struct {
	Env Env
}

func Load() AppConfig {
	return AppConfig{
		Env: LoadEnv(),
	}
}
