package main

import (
	"context"
	"flag"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	storage2 "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/app"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	internalhttp "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage/sql"
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
		log.Fatalf("unable to decode into struct, %v", err)
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

	address := cfg.HTTPServer.Host + ":" + cfg.HTTPServer.Port

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
		os.Exit(1)
	}
}
