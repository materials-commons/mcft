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
	"path/filepath"

	"github.com/gorilla/websocket"
	"github.com/materials-commons/mcft/pkg/protocol"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/apex/log"
)

var (
	cfgFile  string
	upgrader = websocket.Upgrader{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcftservd",
	Short: "Upload/download file server",
	Long:  `Handles upload and download file requests for materials commons from the mcft client.`,
	Run: func(cmd *cobra.Command, args []string) {
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true
		//e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		//e.GET("/ws", handleFileRequests)
		e.GET("/ws", handleUploadDownloadConnection)

		e.Logger.Fatal(e.Start(":1323"))
	},
}

func handleUploadDownloadConnection(c echo.Context) error {
	basePath := "/home/gtarcea/uploads"
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	var command protocol.CommandMsg
	var uploadMsg protocol.UploadMsg
	var fileBlockMsg protocol.FileBlockMsg
	var f *os.File

	defer func() {
		_ = ws.Close()
		if f != nil {
			_ = f.Close()
		}
	}()

	for {

		if err := ws.ReadJSON(&command); err != nil {
			//log.Errorf("Failed reading the command: %s", err)
			break
		}

		switch command.MsgType {
		case protocol.Upload:
			if err := ws.ReadJSON(&uploadMsg); err != nil {
				log.Errorf("Expected upload msg, got err: %s", err)
				return err
			}
			fullPath := filepath.Join(basePath, uploadMsg.Path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0770); err != nil {
				log.Errorf("Unable to create directory: %s", err)
				return err
			}
			f, err = os.Create(fullPath)
			if err != nil {
				log.Errorf("Unable to create file: %s", err)
				return err
			}
			break
		case protocol.FileBlock:
			if err := ws.ReadJSON(&fileBlockMsg); err != nil {
				log.Errorf("Expected FileBlock msg, got err: %s", err)
				return err
			}

			n, err := f.Write(fileBlockMsg.Block)
			if err != nil {
				log.Errorf("Failed writing to file: %s", err)
				break
			}

			if n != len(fileBlockMsg.Block) {
				log.Errorf("Did not write all of block, wrote %d, length %d", n, len(fileBlockMsg.Block))
			}
			break
		}
	}

	return nil
}

func handleFileRequests(c echo.Context) error {
	//basePath := "/home/gtarcea/uploads"
	fileMap := make(map[string]*os.File)
	_ = fileMap
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	f, err := os.Create("/tmp/mcft.out")
	if err != nil {
		log.Errorf("Unable to create file /tmp/mcft.out")
	}
	defer f.Close()

	for {
		//var command protocol.CommandMsg
		var fb protocol.FileBlockMsg
		if err := ws.ReadJSON(&fb); err != nil {
			break
		}
		n, err := f.Write(fb.Block)
		if err != nil {
			log.Errorf("Failed writing to file: %s", err)
			break
		}

		if n != len(fb.Block) {
			log.Errorf("Did not write all of block, wrote %d, length %d", n, len(fb.Block))
		}
		//switch command.MsgType {
		//case protocol.Login:
		//	break
		//case protocol.SetProject:
		//	break
		//case protocol.SendStat:
		//	break
		//case protocol.SendChecksum:
		//	break
		//case protocol.StatInfo:
		//	break
		//case protocol.ChecksumInfo:
		//	break
		//case protocol.SetPosition:
		//	break
		//case protocol.Upload:
		//	break
		//case protocol.FileBlock:
		//	break
		//case protocol.FinishUpload:
		//	break
		//case protocol.Download:
		//	break
		//default:
		//	log.Errorf("Unknown message type: %d", command.MsgType)
		//}
	}

	return nil
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mcftservd.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mcftservd" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".mcftservd")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
