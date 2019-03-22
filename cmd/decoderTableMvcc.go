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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/tidb/model"
	"github.com/pingcap/tidb/tablecodec"
	"github.com/pingcap/tidb/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var decodeTableExample string = `
**prepare execute below sql**:

use test;
create table t (a int, b varchar(20),c datetime default current_timestamp , d timestamp default current_timestamp);
insert into t (a,b,c) values(1,"哈哈 hello",NULL);
alter table t add column e varchar(20);

**then you can use http api to get MVCC data.***

▶ curl "http://$IP:10080/mvcc/key/test/t/1"
{
 "info": {
  "writes": [
   {
    "start_ts": 407171055877619718,
    "commit_ts": 407171055877619719,
    "short_value": "CAQCGOmZiOmcnCBoZWxsbwgGAAgICYCAgIjqi6vRGQ=="
   }
  ]
 }

**then use decodeTable use decode table MVCC data**
▶ ./tidb-ctl decodeTable test.t CAIIAggEAhjlk4jlk4ggaGVsbG8IBgAICAmAgICI0Yyr0Rk=
a:      1
b:      哈哈 hello
c is NULL
d:      2019-03-22 06:20:17
e not found in data


if the table id of test.t is 56, you can also use below command to do the same thing.

▶ ./tidb-ctl decodeTable 60 CAIIAggEAhjlk4jlk4ggaGVsbG8IBgAICAmAgICI0Yyr0Rk=
a:      1
b:      哈哈 hello
c is NULL
d:      2019-03-22 06:20:17
e not found in data

As you can see, data of column c is NULL, and data of column e is not found in data, because e is added latter, TiDB currently have not back fill data for added after column.

`

// schemaRootCmd represents the schema command
var decodeTableCmd = &cobra.Command{
	Use:     "decodeTable",
	Short:   "Decode table mvcc data",
	Long:    "'tidb-ctl decodeTable' decode table mvcc data.",
	Example: decodeTableExample,
	RunE:    decodeTableMVCC,
}

func decodeTableMVCC(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.Errorf("need 2 param. eg: tidb-ctl decodeTable test.t CAIIAggEAhjpmYjpnJwgaGVsbG8IBgAICAmAgICg9oqr0Rk=")
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
		// tread as table id.
		url = "?table_id=" + id
	}

	url = "http://" + host.String() + ":" + strconv.Itoa(int(port)) + "/" + "schema" + url
	var resp *http.Response
	resp, err = http.Get(url)
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
		return nil, errors.Errorf("get  table info status code is no ok. body: %s", string(body))
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
