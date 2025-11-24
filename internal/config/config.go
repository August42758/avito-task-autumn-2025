package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	DBName     string `env:"DB_NAME"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
}

func Load(path string) (*Config, error) {

	if path != "" {
		// путь для докера
		if err := godotenv.Load(path); err != nil {
			return nil, err
		}
	} else {
		// дефолтный путь для локального запуска
		if err := godotenv.Load("././.env"); err != nil {
			return nil, err
		}
	}
	c := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
	}

	//если запуск в докере, то устанавливаем название хоста для БД как имя контейнера
	if os.Getenv("DOCKER") == "1" {
		c.DBHost = "postgresql"
	}

	return c, nil
}
