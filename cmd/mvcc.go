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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	txnPrefix   = "mvcc/txn/"
	keyPrefix   = "mvcc/key/"
	indexPrefix = "mvcc/index/"
)

var (
	mvccTable       string
	mvccHID         int64
	mvccStartTS     uint64
	mvccIndexName   string
	mvccIndexValues []string
)

// mvccCmd represents the mvcc command
var mvccCmd = &cobra.Command{
	Use:   "mvcc",
	Short: "MVCC Information Query",
	Long: `Query for MVCC information, e.g.
	* tidb-ctl mvcc -t mydb.mytable --hid 123
	MVCC info of a specified handle in mydb.mytable
	* tidb-ctl mvcc -t mydb.mytable --start-ts 123
	MVCC info of the first key in mydb.mytable with a specified start ts
	* tidb-ctl mvcc --start-ts 123
	MVCC info of the primary keys with a specified start ts
	* tidb-ctl mvcc -t mydb.mytable --index-name idx --index-values column_name_1: column_value_1, column_name_2: column_value2...
	MVCC info of a specified index record
	`,
	Run: mvccQuery,
}

func mvccQuery(_ *cobra.Command, _ []string) {
	if mvccStartTS != 0 {
		path := txnPrefix + strconv.FormatUint(mvccStartTS, 10)
		if len(mvccTable) != 0 {
			httpPrint(path + "/" + replaceTableFlag(mvccTable))
			return
		}
		httpPrint(path)
		return
	}

	if mvccHID != 0 {
		path := keyPrefix + replaceTableFlag(mvccTable) + "/" + strconv.FormatInt(mvccHID, 10)
		httpPrint(path)
		return
	}

	if len(mvccIndexName) != 0 {
		path := indexPrefix + replaceTableFlag(mvccTable) + "/" + mvccIndexName + "?" + parseValueString(mvccIndexValues)
		httpPrint(path)
		return
	}
}

func parseValueString(values []string) string {
	var str string
	for _, v := range values {
		str += strings.Replace(v, ":", "=", 1) + "&"
	}
	return str[:len(str)-1]
}

func init() {
	rootCmd.AddCommand(mvccCmd)

	mvccCmd.Flags().StringVarP(&mvccTable, "table", "t", "", "Combine with --hid or --start-ts to locate a specified table, database name must included, e.g. `mydb.mytable`.")
	mvccCmd.Flags().Int64Var(&mvccHID, "hid", 0, "Get MVCC info of the key with a specified handle ID, must combine with --table.")
	mvccCmd.Flags().Uint64Var(&mvccStartTS, "start-ts", 0, "Get MVCC info of the primary key, or get MVCC info of the first key in the table (with --table) with a specified start ts.")
	mvccCmd.Flags().StringVar(&mvccIndexName, "index-name", "", "Index Name of a specified index key.")
	mvccCmd.Flags().StringSliceVar(&mvccIndexValues, "index-values", nil, "Get MVCC info of a specified index key, argument example: `column_name_1: column_value_1, column_name_2: column_value2...`")
}
