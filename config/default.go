package config

const Namespace = "Shalgham"

//nolint:lll,gochecknoglobals,gomnd
var def = Config{
	Debug: true,

	Server: Server{
		Address: "127.0.0.1:1234",
	},

	Logger: Logger{
		Level: "debug",
	},

	Postgres: Postgres{
		Host:     "127.0.0.1",
		Port:     54320,
		DBName:   "shalgham",
		Username: "postgres",
		Password: "pass",
	},
}
