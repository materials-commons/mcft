// Copyright © 2021 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	mcdb "github.com/materials-commons/gomcdb"
	"github.com/materials-commons/mcft/pkg/ft"
	"github.com/spf13/cobra"
	"github.com/subosito/gotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/apex/log"
)

var (
	cfgFile    string
	mcfsDir    string
	dotenvPath string
	db         *gorm.DB
	upgrader   = websocket.Upgrader{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcftservd",
	Short: "Upload/download file server",
	Long:  `Handles upload and download file requests for materials commons from the mcft client.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}

		if db, err = gorm.Open(mysql.Open(mcdb.MakeDSNFromEnv()), gormConfig); err != nil {
			log.Fatalf("Failed to open db (%s): %s", mcdb.MakeDSNFromEnv(), err)
		}

		e := echo.New()
		e.HideBanner = true
		e.HidePort = true
		e.Use(middleware.Recover())
		e.GET("/ws", handleUploadDownloadConnection)

		showEnv()

		e.Logger.Fatal(e.Start(":1423"))
	},
}

func handleUploadDownloadConnection(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	fileTransferHandler := ft.NewFileTransferHandler(ws, db)
	defer func() {
		_ = ws.Close()
	}()

	if err := fileTransferHandler.Run(); err != nil {
		status := ft.Error2Status(err)
		_ = ws.WriteJSON(status)
	}

	return nil
}

func showEnv() {
	fmt.Printf("MCFS_ROOT = '%s'\n", ft.GetMCFSRoot())
	fmt.Printf("DSN = '%s'\n", mcdb.MakeDSNFromEnv())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if err := gotenv.Load(MustGetDotenvPath()); err != nil {
		log.Fatalf("Loading dotenv file path %s failed: %s", dotenvPath, err)
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mcftservd.yaml)")

	mcfsDir = os.Getenv("MCFS_DIR")
	if mcfsDir == "" {
		log.Fatalf("MCFS_DIR environment variable not set")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := gotenv.Load(MustGetDotenvPath()); err != nil {
		log.Fatalf("Loading dotenv file path %s failed: %s", dotenvPath, err)
	}
}

func MustGetDotenvPath() string {
	if dotenvPath != "" {
		return dotenvPath
	}

	dotenvPath = os.Getenv("MC_DOTENV_PATH")
	if dotenvPath == "" {
		log.Fatal("MC_DOTENV_PATH not set")
	}

	return dotenvPath
}
