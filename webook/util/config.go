package util

import (
	"github.com/mitchellh/mapstructure"
	"log"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

type config struct {
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
	TokenKey             string        `mapstructure:"TOKEN_KEY"`
	TokenSecretKey       string        `mapstructure:"TOKEN_SECRET_KEY"`
	SessionSecretKey1    []byte        `mapstructure:"SESSION_SECRET_KEY1"`
	SessionSecretKey2    []byte        `mapstructure:"SESSION_SECRET_KEY2"`
}

func stringToByteSliceHookFunc() mapstructure.DecodeHookFunc {
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

func loadConfig(path string) (config *config, err error) {
	viper.AddConfigPath(path)
	log.Printf("读取配置文件:%s/app.env", path)
	// 配置文件名称
	viper.SetConfigName("app")
	// 配置文件格式
	viper.SetConfigType("env")
	// 环境变量中的值会覆盖配置文件中的同名值
	viper.AutomaticEnv()
	optDecode := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToByteSliceHookFunc(),
		),
	)

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config, optDecode)
	return
}

var Config *config

func init() {
	var err error
	// 以当前文件为基准的相对路径
	_, filename, _, _ := runtime.Caller(0)
	Config, err = loadConfig(path.Dir(path.Dir(filename)))
	if err != nil {
		log.Fatalf("配置文件加载失败:%v", err)
	}
	log.Println("配置文件加载成功")
}
