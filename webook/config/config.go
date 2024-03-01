package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"path"
	"reflect"
	"runtime"
	"time"
)

type config struct {
	DB struct {
		MySQL struct {
			Url                string `yaml:"url"`
			MigrationUrl       string `yaml:"migrationUrl"`
			MigrationSourceUrl string `yaml:"migrationSourceUrl"`
		} `yaml:"mysql"`

		Mongo struct {
			Url string `yaml:"url"`
		} `yaml:"mongo"`
	} `yaml:"db"`

	Redis struct {
		Address string `yaml:"address"`
	} `yaml:"redis"`

	App struct {
		HttpServerAddress string `yaml:"httpServerAddress"`
	} `yaml:"app"`

	Duration struct {
		Session      time.Duration `yaml:"session"`
		Cors         time.Duration `yaml:"cors"`
		AccessToken  time.Duration `yaml:"accessToken"`
		RefreshToken time.Duration `yaml:"refreshToken"`
	} `yaml:"duration"`

	SecretKey struct {
		Token    string `yaml:"token"`
		Session1 []byte `yaml:"session1"`
		Session2 []byte `yaml:"session2"`
	} `yaml:"secretKey"`
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
	log.Printf("读取配置文件:%s/dev.yaml", path)
	// 配置文件名称
	viper.SetConfigName("dev")
	// 配置文件格式
	viper.SetConfigType("yaml")
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
	Config, err = loadConfig(path.Dir(filename))
	if err != nil {
		log.Fatalf("配置文件加载失败:%v", err)
	}
	log.Println("配置文件加载成功")
}
