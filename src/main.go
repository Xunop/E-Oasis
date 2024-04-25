package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/store/db"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/server"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/worker"
	"github.com/spf13/cobra"
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

// Because the log need to be initialized before the command is executed, so can't use log package here
var (
	dsn        string
	host       string
	port       int
	data       string
	cfgFile    string
	debug      bool
	configDump bool

	rootCmd = &cobra.Command{
		Use:   "e-oasis",
		Short: "E-Oasis is a e-book management system",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			ctx, cancle := context.WithCancel(context.Background())
			defer cancle()

			// Will create a sqlite database
			db, err := db.NewDB()
			if err != nil {
				cancle()
				fmt.Println("Error connecting to database", err)
				return
			}
			defer db.Close()
			if err := db.Migrate(ctx); err != nil {
				cancle()
				fmt.Println("Error migrating database,", err)
			}

			store := store.NewStore(db.DB)
			if err := store.Ping(); err != nil {
				cancle()
				fmt.Println("Error pinging database", err)
				return
			}

			pool := worker.NewPool(store, config.Opts.WorkerPoolSize)

			// Start Server
			s, err := server.StartServer(ctx, store, pool)
			if err != nil {
				cancle()
				fmt.Println("Error creating server", err)
				return
			}

			if config.Opts.MetricsCollector {
				// TODO: Add metrics
			}

			go func() {
				<-c
				fmt.Println("Received interrupt signal")
				s.Shutdown(ctx)
				log.Logger.Sync()
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

	config.GetConfig()

	rootCmd.PersistentFlags().StringVarP(&dsn, "dsn", "", "", "Database connection string")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "", "localhost", "Server host")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "Server port")
	rootCmd.PersistentFlags().StringVarP(&data, "data", "d", "data", "Data directory")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Debug mode")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file")

	//TODO: Add help flag and dump config file
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Help")
	rootCmd.PersistentFlags().BoolVarP(&configDump, "config-dump", "", false, "Dump config file")

	// viper.SetEnvPrefix("eoasis")
}

// initConfig will run befroe the command is exeuted
// init() -> Getconfig() -> initConfig()
func initConfig() {
	fmt.Println("Initializing config")
	// viper.AutomaticEnv()

	var err error
	if cfgFile != "" {
		config.Opts, err = config.ParseFile(cfgFile)
		if err != nil {
			fmt.Println("Error parsing config file", err)
			panic(err)
		}
	} else {
		if dsn != "" {
			config.Opts.DSN = dsn
		}
		if host != "" {
			config.Opts.Host = host
		}
		if port != 0 {
			config.Opts.Port = port
		}
		if data != "" {
			config.Opts.Data = data
		}
		if debug {
			fmt.Println("Debug mode enabled")
			config.Opts.LogLevel = "debug"
		}
		if configDump {
			// TODO: config impl stringer
		}
	}

	// Initialize logger
	log.Logger = log.NewLogger()

	config := config.Opts
	fmt.Printf(`---
		Server config
		host: %s
		port: %d
		db: %s
 		log_level: %s
 		data: %s
---
	`, config.Host, config.Port, config.DSN, config.LogLevel, config.Data)
}

func main() {
	if err := Execute(); err != nil {
		panic(err)
	}
}

func printGreetings() {
	print(greetingBanner)
}
