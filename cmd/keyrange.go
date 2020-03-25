// Copyright 2020 PingCAP, Inc.
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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	encodeKeys bool
	keysDB     string
	keysTable  string
)

var keyRangeCmd = &cobra.Command{
	Use:   "keyrange",
	Short: "Show key ranges",
	Long:  "Show key ranges. It shows global key ranges if database and table name not provided.",
	RunE:  showKeyRanges,
}

func init() {
	keyRangeCmd.PersistentFlags().BoolVarP(&encodeKeys, "encode", "e", false, "encode keys")
	keyRangeCmd.PersistentFlags().StringVarP(&keysDB, dbFlagName, "d", "", "database name")
	keyRangeCmd.PersistentFlags().StringVarP(&keysTable, tableFlagName, "t", "", "table name")
}

func showKeyRanges(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("too many arguments")
	}
	printGlobalKeyRanges()
	if keysDB == "" || keysTable == "" {
		return nil
	}

	body, status, err := httpGet("schema/" + keysDB + "/" + keysTable)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		fmt.Println("status=", status)
		fmt.Println(string(body))
		return nil
	}

	type response struct {
		TableID   int64 `json:"id"`
		TableName struct {
			Name string `json:"O"`
		} `json:"name"`
		Indexes []struct {
			IndexID   int64 `json:"id"`
			IndexName struct {
				Name string `json:"O"`
			} `json:"idx_name"`
		} `json:"index_info"`
	}
	var res response
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("invalid response:", string(body))
		return err
	}
	var indexIDs []int64
	var indexNames []string
	for _, idx := range res.Indexes {
		indexIDs = append(indexIDs, idx.IndexID)
		indexNames = append(indexNames, idx.IndexName.Name)
	}
	printTableKeyRanges(res.TableID, keysTable, indexIDs, indexNames)
	return nil
}

func printGlobalKeyRanges() {
	fmt.Println("global ranges:")
	fmt.Printf("  meta: (%s, %s)\n", fmtKey([]byte("m")), fmtKey([]byte("n")))
	fmt.Printf("  table: (%s, %s)\n", fmtKey([]byte("t")), fmtKey([]byte("u")))
}

func printTableKeyRanges(tableID int64, tableName string, indexIDs []int64, indexNames []string) {
	tablePrefix := encodeInt([]byte("t"), tableID)
	tableEnd := encodeInt([]byte("t"), tableID+1)
	fmt.Printf("table %s ranges: (NOTE: key range might be changed after DDL)\n", tableName)
	fmt.Printf("  table: (%s, %s)\n", fmtKey(tablePrefix), fmtKey(tableEnd))
	indexPrefix := append(tablePrefix, '_', 'i')
	rowPrefix := append(tablePrefix[:len(tablePrefix):len(tablePrefix)], '_', 'r')
	fmt.Printf("  table indexes: (%s, %s)\n", fmtKey(indexPrefix), fmtKey(rowPrefix))
	for i := range indexIDs {
		prefix := encodeInt(indexPrefix, indexIDs[i])
		end := encodeInt(indexPrefix, indexIDs[i]+1)
		fmt.Printf("    index %s: (%s, %s)\n", indexNames[i], fmtKey(prefix), fmtKey(end))
	}
	fmt.Printf("  table rows: (%s, %s)\n", fmtKey(rowPrefix), fmtKey(tableEnd))
}

func fmtKey(k []byte) string {
	if encodeKeys {
		k = encodeBytes(k)
	}
	return hex.EncodeToString(k)
}

const (
	encGroupSize = 8
	encMarker    = byte(0xFF)
)

var pads = make([]byte, encGroupSize)

// encodeBytes encodes a byte slice into TiDB's encoded form.
func encodeBytes(b []byte) []byte {
	dLen := len(b)
	reallocSize := (dLen/encGroupSize + 1) * (encGroupSize + 1)
	result := make([]byte, 0, reallocSize)
	for idx := 0; idx <= dLen; idx += encGroupSize {
		remain := dLen - idx
		padCount := 0
		if remain >= encGroupSize {
			result = append(result, b[idx:idx+encGroupSize]...)
		} else {
			padCount = encGroupSize - remain
			result = append(result, b[idx:]...)
			result = append(result, pads[:padCount]...)
		}

		marker := encMarker - byte(padCount)
		result = append(result, marker)
	}
	return result
}

func encodeInt(b []byte, v int64) []byte {
	var data [8]byte
	u := uint64(v) ^ 0x8000000000000000
	binary.BigEndian.PutUint64(data[:], u)
	return append(b, data[:]...)
}
