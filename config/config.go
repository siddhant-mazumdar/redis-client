package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type IConfig interface {
	GetConfig() config
}

type EmailTrackingConfig struct {
	OpenTrackingURL  string `mapstructure:"openTrackingURL"`
	ClickTrackingURL string `mapstructure:"clickTrackingURL"`
}

type SendGridConfig struct {
	Key             string   `mapstructure:"key"`
	SubUserPassword string   `mapstructure:"subUserPassword"`
	Ips             []string `mapstructure:"ips"`
	Email           string   `mapstructure:"email"`
	Url             string   `mapstructure:"url"`
}

type SubscriptionConfig struct {
	DBName  string `mapstructure:"dbName"`
	BotUser string `mapstructure:"botUser"`
}

type NoReplyEmailConfig struct {
	FromEmail string `mapstructure:"fromEmail"`
	Name      string `mapstructure:"name"`
}

type config struct {
	DatabaseUrl           string             `mapstructure:"databaseUrl"`
	DatabaseName          string             `mapstructure:"databaseName"`
	SQLiteDBPath          string             `mapstructure:"sqliteDbPath"`
	MaxDatabaseConnection int                `mapstructure:"maxDatabaseConnection"`
	KafkaServer           string             `mapstructure:"kafkaServer"`
	RedisServer           string             `mapstructure:"redisServer"`
	NatsServer            string             `mapstructure:"natsServer"`
	ServiceEndpoints      []serviceEndpoints `mapstructure:"serviceEndpoints"`
	Environment           string
	EmailTracking         EmailTrackingConfig    `mapstructure:"emailTracking"`
	SendGrid              SendGridConfig         `mapstructure:"sendGrid"`
	ServiceEndpointPrefix string                 `mapstructure:"serviceEndpointPrefix"`
	BotUserId             int                    `mapstructure:"botUserId"`
	BucketName            string                 `mapstructure:"bucketName"`
	GcpUrl                string                 `mapstructure:"gcpUrl"`
	FromEmail             string                 `mapstructure:"fromEmail"`
	MessagingQueueTopics  []messagingQueueConfig `mapstructure:"messagingQueueTopics"`
}

type jetStreamConfiguration struct {
	Stream   string `json:"stream"`
	Subject  string `json:"subject"`
	Consumer string `json:"consumer"`
}

func NewConfig() IConfig {
	environment := os.Getenv("GO_ENV")
	if environment == "" {
		environment = "development"
	}
	configPath := filepath.Join("config", "environments", environment)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("common-config")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}

	viper.SetConfigName("service-endpoints")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
	c := config{}
	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}
	c.Environment = environment
	return c
}

func (c config) GetConfig() config {
	return c
}
