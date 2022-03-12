package config

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8081"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://localhost:5432/gophermart?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"localhost:8080"`
}
