package main

import (
	"context"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/database"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/server"
	"github.com/spf13/cobra"
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
	databaseURL    string
	host           string
	port           string
	data           string
	configInstance *config.Config

	rootCmd = &cobra.Command{
		Use:   "e-oasis",
		Short: "E-Oasis is a e-book management system",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancle := context.WithCancel(context.Background())
			defer cancle()

			db, err := database.NewDB(configInstance)
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

			store := store.NewStore(db, configInstance)
			if err := store.Ping(); err != nil {
				cancle()
				log.Error("Error pinging database", zap.Error(err))
				return
			}

			s, err := server.NewServer(ctx, store)
			if err != nil {
				cancle()
				log.Error("Error creating server", zap.Error(err))
				return
			}

		},
	}
)

func main() {
	//StartServer()
	log.Info("Server started")
	defer log.Logger.Sync()
}
