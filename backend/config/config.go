package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Log           LogConfig   `mapstructure:"log"`
	HTTP          HTTPConfig  `mapstructure:"http"`
	AdminPassword string      `mapstructure:"admin_password"`
	PG            PGConfig    `mapstructure:"pg"`
	MQ            MQConfig    `mapstructure:"mq"`
	RAG           RAGConfig   `mapstructure:"rag"`
	Redis         RedisConfig `mapstructure:"redis"`
	Auth          AuthConfig  `mapstructure:"auth"`
	S3            S3Config    `mapstructure:"s3"`
	CaddyAPI      string      `mapstructure:"caddy_api"`
	SubnetPrefix  string      `mapstructure:"subnet_prefix"`
}

type LogConfig struct {
	Level int `mapstructure:"level"`
}

type HTTPConfig struct {
	Port int `mapstructure:"port"`
}

type PGConfig struct {
	DSN string `mapstructure:"dsn"`
}

type MQConfig struct {
	Type string     `mapstructure:"type"`
	NATS NATSConfig `mapstructure:"nats"`
}

type NATSConfig struct {
	Server   string `mapstructure:"server"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type RAGConfig struct {
	Provider string      `mapstructure:"provider"`
	CTRAG    CTRAGConfig `mapstructure:"ct_rag"`
}

type CTRAGConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
}

type AuthConfig struct {
	Type string    `mapstructure:"type"`
	JWT  JWTConfig `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// getMinioEndpoint returns minio endpoint from environment variable or default
func getMinioEndpoint() string {
	if minioHost := os.Getenv("MINIO_HOST"); minioHost != "" {
		return minioHost + ":9000"
	}
	return "panda-wiki-minio:9000"
}

// getPostgresHost returns postgres host from environment variable or default
func getPostgresHost() string {
	if pgHost := os.Getenv("POSTGRES_HOST"); pgHost != "" {
		return pgHost
	}
	return "panda-wiki-postgres"
}

// getRedisAddr returns redis address from environment variable or default
func getRedisAddr() string {
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		return redisHost + ":6379"
	}
	return "panda-wiki-redis:6379"
}

// getNatsServer returns NATS server URL from environment variable or default
func getNatsServer() string {
	if natsHost := os.Getenv("NATS_HOST"); natsHost != "" {
		return fmt.Sprintf("nats://%s:4222", natsHost)
	}
	// fallback to SUBNET_PREFIX based configuration
	SUBNET_PREFIX := os.Getenv("SUBNET_PREFIX")
	if SUBNET_PREFIX == "" {
		SUBNET_PREFIX = "169.254.15"
	}
	return fmt.Sprintf("nats://%s.13:4222", SUBNET_PREFIX)
}

// getRagBaseURL returns RAG base URL from environment variable or default
func getRagBaseURL() string {
	if ragHost := os.Getenv("RAG_HOST"); ragHost != "" {
		return fmt.Sprintf("http://%s:8080/api/v1", ragHost)
	}
	// fallback to SUBNET_PREFIX based configuration
	SUBNET_PREFIX := os.Getenv("SUBNET_PREFIX")
	if SUBNET_PREFIX == "" {
		SUBNET_PREFIX = "169.254.15"
	}
	return fmt.Sprintf("http://%s.18:8080/api/v1", SUBNET_PREFIX)
}

func NewConfig() (*Config, error) {
	// set default config
	SUBNET_PREFIX := os.Getenv("SUBNET_PREFIX")
	if SUBNET_PREFIX == "" {
		SUBNET_PREFIX = "169.254.15"
	}
	defaultConfig := &Config{
		Log: LogConfig{
			Level: 0,
		},
		AdminPassword: "",
		HTTP: HTTPConfig{
			Port: 8000,
		},
		PG: PGConfig{
			DSN: fmt.Sprintf("host=%s user=panda-wiki password=panda-wiki-secret dbname=panda-wiki port=5432 sslmode=disable TimeZone=Asia/Shanghai", getPostgresHost()),
		},
		MQ: MQConfig{
			Type: "nats",
			NATS: NATSConfig{
				Server:   getNatsServer(),
				User:     "panda-wiki",
				Password: "",
			},
		},
		RAG: RAGConfig{
			Provider: "ct",
			CTRAG: CTRAGConfig{
				BaseURL: getRagBaseURL(),
				APIKey:  "sk-1234567890",
			},
		},
		Redis: RedisConfig{
			Addr:     getRedisAddr(),
			Password: "",
		},
		Auth: AuthConfig{
			Type: "jwt",
			JWT:  JWTConfig{Secret: ""},
		},
		S3: S3Config{
			Endpoint:  getMinioEndpoint(),
			AccessKey: "s3panda-wiki",
			SecretKey: "",
		},
		CaddyAPI:     "/app/run/caddy-admin.sock",
		SubnetPrefix: "169.254.15",
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	// try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// if config file not found, return default config
			return nil, err
		}
	}

	// merge config file values to default config
	if err := viper.Unmarshal(defaultConfig); err != nil {
		return nil, err
	}

	// finally, override sensitive info with env variables
	overrideWithEnv(defaultConfig)

	return defaultConfig, nil
}

// overrideWithEnv override sensitive info with env variables
func overrideWithEnv(c *Config) {
	if env := os.Getenv("POSTGRES_PASSWORD"); env != "" {
		c.PG.DSN = fmt.Sprintf("host=%s user=panda-wiki password=%s dbname=panda-wiki port=5432 sslmode=disable TimeZone=Asia/Shanghai", getPostgresHost(), env)
	}
	if env := os.Getenv("NATS_PASSWORD"); env != "" {
		c.MQ.NATS.Password = env
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		c.Redis.Password = env
	}
	if env := os.Getenv("JWT_SECRET"); env != "" {
		c.Auth.JWT.Secret = env
	}
	if env := os.Getenv("S3_SECRET_KEY"); env != "" {
		c.S3.SecretKey = env
	}
	if env := os.Getenv("ADMIN_PASSWORD"); env != "" {
		c.AdminPassword = env
	}
	if env := os.Getenv("SUBNET_PREFIX"); env != "" {
		c.SubnetPrefix = env
	}
	// pg
	if env := os.Getenv("PG_DSN"); env != "" {
		c.PG.DSN = env
	}
	// nats
	if env := os.Getenv("MQ_NATS_SERVER"); env != "" {
		c.MQ.NATS.Server = env
	}
	// rag
	if env := os.Getenv("RAG_CT_RAG_BASE_URL"); env != "" {
		c.RAG.CTRAG.BaseURL = env
	}
	// redis
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		c.Redis.Addr = env
	}
	// s3
	if env := os.Getenv("S3_ENDPOINT"); env != "" {
		c.S3.Endpoint = env
	}
}

func (*Config) GetString(key string) string {
	return viper.GetString(key)
}

func (*Config) GetInt(key string) int {
	return viper.GetInt(key)
}

func (*Config) GetUint64(key string) uint64 {
	return viper.GetUint64(key)
}

func (*Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (*Config) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (*Config) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}
