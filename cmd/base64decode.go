// Copyright 2018 PingCAP, Inc.
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
	"encoding/binary"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	inputValue string
)

// base64decodeCmd represents the base64decode command
var newBase64decodeCmd = &cobra.Command{
	Use:   "base64decode",
	Short: "decode base64 value",
	Long:  "decode base64 value to hex and uint64",
	RunE:  base64decodeCmd,
}

func base64decodeCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Only support one argument")
	}
	inputValue = args[0]
	uDec, err := base64Decode(inputValue)
	if err != nil {
		return err
	}
	if len(uDec) <= 8 {
		var num uint64
		fmt.Printf("hex: %s\n", uDec)
		err = binary.Read(bytes.NewBuffer([]byte(uDec)[0:8]), binary.BigEndian, &num)
		if err != nil {
			return err
		}
		fmt.Printf("uint64: %d\n", num)
	}
	return nil
}

func init() {
	newBase64decodeCmd.Flags().StringVarP(&inputValue, "value", "v", "", "the value you want decode")
}
