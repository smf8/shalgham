package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		Debug    bool     `koanf:"debug"`
		Postgres Postgres `koanf:"postgres"`
		Server   Server   `koanf:"server"`
		Logger   Logger   `koanf:"logger"`
	}

	Server struct {
		Address string `koanf:"address"`
	}

	Logger struct {
		Level   string `koanf:"level"`
		Enabled bool   `koanf:"enabled"`
	}

	Postgres struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Username string `koanf:"username"`
		Password string `koanf:"password"`
		DBName   string `koanf:"dbname"`
	}
)

func New() Config {
	var instance Config

	k := koanf.New(".")

	if err := k.Load(structs.Provider(def, "konaf"), nil); err != nil {
		logrus.Fatalf("error loading default: %s", err)
	}

	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		logrus.Errorf("error loading file: %s", err)
	}

	if err := k.Unmarshal("", &instance); err != nil {
		logrus.Fatalf("error unmarshalling config: %s", err)
	}

	logrus.Infof("following configuration is loaded:\n%+v", instance)

	return instance
}
