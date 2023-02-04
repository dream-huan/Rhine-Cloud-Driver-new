package config

type Config struct {
	Server                    ServerConfig              `yaml:"server"`
	RedisManager              RedisConfig               `yaml:"redis"`
	MysqlManager              MysqlConfig               `yaml:"mysql"`
	GoogleRecaptchaPrivateKey RecaptchaPrivateKeyConfig `yaml:"googlerecaptchaprivatekey"`
	JwtKey                    JwtConfig                 `yaml:"jwtkey"`
	Log                       LogConfig                 `yaml:"log"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
}

type RedisConfig struct {
	Address  []string `yaml:"addr"`
	Password string   `yaml:"pwd"`
}

type MysqlConfig struct {
	Address  string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"pwd"`
	Database string `yaml:"database"`
}

type RecaptchaPrivateKeyConfig struct {
	Key string `yaml:"key"`
}

type JwtConfig struct {
	Key string `yaml:"key"`
}

type LogConfig struct {
	LogPath     string `yaml:"log_path"`
	MaxSize     int    `yaml:"max_log_size"`
	MaxBackup   int    `yaml:"max_backup"`
	MaxAge      int    `yaml:"max_age"`
	ServiceName string `yaml:"service_name"`
	LogLevel    int    `yaml:"level"`
	Compress    bool   `yaml:"compress"`
	LogConsole  bool   `yaml:"log_console"`
}

type ShareKeyHaltConfig struct {
	Key string `yaml:"key"`
}
