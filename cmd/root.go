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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// root command flags
var (
	host      net.IP
	port      uint16
	genDoc    bool
	pdHost    net.IP
	pdPort    uint16
	ca        string
	sslCert   string
	sslKey    string
	ctlClient *http.Client
	schema    string
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
	docCmd.AddCommand(mvccRootCmd, schemaRootCmd, regionRootCmd, tableRootCmd, decoderCmd, newBase64decodeCmd, newEtcdCommand(), keyRangeCmd)
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

func httpGet(path string) (body []byte, status int, err error) {
	url := schema + "://" + host.String() + ":" + strconv.Itoa(int(port)) + "/" + path
	resp, err := ctlClient.Get(url)
	if err != nil {
		return
	}
	status = resp.StatusCode
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()
	body, err = ioutil.ReadAll(resp.Body)
	return
}

func httpPrint(path string) error {
	body, status, err := httpGet(path)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		// Print response body directly if status is not ok.
		fmt.Println(string(body))
		return nil
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(prettyJSON.String())
	return nil
}

const (
	hostFlagName   = "host"
	portFlagName   = "port"
	docFlagName    = "doc"
	pdHostFlagName = "pdhost"
	pdPortFlagName = "pdport"
	caName         = "ca"
	sslKeyName     = "ssl-key"
	sslCertName    = "ssl-cert"
)

func init() {
	rootCmd.AddCommand(mvccRootCmd, schemaRootCmd, regionRootCmd, tableRootCmd, newBase64decodeCmd, decoderCmd, logCmd, newEtcdCommand(), keyRangeCmd)

	rootCmd.PersistentFlags().IPVarP(&host, hostFlagName, "", net.ParseIP("127.0.0.1"), "TiDB server host")
	rootCmd.PersistentFlags().Uint16VarP(&port, portFlagName, "", 10080, "TiDB server port")
	rootCmd.PersistentFlags().IPVarP(&pdHost, pdHostFlagName, "", net.ParseIP("127.0.0.1"), "PD server host")
	rootCmd.PersistentFlags().Uint16VarP(&pdPort, pdPortFlagName, "", 2379, "PD server port")
	rootCmd.PersistentFlags().StringVarP(&ca, caName, "", "", "TLS CA path")
	rootCmd.PersistentFlags().StringVarP(&sslKey, sslKeyName, "", "", "TLS Key path")
	rootCmd.PersistentFlags().StringVarP(&sslCert, sslCertName, "", "", "TLS Cert path")
	rootCmd.Flags().BoolVar(&genDoc, docFlagName, false, "generate doc file")
	if err := rootCmd.Flags().MarkHidden(docFlagName); err != nil {
		fmt.Printf("can not mark hidden flag, flag %s is not found", docFlagName)
		return
	}
	cobra.OnInitialize(func() {
		tlsConfig, err := prepareTLSConfig()
		if err != nil {
			fmt.Printf("cannot setup tls: %v", err)
		}
		if tlsConfig != nil {
			schema = "https"
		} else {
			schema = "http"
		}
		ctlClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	})
}

func prepareTLSConfig() (tlsConfig *tls.Config, err error) {
	if len(ca) != 0 {
		tlsConfig = &tls.Config{}
		certPool := x509.NewCertPool()
		// Create a certificate pool from the certificate authority
		var caBytes []byte
		caBytes, err = ioutil.ReadFile(ca)
		if err != nil {
			err = errors.Errorf("could not read ca certificate: %s", err)
			return
		}
		// Append the certificates from the CA
		if !certPool.AppendCertsFromPEM(caBytes) {
			err = errors.New("failed to append ca certs")
			return
		}
		tlsConfig.RootCAs = certPool
	}
	if len(sslCert) != 0 && len(sslKey) != 0 {
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		getCert := func() (*tls.Certificate, error) {
			// Load the client certificates from disk
			cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
			if err != nil {
				return nil, errors.Errorf("could not load client key pair: %s", err)
			}
			return &cert, nil
		}
		// pre-test cert's loading.
		if _, err = getCert(); err != nil {
			return
		}
		tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (certificate *tls.Certificate, err error) {
			return getCert()
		}
		tlsConfig.GetCertificate = func(info *tls.ClientHelloInfo) (certificate *tls.Certificate, err error) {
			return getCert()
		}
	}
	return
}
