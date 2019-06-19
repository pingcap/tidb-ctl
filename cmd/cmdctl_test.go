// Copyright 2019 PingCAP, Inc.
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
	"fmt"
	"net"
	"testing"

	. "github.com/pingcap/check"
	"github.com/spf13/cobra"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&cmdTestSuite{})

type cmdTestSuite struct{}

func (s *cmdTestSuite) TestBase64Decode(c *C) {
	cmd := initCommand()
	args := []string{"base64decode", "AAAAACqPhb0="}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "hex: 000000002a8f85bd\nuint64: 714048957\n")
}

func initCommand() *cobra.Command {
	hostFlagName := "host"
	portFlagName := "port"
	docFlagName := "doc"
	pdHostFlagName := "pdhost"
	pdPortFlagName := "pdport"
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(mvccRootCmd, schemaRootCmd, regionRootCmd, tableRootCmd, newBase64decodeCmd, decoderCmd, newEtcdCommand())

	rootCmd.PersistentFlags().IPVarP(&host, hostFlagName, "H", net.ParseIP("127.0.0.1"), "TiDB server host")
	rootCmd.PersistentFlags().Uint16VarP(&port, portFlagName, "P", 10080, "TiDB server port")
	rootCmd.PersistentFlags().IPVarP(&pdHost, pdHostFlagName, "i", net.ParseIP("127.0.0.1"), "PD server host")
	rootCmd.PersistentFlags().Uint16VarP(&pdPort, pdPortFlagName, "p", 2379, "PD server port")
	rootCmd.Flags().BoolVar(&genDoc, docFlagName, false, "generate doc file")
	if err := rootCmd.Flags().MarkHidden(docFlagName); err != nil {
		fmt.Printf("can not mark hidden flag, flag %s is not found", docFlagName)

	}
	return rootCmd
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output []byte, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()
	return c, buf.Bytes(), err
}
