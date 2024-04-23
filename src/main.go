package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/database"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/server"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/version"
	"github.com/Xunop/e-oasis/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	greetingBanner = `
███████        ██████   █████  ███████ ██ ███████ 
██            ██    ██ ██   ██ ██      ██ ██      
█████   █████ ██    ██ ███████ ███████ ██ ███████ 
██            ██    ██ ██   ██      ██ ██      ██ 
███████        ██████  ██   ██ ███████ ██ ███████ 
`
)

var (
	dsn     string
	host    string
	port    string
	data    string
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "e-oasis",
		Short: "E-Oasis is a e-book management system",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			ctx, cancle := context.WithCancel(context.Background())
			defer cancle()

			db, err := database.NewDB()
			if err != nil {
				cancle()
				log.Error("Error connecting to database", zap.Error(err))
				return
			}
			defer db.Close()
			if err := database.Migrate(db, ctx); err != nil {
				cancle()
				log.Error("Error migrating database", zap.Error(err))
			}

			store := store.NewStore(db)
			if err := store.Ping(); err != nil {
				cancle()
				log.Error("Error pinging database", zap.Error(err))
				return
			}

			pool := worker.NewPool(store, config.Opts.WorkerPoolSize)

			s, err := server.StartServer(ctx, store, pool)
			if err != nil {
				cancle()
				log.Error("Error creating server", zap.Error(err))
				return
			}

			if config.Opts.MetricsCollector {
				// TODO: Add metrics
			}

			go func() {
				<-c
				log.Debug("Received interrupt signal", zap.String("signal", "SIGINT"))
				s.Shutdown(ctx)
				cancle()
			}()

			printGreetings()

			// Waitting for signal
			<-ctx.Done()
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&dsn, "dsn", "", "", "Database connection string")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "", "localhost", "Server host")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Server port")
	rootCmd.PersistentFlags().StringVarP(&data, "data", "d", "data", "Data directory")
	rootCmd.PersistentFlags().BoolP("debug", "", false, "Debug mode")

	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version")
	rootCmd.PersistentFlags().BoolP("config-dump", "", false, "Dump config file")
	//TODO: Add help flag and dump config file
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Help")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file")

	err := viper.BindPFlag("dsn", rootCmd.PersistentFlags().Lookup("dsn"))
	if err != nil {
		log.Error("Error binding dsn flag", zap.Error(err))
		panic(err)
	}

	err = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	if err != nil {
		log.Error("Error binding host flag", zap.Error(err))
		panic(err)
	}

	err = viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		log.Error("Error binding port flag", zap.Error(err))
		panic(err)
	}

	err = viper.BindPFlag("data", rootCmd.PersistentFlags().Lookup("data"))
	if err != nil {
		log.Error("Error binding data flag", zap.Error(err))
		panic(err)
	}

	if rootCmd.PersistentFlags().Lookup("version").Changed {
		fmt.Println(version.GetCurrentVersion())
	}

	if rootCmd.PersistentFlags().Lookup("debug").Changed {
		config.Opts.LogLevel = "debug"
	}

	if rootCmd.PersistentFlags().Lookup("config").Value.String() != "" {
		config.Opts, err = config.ParseFile(rootCmd.PersistentFlags().Lookup("config").Value.String())
		if err != nil {
			log.Error("Error parsing config file", zap.Error(err))
			panic(err)
		}
	}
	viper.SetEnvPrefix("eoasis")
}

func initConfig() {
	viper.AutomaticEnv()
	config := config.Opts

	if cfgFile != "" {

	}

	fmt.Printf(`---
		Server config
		version: %s
		host: %s
		port: %d
		db: %s
 		log_level: %s
 		data: %s
---
	`, config.Version, config.Host, config.Port, config.DSN, config.LogLevel, config.Data)
}

func main() {
	//StartServer()
	log.Info("Server started")
	defer log.Logger.Sync()
	if err := Execute(); err != nil {
		panic(err)
	}
}

func printGreetings() {
	print(greetingBanner)
}
