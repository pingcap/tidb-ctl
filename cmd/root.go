// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	urlHeader           = "http://"
	hostFlagName        = "host"
	portFlagName        = "port"
	defaultHost         = "127.0.0.1"
	defaultPort  uint16 = 10080
)

var (
	host    net.IP
	port    uint16
	baseURL string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tidb-ctl",
	Short: "TiDB Controller",
	Long:  `TiDB Controller (tidb-ctl) is a command line tool for TiDB Server (tidb-server).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func replaceTableFlag(table string) string {
	return strings.Replace(table, ".", "/", 1)
}

func httpPrint(path string) {
	baseURL = urlHeader + host.String() + ":" + strconv.Itoa(int(port)) + "/"
	resp, err := http.Get(baseURL + path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "    ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(prettyJSON.Bytes()))
}

func init() {
	rootCmd.PersistentFlags().IPVar(&host, hostFlagName, net.IP(defaultHost), "TiDB server host")
	rootCmd.PersistentFlags().Uint16Var(&port, portFlagName, defaultPort, "TiDB server port")
	rootCmd.MarkPersistentFlagRequired(hostFlagName)
	rootCmd.MarkPersistentFlagRequired(portFlagName)
}
