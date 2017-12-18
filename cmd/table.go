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

	"github.com/spf13/cobra"
)

const (
	tablePrefix  = "tables/"
	regionSuffix = "regions"
	usageSurffix = "disk-usage"
)

var (
	tableDB    string
	tableTable string
)

// tableCmd represents the table command
var tableRootCmd = &cobra.Command{
	Use:   "table",
	Short: "Table information",
	Long:  `tidb-ctl table`,
}

func init() {
	tableRootCmd.AddCommand(regionCmd, diskUsageCmd)
	tableRootCmd.PersistentFlags().StringVarP(&tableDB, dbFlagName, "d", "", "database name")
	tableRootCmd.PersistentFlags().StringVarP(&tableTable, tableFlagName, "t", "", "table name")
	if err := tableRootCmd.MarkPersistentFlagRequired(dbFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", dbFlagName)
		return
	}
	if err := tableRootCmd.MarkPersistentFlagRequired(tableFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", tableFlagName)
		return
	}
}

var regionCmd = &cobra.Command{
	Use:   regionSuffix,
	Short: "region info of table",
	Long:  "tidb-ctl table region --database(-d) [database name] --table(-t) [table name]",
	RunE:  getTableRegion,
}

func getTableRegion(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	return httpPrint(tablePrefix + tableDB + "/" + tableTable + "/" + regionSuffix)
}

var diskUsageCmd = &cobra.Command{
	Use:   usageSurffix,
	Short: "disk usage of table",
	Long:  "tidb-ctl table disk-usage --database(-d) [database name] --table(-t) [table name]",
	RunE:  getTableDiskUsage,
}

func getTableDiskUsage(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	return httpPrint(tablePrefix + tableDB + "/" + tableTable + "/" + usageSurffix)
}
