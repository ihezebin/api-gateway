package cmd

import (
	"api-gateway/component/cache"
	"api-gateway/component/storage"
	"api-gateway/config"
	"api-gateway/domain/repository"
	"api-gateway/domain/service"
	"api-gateway/server"
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/ihezebin/oneness/logger"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	configPath string
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
			&cli.StringFlag{Destination: &configPath, Name: "config", Aliases: []string{"c"}, Value: "./config/config.toml", Usage: "config file path (default find file from pwd and exec dir"},
		},
		Before: func(c *cli.Context) error {
			if configPath == "" {
				return errors.New("config path is empty")
			}

			conf, err := config.Load(configPath)
			if err != nil {
				return errors.Wrapf(err, "load config error, path: %s", configPath)
			}

			if err = initComponents(ctx, conf); err != nil {
				return errors.Wrap(err, "init components error")
			}

			logger.Debugf(ctx, "component init success, config: %+v", *conf)

			return nil
		},
		Action: func(c *cli.Context) error {
			if err := server.Run(ctx, config.GetConfig().Port); err != nil {
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
			logger.WithCallerPrettyHook(),
			logger.WithTimestampHook(),
			logger.WithLevel(conf.Logger.Level),
			logger.WithLocalFsHook(filepath.Join(conf.Pwd, conf.Logger.Filename)),
		)
	}

	// init storage
	if conf.MongoDsn != "" {
		if err := storage.InitMongoStorageClient(ctx, conf.MongoDsn); err != nil {
			return errors.Wrap(err, "init mongo storage client error")
		}
	}

	// init cache
	cache.InitMemoryCache(time.Minute*5, time.Minute)
	if conf.Redis != nil {
		if err := cache.InitRedisCache(ctx, conf.Redis.Addr, conf.Redis.Password); err != nil {
			return errors.Wrap(err, "init redis cache client error")
		}
	}

	// init repository
	repository.Init()

	// init service
	service.Init()

	return nil
}