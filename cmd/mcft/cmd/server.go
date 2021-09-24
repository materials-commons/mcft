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
	"crypto/tls"
	"net/url"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
	"github.com/materials-commons/mcft/pkg/protocol"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start mcft as a server process",
	Long:  `Connect to remote Materials Commons server for uploads.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := mustReadApiKey()
		// Websocket connection defaults to wss, but can be overridden. Useful for local testing.
		wsScheme := os.Getenv("MC_WS_SCHEME")
		if wsScheme == "" {
			wsScheme = "wss"
		}

		u := url.URL{Scheme: wsScheme, Host: serverAddress, Path: "/ws"}
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatalf("Unable to connect to %s: %s", u.String(), err)
		}
		defer c.Close()

		if !authenticate(c, apiKey) {
			log.Fatalf("Unable to authenticate")
		}

		var reqType protocol.IncomingRequestType
		reqType.RequestType = protocol.ServerConnectRequestType
		if err := c.WriteJSON(reqType); err != nil {
			log.Fatalf("Unable to initiate server connection: %s", err)
		}

		// Write server connect here

		for {
			time.Sleep(3 * time.Second)
			// read to see if we are supposed to do anything
			// send ping
			// if there was something from read then do it
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringVarP(&serverAddress, "server-address", "s", "materialscommons.org", "Server to connect to")
}
