package config

const Namespace = "Shalgham"

//nolint:lll
var Default = Config{
	Debug: true,

	Server: Server{
		Address: "127.0.0.1:1234",
	},

	Logger: Logger{
		Level: "debug",
	},

	Postgres: Postgres{
		Host:     "127.0.0.1",
		Port:     5432,
		DBName:   "shalgham",
		Username: "postgres",
		Password: "pass",
	},
}
