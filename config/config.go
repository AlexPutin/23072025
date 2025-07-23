package config

type Config struct {
	Server struct {
		Host string
		Port int
	}

	Service struct {
		MaxFilesPerTask   int
		AllowedExtensions []string
	}
}
