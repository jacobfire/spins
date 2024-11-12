package main

import (
	"NodeArt/internal/config"
	"NodeArt/internal/db"
	"NodeArt/internal/server"
	"context"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

func main() {
	var configDir string

	flag.StringVar(&configDir, "c", "-c ./../etc/", "Path to override configuration")
	flag.Parse()

	if configDir == "" {
		log.Fatal("Specify config directory with -c flag.")
	}
	if err := config.Init(configDir); err != nil {
		fmt.Printf("load of configuration failed: %s", err.Error())
	}
	ctx := context.Background()
	DBConf := db.Config{
		Host:     viper.GetString("db.host"),
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.ssl_mode"),
	}
	conn, err := db.NewConnection(ctx, DBConf)
	if err != nil {
		log.Printf("db init failed: %s", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	appConfig := server.Config{
		Port: viper.GetString("app.port"),
		Conn: conn,
	}
	ws := server.New(appConfig)
	ws.Run()
}
