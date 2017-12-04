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

	"github.com/spf13/cobra"
)

const (
	schemaPrefix  = "schema/"
	tableIDPrefix = "schema?table_id="
)

var (
	schemaDB    string
	schemaTable string
	schemaTID   int64
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Schema Information Query",
	Long: `Query for Schema information, e.g.
	* tidb-ctl schema
	Show all databases schema info.
	* tidb-ctl schema -d dbname
	Show all tables schema info of specified database.
	* tidb-ctl schema -t mydb.mytable
	Get schema info of a specified table, database name must included.
	* tidb-ctl schema --tid 123
	Get schema info of a specified table id.
	`,
	Run: schemaQuery,
}

func schemaQuery(_ *cobra.Command, _ []string) {
	var path string
	if len(schemaDB) != 0 {
		path = schemaPrefix + schemaDB
		httpPrint(path)
		return
	}

	if len(schemaTable) != 0 {
		path = schemaPrefix + replaceTableFlag(schemaTable)
		httpPrint(path)
		return
	}

	if schemaTID != 0 {
		path = tableIDPrefix + strconv.FormatInt(schemaTID, 10)
		httpPrint(path)
		return
	}

	path = schemaPrefix[:len(schemaPrefix)-1]
	httpPrint(path)
}

func init() {
	rootCmd.AddCommand(schemaCmd)

	schemaCmd.Flags().StringVarP(&schemaDB, "database", "d", "", "Show all tables schema info of specified database.")
	schemaCmd.Flags().StringVarP(&schemaTable, "table", "t", "", "Get schema info of a specified table, database name must included, e.g. `mydb.mytable`.")
	schemaCmd.Flags().Int64Var(&schemaTID, "tid", 0, "Get schema info of a specified table id.")
}
