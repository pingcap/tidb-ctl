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
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

const (
	regionPrefix = "regions/"
)

// regionCmd represents the region command
var regionRootCmd = &cobra.Command{
	Use:   "region",
	Short: "Region information",
	Long: `tidb-ctl region [region id]
If no region id specified, it will return region info where
meta data located.`,
	RunE: getRegionInfo,
}

func getRegionInfo(_ *cobra.Command, args []string) error {
	switch len(args) {
	case 0:
		return httpPrint(regionPrefix + "meta")
	case 1:
		if _, err := strconv.ParseUint(args[0], 10, 64); err != nil {
			return err
		}
		return httpPrint(regionPrefix + args[0])
	default:
		return fmt.Errorf("too many arguments")
	}
}
