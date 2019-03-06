// Copyright 2019 PingCAP, Inc.
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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type parameter struct {
	// It is parameter for api_grpc_gateway
	Key      string `json:"key"`
	RangeEnd string `json:"range_end"`
}

var (
	dialClient       = &http.Client{}
	rangeQueryPrefix = "v3/kv/range"
	rangeDelPrefix   = "v3/kv/deleterange"
	putPrefix        = "v3/kv/put"
)

// newEtcdCommand returns a etcd subcommand of rootCmd.
func newEtcdCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "etcd",
		Short: "control the info about etcd by grpc_gateway",
	}
	m.AddCommand(newShowDDLInfoCommand())
	m.AddCommand(newDelKeyCommand())
	m.AddCommand(newPutKeyCommand())
	return m
}

// newShowDDLInfoCommand returns a show ddl information subcommand of EtcdCommand.
func newShowDDLInfoCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "ddlinfo",
		Short: "Show All Information about DDL",
		Run:   showDDLInfoCommandFunc,
	}
	return m
}

// newDelKeyCommand returns a delete key subcommand of EtcdCommand.
func newDelKeyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "delkey",
		Short: "del key by `delkey [key]`",
		Run:   delKeyCommandFunc,
	}
	return m
}

// newPutKeyCommand returns a put key subcommand of EtcdCommand.
func newPutKeyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "putkey",
		Short: "[ONLY FOR TEST!] put key in schema version by `putkey [key] [value]`",
		Run:   putKeyCommandFunc,
	}
	return m
}

func showDDLInfoCommandFunc(cmd *cobra.Command, args []string) {
	res, err := getDDLInfo()
	if err != nil {
		cmd.Printf("Failed to show DDLInfo: %v\n", err)
		return
	}
	cmd.Println(res)
}

func delKeyCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Printf("Only one argument!")
		return
	}

	key := args[0]
	ddlOwnerKeyPrefix := "/tidb/ddl/fg/owner/"
	ddlAllSchemaVersionsPrefix = "/tidb/ddl/all_schema_versions/"
	if !(strings.HasPrefix(key, ddlOwnerKeyPrefix) || strings.HasPrefix(key, ddlAllSchemaVersionsPrefix)) {
		cmd.Printf("This function only for delete something about DDL")
		return
	}

	ddlInfo, err := getDDLInfo()
	if err != nil {
		cmd.Printf("Failed to delete key: %v\n", err)
		return
	}

	findKey := false
	var jsn struct {
		Count  string              `json:"count"`
		Header map[string]string   `json:"header"`
		Kvs    []map[string]string `json:"kvs"`
	}
	err = json.Unmarshal([]byte(ddlInfo), &jsn)
	if err != nil {
		cmd.Printf("Failed to delete key: %v\n", err)
		return
	}
	// Check key exists
	for _, v := range jsn.Kvs {
		if findKey {
			break
		}
		for kk, vv := range v {
			if kk == "key" && vv == key {
				findKey = true
				break
			}
		}
	}

	if !findKey {
		cmd.Printf("Failed to delete key: Key not found!")
		return
	}

	var para = &parameter{
		Key: base64Encode(key),
	}

	reqData, err := json.Marshal(para)
	if err != nil {
		cmd.Printf("Failed to delete key: %v\n", err)
		return
	}
	req, err := getRequest(rangeDelPrefix, http.MethodPost, "application/json",
		bytes.NewBuffer(reqData))
	if err != nil {
		cmd.Printf("Failed to delete key: %v\n", err)
		return
	}
	res, err := dail(req)
	if err != nil {
		cmd.Printf("Failed to delete key: %v\n", err)
		return
	}

	cmd.Println(res)
}

func putKeyCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Printf("Only two arguments!")
		return
	}
	var putParameter struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	putKeyPreFix := "/tidb/ddl/all_schema_versions/"
	putParameter.Key = base64Encode(putKeyPreFix + args[0])
	putParameter.Value = base64Encode(args[1])

	reqData, err := json.Marshal(putParameter)
	if err != nil {
		cmd.Printf("Failed to put key: %v\n", err)
		return
	}
	req, err := getRequest(putPrefix, http.MethodPost, "application/json",
		bytes.NewBuffer(reqData))
	if err != nil {
		cmd.Printf("Failed to put key: %v\n", err)
		return
	}
	res, err := dail(req)
	if err != nil {
		cmd.Printf("Failed to put key: %v\n", err)
		return
	}

	cmd.Println(res)
}

func getDDLInfo() (string, error) {
	st := "/tidb/ddl"
	ed := "/tidb/ddm"
	var rangeQueryDDLInfo = &parameter{
		Key:      base64Encode(st),
		RangeEnd: base64Encode(ed),
	}

	reqData, err := json.Marshal(rangeQueryDDLInfo)
	if err != nil {
		return "", err
	}
	req, err := getRequest(rangeQueryPrefix, http.MethodPost, "application/json",
		bytes.NewBuffer(reqData))
	if err != nil {
		return "", err
	}
	res, err := dail(req)
	if err != nil {
		return "", err
	}

	res, err = formatJSONAndBase64Decode(res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func getRequest(prefix string, method string, bodyType string, body io.Reader) (*http.Request, error) {
	if method == "" {
		method = http.MethodGet
	}
	url := "http://" + pdHost.String() + ":" + strconv.Itoa(int(pdPort)) + "/" + prefix
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return req, err
}

func dail(req *http.Request) (string, error) {
	var res string
	reps, err := dialClient.Do(req)
	if err != nil {
		return res, err
	}
	defer reps.Body.Close()
	if reps.StatusCode != http.StatusOK {
		return res, genResponseError(reps)
	}

	r, err := ioutil.ReadAll(reps.Body)
	if err != nil {
		return res, err
	}
	res = string(r)
	return res, nil
}

func genResponseError(r *http.Response) error {
	res, _ := ioutil.ReadAll(r.Body)
	return errors.Errorf("[%d] %s", r.StatusCode, res)
}

func formatJSONAndBase64Decode(str string) (string, error) {
	var jsn struct {
		Count  string              `json:"count"`
		Header map[string]string   `json:"header"`
		Kvs    []map[string]string `json:"kvs"`
	}

	err := json.Unmarshal([]byte(str), &jsn)

	// Base64Decode for key and value.
	for k, v := range jsn.Kvs {
		for kk, vv := range v {
			if kk == "key" || kk == "value" {
				vv, err = base64Decode(vv)
				if err != nil {
					return "", err
				}
				jsn.Kvs[k][kk] = vv
			}
		}
	}

	if err != nil {
		return "", err
	}
	resByte, err := json.MarshalIndent(&jsn, "", "\t")
	if err != nil {
		return "", err
	}
	res := string(resByte)
	return res, nil
}
