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
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/model"
	"github.com/pingcap/tidb/tablecodec"
	"github.com/pingcap/tidb/types"
	"github.com/spf13/cobra"
)

// base64decodeCmd represents the base64decode command
var newBase64decodeCmd = &cobra.Command{
	Use:     "base64decode",
	Short:   "decode base64 value",
	Long:    "decode base64 value to hex and uint64",
	Example: "tidb-ctl base64decode [base64_data]\ntidb-ctl base64decode [db_name.table_name] [base64_data]\ntidb-ctl base64decode [table_id] [base64_data]",
	RunE:    base64decodeCmd,
}

func base64decodeCmd(c *cobra.Command, args []string) error {
	if len(args) == 1 {
		return decodeBase64Value(c, args[0])
	} else if len(args) == 2 {
		return decodeTableMVCC(c, args)
	} else {
		return fmt.Errorf("only support 1 or 2 argument")
	}
}

func decodeBase64Value(c *cobra.Command, inputValue string) error {
	uDec, err := base64Decode(inputValue)
	if err != nil {
		return err
	}
	if len(uDec) <= 8 {
		var num uint64
		hexStr := hex.EncodeToString([]byte(uDec))
		c.Printf("hex: %s\n", hexStr)
		err = binary.Read(bytes.NewBuffer([]byte(uDec)[0:8]), binary.BigEndian, &num)
		if err != nil {
			return err
		}
		c.Printf("uint64: %d\n", num)
	}
	return nil
}

func decodeTableMVCC(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.Errorf("need 2 param. eg: tidb-ctl decodeTable dbName.tableName raw_data")
	}
	tblInfo, err := getTableInfo(args[0])
	if err != nil {
		return err
	}
	result, err := decodeMVCC(tblInfo, args[1])
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func getTableInfo(id string) (tblInfo *model.TableInfo, err error) {
	url := ""
	if strings.Contains(id, ".") {
		fields := strings.Split(id, ".")
		if len(fields) != 2 {
			return nil, errors.Errorf("wrong table name. need like: test.t1")
		}
		url = "/" + fields[0] + "/" + fields[1]
	} else {
		// treat as table id.
		url = "?table_id=" + id
	}

	url = schema + "://" + host.String() + ":" + strconv.Itoa(int(port)) + "/" + "schema" + url
	var resp *http.Response
	resp, err = ctlClient.Get(url)
	if err != nil {
		return
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("get table info status code is no ok. body: %s", string(body))
	}
	tblInfo = &model.TableInfo{}
	err = json.Unmarshal(body, tblInfo)
	return tblInfo, err
}

func decodeMVCC(tbl *model.TableInfo, base64Str string) (string, error) {
	if len(base64Str) == 0 {
		return "", errors.Errorf("no data?")
	}
	var buf bytes.Buffer
	bs, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}
	colMap := make(map[int64]*types.FieldType, 3)
	for _, col := range tbl.Columns {
		colMap[col.ID] = &col.FieldType
	}

	r, err := tablecodec.DecodeRow(bs, colMap, time.UTC)
	if err != nil {
		return "", err
	}
	if r == nil {
		return "", errors.Errorf("no data???")
	}

	for _, col := range tbl.Columns {
		if v, ok := r[col.ID]; ok {
			if v.IsNull() {
				buf.WriteString(col.Name.L + " is NULL\n")
				continue
			}
			ss, err := v.ToString()
			if err != nil {
				buf.WriteString(col.Name.L + " ToString error: " + err.Error() + fmt.Sprintf("datum: %#v", v) + "\n")
				continue
			}
			buf.WriteString(col.Name.L + ":\t" + ss + "\n")
		} else {
			buf.WriteString(col.Name.L + " not found in data\n")
		}
	}
	return buf.String(), nil
}
