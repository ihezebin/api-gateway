package cmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ihezebin/olympus"
	"github.com/ihezebin/olympus/logger"
	"github.com/ihezebin/olympus/runner"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"api-gateway/component/cache"
	"api-gateway/config"
	"api-gateway/server"
)

var (
	configPath string
	rulePaths  cli.StringSlice
)

func Run(ctx context.Context) error {

	app := &cli.App{
		Name:    "go-template-ddd",
		Version: "v1.0.1",
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
			httpServer, err := server.NewServer(ctx, config.GetConfig())
			if err != nil {
				return errors.Wrap(err, "new http server err")
			}

			tasks := make([]runner.Task, 0)
			tasks = append(tasks, httpServer)

			runner.NewRunner(tasks...).Run(ctx)

			return nil
		},
	}

	return app.Run(os.Args)
}

func initComponents(ctx context.Context, conf *config.Config) error {
	// init logger
	if conf.Logger != nil {
		logger.ResetLoggerWithOptions(
			logger.WithLoggerType(logger.LoggerTypeZap),
			logger.WithServiceName(conf.ServiceName),
			logger.WithCaller(),
			logger.WithTimestamp(),
			logger.WithLevel(conf.Logger.Level),
			//logger.WithLocalFsHook(filepath.Join(conf.Pwd, conf.Logger.Filename)),
			logger.WithRotate(logger.RotateConfig{
				Path:               filepath.Join(conf.Pwd, conf.Logger.Filename),
				MaxSizeKB:          1024 * 500, // 500 MB
				MaxAge:             time.Hour * 24 * 7,
				MaxRetainFileCount: 3,
				Compress:           true,
			}),
		)
	}

	// init cache
	cache.InitMemoryCache(time.Minute*5, time.Minute)
	if conf.Redis != nil {
		if err := cache.InitRedisCache(ctx, conf.Redis.Addrs, conf.Redis.Password); err != nil {
			return errors.Wrap(err, "init redis cache client error")
		}
	}
	if conf.Redis != nil {
		if err := cache.InitRedisCache(ctx, conf.Redis.Addrs, conf.Redis.Password); err != nil {
			return errors.Wrap(err, "init redis cache client error")
		}
	}

	return nil
}
