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
	schemaRoot       = "schema"
	schemaRootPrefix = schemaRoot + "/"
	tableIDPrefix    = schemaRootPrefix + "?table_id="
)

// schema command flags
var (
	schemaTable string
	schemaTID   int64
)

// schemaRootCmd represents the schema command
var schemaRootCmd = &cobra.Command{
	Use:   "schema",
	Short: "Schema Information",
	Long:  "'tidb-ctl schema' to list all databases schema info.",
	RunE:  listDatabases,
}

func listDatabases(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	return httpPrint(schemaRoot)
}

func init() {
	idFlagName := "id"

	schemaRootCmd.AddCommand(listTableByNameCmd, listTableByIDCmd)

	listTableByNameCmd.Flags().StringVarP(&schemaTable, "name", "n", "", "get schema info of a specified table.")
	listTableByIDCmd.Flags().Int64VarP(&schemaTID, idFlagName, "i", 0, "get schema info of a specified table id.")

	if err := listTableByIDCmd.MarkFlagRequired(idFlagName); err != nil {
		fmt.Printf("can not mark required flag, flag %s is not found", idFlagName)
		return
	}
}

// listTableByNameCmd represents the list table schema by name command
var listTableByNameCmd = &cobra.Command{
	Use:   "in",
	Short: "Schema Information of Tables In Database",
	Long: `Get for Schema information, e.g.
* tidb-ctl schema in [database name]
Show all tables schema info of specified database.
* tidb-ctl schema in [database name] --name(-n) [table name]
Get schema info of a specified table in database.
`,
	RunE: listTableByName,
}

func listTableByName(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expect one argument as database name")
	}
	if len(schemaTable) != 0 {
		return httpPrint(schemaRootPrefix + args[0] + "/" + schemaTable)
	}
	return httpPrint(schemaRootPrefix + args[0])
}

var listTableByIDCmd = &cobra.Command{
	Use:   "tid",
	Short: "Schema Information of Tables By TableID",
	Long:  "'tidb-ctl schema tid --id(-i) [tableID]' to get schema info of a specified table id.",
	RunE:  listTableByID,
}

func listTableByID(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	return httpPrint(tableIDPrefix + strconv.FormatInt(schemaTID, 10))
}
