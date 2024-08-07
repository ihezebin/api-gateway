package cmd

import (
	"api-gateway/component/cache"
	"api-gateway/config"
	"api-gateway/server"
	"context"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ihezebin/oneness"
	"github.com/ihezebin/oneness/logger"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	configPath string
	rulePaths  cli.StringSlice
)

func Run(ctx context.Context) error {

	app := &cli.App{
		Name:    "go-template-ddd",
		Version: "v1.0.0",
		Usage:   "Rapid construction template of Web service based on DDD architecture",
		Authors: []*cli.Author{
			{Name: "hezebin", Email: "ihezebin@qq.com"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &configPath,
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "./config/config.toml",
				Usage:       "config file path (default find file from pwd and exec dir"},
			&cli.StringSliceFlag{
				Destination: &rulePaths, Name: "rule_path",
				Aliases: []string{"r"},
				Value:   cli.NewStringSlice("./config/global.toml", "./config/blog.toml", "./config/sso.toml"),
				Usage:   "config file path (default find file from pwd and exec dir"},
		},
		Before: func(c *cli.Context) error {
			if configPath == "" {
				return errors.New("config path is empty")
			}

			logger.Infof(ctx, "config path: %s, rule path: %s", configPath, rulePaths.Value())

			conf, err := config.Load(configPath, rulePaths.Value()...)
			if err != nil {
				return errors.Wrapf(err, "load config error, path: %s", configPath)
			}

			if err = initComponents(ctx, conf); err != nil {
				return errors.Wrap(err, "init components error")
			}

			logger.Debugf(ctx, "component init success, config: %s", conf.String())

			return nil
		},
		Action: func(c *cli.Context) error {
			if err := server.Run(ctx, config.GetConfig()); err != nil {
				logger.WithError(err).Fatalf(ctx, "server run error, port: %d", config.GetConfig().Port)
			}

			return nil
		},
	}

	return app.Run(os.Args)
}

func initComponents(ctx context.Context, conf *config.Config) error {
	// init logger
	if conf.Logger != nil {
		logger.ResetLoggerWithOptions(
			logger.WithServiceName(conf.ServiceName),
			logger.WithPrettyCallerHook(),
			logger.WithTimestampHook(),
			logger.WithLevel(conf.Logger.Level),
			//logger.WithLocalFsHook(filepath.Join(conf.Pwd, conf.Logger.Filename)),
			// 每天切割，保留 3 天的日志
			logger.WithRotateLogsHook(filepath.Join(conf.Pwd, conf.Logger.Filename), time.Hour*24, time.Hour*24*3),
		)
	}

	// init cache
	cache.InitMemoryCache(time.Minute*5, time.Minute)
	if conf.Redis != nil {
		if err := cache.InitRedisCache(ctx, conf.Redis.Addrs, conf.Redis.Password); err != nil {
			return errors.Wrap(err, "init redis cache client error")
		}
	}

	return nil
}
