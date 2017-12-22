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
	"strings"

	"github.com/spf13/cobra"
)

const (
	txnPrefix = "mvcc/txn/"
	keyPrefix = "mvcc/key/"
	hexPrefix = "mvcc/hex/"
	idxPrefix = "mvcc/index/"
)

// mvcc command flags
var (
	mvccDB          string
	mvccTable       string
	mvccHID         int64
	mvccStartTS     uint64
	mvccIndexName   string
	mvccIndexValues string
)

// mvccCmd represents the mvcc command
var mvccRootCmd = &cobra.Command{
	Use:   "mvcc",
	Short: "MVCC Information",
	Long:  "Get for MVCC information",
}

func init() {
	handleFlagName := "hid"
	startTSFlagName := "start-ts"
	indexNameFlagName := "name"
	indexValueFlagName := "values"

	mvccRootCmd.AddCommand(keyCmd, txnCmd, hexCmd, idxCmd)

	keyCmd.Flags().StringVarP(&mvccDB, dbFlagName, "d", "", "database name")
	keyCmd.Flags().StringVarP(&mvccTable, tableFlagName, "t", "", "table name")
	keyCmd.Flags().Int64VarP(&mvccHID, handleFlagName, "i", 0, "get MVCC info of the key with a specified handle ID.")
	if err := keyCmd.MarkFlagRequired(dbFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", dbFlagName)
		return
	}
	if err := keyCmd.MarkFlagRequired(tableFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", tableFlagName)
		return
	}
	if err := keyCmd.MarkFlagRequired(handleFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", handleFlagName)
		return
	}

	txnCmd.Flags().StringVarP(&mvccDB, dbFlagName, "d", "", "database name")
	txnCmd.Flags().StringVarP(&mvccTable, tableFlagName, "t", "", "table name")
	txnCmd.Flags().Uint64VarP(&mvccStartTS, startTSFlagName, "s", 0,
		"get MVCC info of the primary key, or get MVCC info of the first key in the table (with --table) with a specified start ts.")
	if err := txnCmd.MarkFlagRequired(startTSFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", startTSFlagName)
		return
	}

	idxCmd.Flags().StringVarP(&mvccIndexName, indexNameFlagName, "n", "", "index name of a specified index key.")
	idxCmd.Flags().StringVarP(&mvccDB, dbFlagName, "d", "", "database name")
	idxCmd.Flags().StringVarP(&mvccTable, tableFlagName, "t", "", "table name")
	idxCmd.Flags().Int64VarP(&mvccHID, handleFlagName, "i", 0, "get MVCC info of the key with a specified handle ID.")
	idxCmd.Flags().StringVarP(&mvccIndexValues, indexValueFlagName, "v", "",
		"get MVCC info of a specified index key, argument example: `column_name_1=column_value_1,column_name_2=column_value2...`")
	if err := idxCmd.MarkFlagRequired(indexNameFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", indexNameFlagName)
		return
	}
	if err := idxCmd.MarkFlagRequired(dbFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", dbFlagName)
		return
	}
	if err := idxCmd.MarkFlagRequired(tableFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", tableFlagName)
		return
	}
	if err := idxCmd.MarkFlagRequired(handleFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", handleFlagName)
		return
	}
	if err := idxCmd.MarkFlagRequired(indexValueFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", indexValueFlagName)
		return
	}
}

// keyCmd represents the mvcc by key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "MVCC Information of table record key",
	Long:  "tidb-ctl mvcc key --database(-d) [database name] --table(-t) [table name] --hid(-i) [handle]",
	RunE:  mvccKeyQuery,
}

func mvccKeyQuery(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	return httpPrint(keyPrefix + mvccDB + "/" + mvccTable + "/" + strconv.FormatInt(mvccHID, 10))
}

// txnCmd represents the mvcc by transaction command
var txnCmd = &cobra.Command{
	Use:   "txn",
	Short: "MVCC Information of transaction",
	Long:  "tidb-ctl mvcc txn --start-ts(-s) [start timestamp] --database(-d) [database name] --table(-t) [table name]",
	RunE:  mvccTxnQuery,
}

func mvccTxnQuery(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}

	if len(mvccDB) > 0 && len(mvccTable) > 0 {
		return httpPrint(txnPrefix + strconv.FormatUint(mvccStartTS, 10) + "/" + mvccDB + "/" + mvccTable)
	} else if len(mvccDB) == 0 && len(mvccTable) == 0 {
		return httpPrint(txnPrefix + strconv.FormatUint(mvccStartTS, 10))
	}
	return fmt.Errorf("wrong arguments, database name and table name should be set simultaneously")
}

// hexCmd represents the mvcc by hex command
var hexCmd = &cobra.Command{
	Use:   "hex",
	Short: "MVCC Information by a hex value",
	Long:  "tidb-ctl mvcc hex [hex value]",
	RunE:  mvccHexQuery,
}

func mvccHexQuery(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need a hex value")
	}
	return httpPrint(hexPrefix + args[0])
}

// idxCmd represents the mvcc by index value command
var idxCmd = &cobra.Command{
	Use:   "index",
	Short: "MVCC Information of index record key",
	Long: `tidb-ctl mvcc index --database(-d) [database name] --table(-t) [table name] --hid(-i) [handle] [index values]

	index values should be like "column_name_1:column_value_1,column_name_2:column_value2..."`,
	RunE: mvccIdxQuery,
}

func mvccIdxQuery(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	queryPrefix := idxPrefix + mvccDB + "/" + mvccTable + "/" + mvccIndexName + "/" + strconv.FormatInt(mvccHID, 10) + "?"
	return httpPrint(queryPrefix + strings.Replace(mvccIndexValues, ",", "&", -1))
}
