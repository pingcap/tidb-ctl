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

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// root command flags
var (
	host   net.IP
	port   uint16
	genDoc bool
	pdHost net.IP
	pdPort uint16
)

const (
	rootUse       = "tidb-ctl"
	rootShort     = "TiDB Controller"
	rootLong      = "TiDB Controller (tidb-ctl) is a command line tool for TiDB Server (tidb-server)."
	dbFlagName    = "database"
	tableFlagName = "table"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   rootUse,
	Short: rootShort,
	Long:  rootLong,
	RunE:  genDocument,
}

func genDocument(c *cobra.Command, args []string) error {
	if !genDoc || len(args) != 0 {
		return c.Usage()
	}
	docDir := "./doc"
	docCmd := &cobra.Command{
		Use:   rootUse,
		Short: rootShort,
		Long:  rootLong,
	}
	docCmd.AddCommand(mvccRootCmd, schemaRootCmd, regionRootCmd, tableRootCmd, decoderCmd, newBase64decodeCmd)
	fmt.Println("Generating documents...")
	if err := doc.GenMarkdownTree(docCmd, docDir); err != nil {
		return err
	}
	fmt.Println("Done!")
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

func httpPrint(path string) error {
	url := "http://" + host.String() + ":" + strconv.Itoa(int(port)) + "/" + path
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		// Print response body directly if status is not ok.
		fmt.Println(string(body))
		return nil
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(prettyJSON.Bytes()))
	return nil
}

func init() {
	hostFlagName := "host"
	portFlagName := "port"
	docFlagName := "doc"

	rootCmd.AddCommand(mvccRootCmd, schemaRootCmd, regionRootCmd, tableRootCmd, newBase64decodeCmd, decoderCmd, logCmd)

	rootCmd.PersistentFlags().IPVarP(&host, hostFlagName, "", net.ParseIP("127.0.0.1"), "TiDB server host")
	rootCmd.PersistentFlags().Uint16VarP(&port, portFlagName, "", 10080, "TiDB server port")
	rootCmd.Flags().BoolVar(&genDoc, docFlagName, false, "generate doc file")
	if err := rootCmd.Flags().MarkHidden(docFlagName); err != nil {
		fmt.Printf("can not mark hidden flag, flag %s is not found", docFlagName)
		return
	}
}
