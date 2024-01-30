package util

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ENVIRONMENT          string        `mapstructure:"ENVIRONMENT"`
	DBUrlMySQL           string        `mapstructure:"DB_URL_MYSQL"`
	MigrationDBUrl       string        `mapstructure:"MIGRATION_DB_URL"`
	MigrationSourceUrl   string        `mapstructure:"MIGRATION_SOURCE_URL"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	SessionDuration      time.Duration `mapstructure:"SESSION_DURATION"`
	CorsDuration         time.Duration `mapstructure:"CORS_DURATION"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	TokenKey             string        `mapstructure:"REFRESH_TOKEN_DURATION"`
	TokenSecretKey       string        `mapstructure:"TOKEN_SECRET_KEY"`
	SessionSecretKey1    []byte        `mapstructure:"SESSION_SECRET_KEY1"`
	SessionSecretKey2    []byte        `mapstructure:"SESSION_SECRET_KEY2"`
}

func StringToByteSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t.Kind() != reflect.Slice {
			return data, nil
		}
		if t.Elem().Kind() != reflect.Uint8 {
			return []byte(data.(string)), nil
		}
		return data, nil
	}
}

func LoadConfig(path string) (config *Config, err error) {
	viper.AddConfigPath(path)
	// 配置文件名称
	viper.SetConfigName("app")
	// 配置文件格式
	viper.SetConfigType("env")
	// 环境变量中的值会覆盖配置文件中的同名值
	viper.AutomaticEnv()

	optDecode := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		StringToByteSliceHookFunc(),
	))

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config, optDecode)
	return
}
