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
	regionPrefix     = "regions/"
	metaFlagName     = "meta"
	regionIDFlagName = "rid"
)

var (
	isMeta   bool
	regionID uint64
)

// regionCmd represents the region command
var regionRootCmd = &cobra.Command{
	Use:   "region",
	Short: "Region information",
	Long: `tidb-ctl region --meta(-m) | --rid(-i) [region id]
* --meta will return region info where meta data located
* --rid will return region info by region id`,
	RunE: getRegionInfo,
}

func init() {
	regionRootCmd.Flags().BoolVarP(&isMeta, metaFlagName, "m", false, "region info where meta data located")
	regionRootCmd.Flags().Uint64VarP(&regionID, regionIDFlagName, "i", 0, "region id")
}

func getRegionInfo(c *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	if c.Flag(metaFlagName).Changed && c.Flag(regionIDFlagName).Changed {
		return fmt.Errorf("%s and %s can not be set simultaneously", metaFlagName, regionIDFlagName)
	}
	if c.Flag(metaFlagName).Changed {
		return httpPrint(regionPrefix + "meta")
	}
	if c.Flag(regionIDFlagName).Changed {
		return httpPrint(regionPrefix + strconv.FormatUint(regionID, 10))
	}
	return c.Usage()
}
