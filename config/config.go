package config

import (
	"api-gateway/domain/entity"
	"os"

	"github.com/ihezebin/oneness/config"
	"github.com/ihezebin/oneness/logger"
	"github.com/pkg/errors"
)

type Config struct {
	ServiceName string             `json:"service_name" mapstructure:"service_name"`
	Port        uint               `json:"port" mapstructure:"port"`
	Pwd         string             `json:"-" mapstructure:"-"`
	Logger      *LoggerConfig      `json:"logger" mapstructure:"logger"`
	Redis       *RedisConfig       `json:"redis" mapstructure:"redis"`
	Endpoints   []*entity.Endpoint `json:"endpoints" mapstructure:"endpoints"`
	Rules       []*entity.Rule     `json:"-" mapstructure:"-"`
}

type RedisConfig struct {
	Addr     string `json:"addr" mapstructure:"addr"`
	Password string `json:"password" mapstructure:"password"`
}

type LoggerConfig struct {
	Level    logger.Level `json:"level" mapstructure:"level"`
	Filename string       `json:"filename" mapstructure:"filename"`
}

var gConfig *Config = &Config{}

func GetConfig() *Config {
	return gConfig
}

func Load(path string, rulePaths ...string) (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "get pwd error")
	}

	if err = config.NewWithFilePath(path).Load(gConfig); err != nil {
		return nil, errors.Wrap(err, "load config error")
	}
	gConfig.Pwd = pwd

	rules := make([]*entity.Rule, 0)
	for _, rulePath := range rulePaths {
		ruleConfig := new(entity.Rule)
		if err = config.NewWithFilePath(rulePath).Load(ruleConfig); err != nil {
			return nil, errors.Wrapf(err, "load rule config error, rule path: %s", rulePath)
		}
		rules = append(rules, ruleConfig)
	}
	gConfig.Rules = rules

	return gConfig, nil
}
