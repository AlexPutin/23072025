package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host         string        `mapstructure:"host"`
		Port         int           `mapstructure:"port"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	} `mapstructure:"server"`

	Service struct {
		MaxActiveTasks    int      `mapstructure:"max_active_tasks"`
		MaxFilesPerTask   int      `mapstructure:"max_files_per_task"`
		AllowedExtensions []string `mapstructure:"allowed_extensions"`
		ArchiveDirectory  string   `mapstructure:"archive_directory"`
	} `mapstructure:"service"`
}

func MustLoad(file string) *Config {
	viper.SetConfigFile(file)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var appConfig Config
	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %s", err)
	}

	return &appConfig
}
