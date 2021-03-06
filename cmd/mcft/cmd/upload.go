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
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/gorilla/websocket"
	"github.com/materials-commons/mcft/pkg/protocol"
	"github.com/saracen/walker"
	"github.com/spf13/cobra"
)

var (
	uploadTo      string
	serverAddress string
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:     "upload <files-or-directories>",
	Aliases: []string{"up"},
	Short:   "Upload files/directories to Materials Commons",
	Long:    `Upload files/directories to Materials Commons`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		apiKey := mustReadApiKey()

		for _, fileOrDirPath := range args {
			basePath, _ := filepath.Abs(fileOrDirPath)
			basePath = filepath.Dir(basePath)
			fi, err := os.Stat(fileOrDirPath)
			if err != nil {
				log.Errorf("Unable to read %s, skipping...", err)
				continue
			}

			if fi.IsDir() {
				// walk function called for every path found
				walkFn := func(pathname string, fi os.FileInfo) error {
					if !fi.Mode().IsRegular() {
						return nil
					}

					if !strings.HasPrefix(pathname, "/") {
						pathname, _ = filepath.Abs(pathname)
					}
					uploadPath := filepath.Join("/", strings.Replace(pathname, basePath, uploadTo, 1))
					if err := uploadFile(pathname, uploadPath, apiKey); err != nil {
						log.Errorf("Upload failed for %s: %s", pathname, err)
					}

					return nil
				}

				// error function called for every error encountered
				errorCallbackOption := walker.WithErrorCallback(func(pathname string, err error) error {
					// ignore permission errors
					if os.IsPermission(err) {
						return nil
					}
					// halt traversal on any other error
					return err
				})

				_ = walker.Walk(fileOrDirPath, walkFn, errorCallbackOption)
			} else {
				if !strings.HasPrefix(fileOrDirPath, "/") {
					fileOrDirPath, _ = filepath.Abs(fileOrDirPath)
				}

				uploadPath := filepath.Join("/", strings.Replace(fileOrDirPath, basePath, uploadTo, 1))

				fi, err := os.Stat(fileOrDirPath)
				if err != nil {
					log.Errorf("Unable to Stat(%s): %s", fileOrDirPath, err)
					continue
				}

				if !fi.Mode().IsRegular() {
					continue
				}

				fmt.Printf("Uploading file: %s to %s\n\n", fileOrDirPath, uploadPath)
				if err := uploadFile(fileOrDirPath, uploadPath, apiKey); err != nil {
					log.Errorf("Upload failed for %s: %s", fileOrDirPath, err)
				}
			}
		}
	},
}

func uploadFile(pathToFile, uploadToPath, apiKey string) error {
	u := url.URL{Scheme: "ws", Host: serverAddress, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Unable to connect to %s: %s", u.String(), err)
	}
	defer c.Close()

	f, err := os.Open(pathToFile)

	if err != nil {
		log.Fatalf("Unable to open %s: %s", pathToFile, err)
	}
	defer f.Close()

	var incomingReq protocol.IncomingRequestType

	if !authenticate(c, apiKey) {
		log.Fatalf("Unable to authenticate")
	}

	incomingReq.RequestType = protocol.UploadFileReq
	if err := c.WriteJSON(incomingReq); err != nil {
		//log.Errorf("Unable to initiate upload: %s", err)
		return err
	}

	// First send notice of upload
	uploadMsg := protocol.UploadFileRequest{
		Path: uploadToPath,
	}

	if err := c.WriteJSON(uploadMsg); err != nil {
		//log.Errorf("Unable to initiate upload: %s", err)
		return err
	}

	data := make([]byte, 32*1024)
	fb := protocol.FileBlockRequest{}
	for {

		n, err := f.Read(data)
		if err != nil {
			if err != io.EOF {
				//log.Errorf("Read returned error: %s", err)
				return err
			}
			break
		}

		incomingReq.RequestType = protocol.FileBlockReq
		if err := c.WriteJSON(incomingReq); err != nil {
			//log.Errorf("Unable to initiate upload: %s", err)
			return err
		}

		fb.Block = data[:n]
		if err := c.WriteJSON(fb); err != nil {
			//log.Errorf("WriteJSON failed: %s", err)
			return err
		}
	}

	return nil
}

func authenticate(c *websocket.Conn, key string) bool {
	var req protocol.IncomingRequestType
	req.RequestType = protocol.AuthenticateReq
	if err := c.WriteJSON(req); err != nil {
		return false
	}

	auth := protocol.AuthenticateRequest{
		APIToken: key,
	}

	if err := c.WriteJSON(auth); err != nil {
		return false
	}

	return true
}

func mustReadApiKey() string {
	u, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to identify user: %s", err)
	}

	configPath := filepath.Join(u.HomeDir, ".materialscommmons", "config.json")
	contents, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Unable to read %s: %s", configPath, err)
	}

	var config struct {
		APIKey string `json:"api_key"`
	}

	if err := json.Unmarshal(contents, &config); err != nil {
		log.Fatalf("Unable to parse (%s): %s", configPath, err)
	}

	return config.APIKey
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.PersistentFlags().StringVarP(&uploadTo, "upload-to", "t", "", "Path to upload to in project")
	uploadCmd.PersistentFlags().StringVarP(&serverAddress, "server-address", "s", "materialscommons.org", "Server to connect to")
}
