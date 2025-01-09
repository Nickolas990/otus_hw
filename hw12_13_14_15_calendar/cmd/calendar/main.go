package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/app"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	//nolint:depguard
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	//nolint:depguard
	internalhttp "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/server/http"
	//nolint:depguard
	storage2 "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	//nolint:depguard
	memorystorage "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage/memory"
	//nolint:depguard
	sqlstorage "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage/sql"
	//nolint:depguard
	"github.com/spf13/viper"
)

var (
	configFile string
	storage    storage2.Storage
)

func init() {
	flag.StringVar(&configFile, "config", "configs/sample_config.yml", "Path to configuration file")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
		cancel()
	}
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Printf("unable to decode into struct, %v", err)
		cancel()
	}

	logg := logger.New(cfg.Logger.Level)
	log.Printf("Loaded configuration: %+v\n", cfg)

	if cfg.StorageType == "memory" {
		storage = memorystorage.New(logg)
	} else if cfg.StorageType == "db" {
		storage = sqlstorage.New(logg)
		err := storage.Connect(ctx, cfg)
		if err != nil {
			logg.Error(err.Error())
			return
		}

		defer func(storage storage2.Storage, ctx context.Context) {
			err := storage.Close(ctx)
			if err != nil {
				logg.Error(err.Error())
			}
		}(storage, ctx)
	}
	calendar := app.New(logg, storage)

	address := cfg.HTTPS.Host + ":" + cfg.HTTPS.Port

	server := internalhttp.NewServer(logg, calendar, address)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
	}
}
