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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	. "github.com/pingcap/check"
)

var _ = Suite(&etcdTestSuite{})

type etcdTestSuite struct{}

func (s *etcdTestSuite) TestDDLInfo(c *C) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		c.Assert(r.Method, Equals, http.MethodPost)
		c.Assert(r.URL.EscapedPath(), Equals, rangeQueryPrefix)
		result, err := ioutil.ReadAll(r.Body)
		c.Assert(err, IsNil)
		var resp parameter
		err = json.Unmarshal(result, &resp)
		c.Assert(err, IsNil)
		c.Assert(resp.Key, Equals, base64Encode("/tidb/ddl"))
		c.Assert(resp.RangeEnd, Equals, base64Encode("/tidb/ddm"))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	c.Assert(err, IsNil)
	uArr := strings.Split(u.Host, ":")
	cmd := initCommand()
	args := []string{"etcd", "ddlinfo", "-i", uArr[0], "-p", uArr[1]}
	_, _, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
}

func (s *etcdTestSuite) TestDelKey(c *C) {
	testKey := "test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		c.Assert(r.Method, Equals, http.MethodPost)
		c.Assert(r.URL.EscapedPath(), Equals, rangeDelPrefix)
		result, err := ioutil.ReadAll(r.Body)
		c.Assert(err, IsNil)
		var resp parameter
		err = json.Unmarshal(result, &resp)
		c.Assert(err, IsNil)
		c.Assert(resp.Key, Equals, base64Encode(testKey))
		c.Assert(resp.RangeEnd, Equals, "")
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	c.Assert(err, IsNil)
	uArr := strings.Split(u.Host, ":")
	cmd := initCommand()
	args := []string{"etcd", "delkey", testKey, "-i", uArr[0], "-p", uArr[1]}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Assert(string(output), Equals, "This function only for delete the key-value about DDL\n")

	testKey = "/tidb/ddl/12345"
	args = []string{"etcd", "delkey", testKey, "-i", uArr[0], "-p", uArr[1]}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Assert(output, NotNil)
}

func (s *etcdTestSuite) TestPutKey(c *C) {
	testKey := "test"
	testValue := "test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		c.Assert(r.Method, Equals, http.MethodPost)
		c.Assert(r.URL.EscapedPath(), Equals, putPrefix)
		result, err := ioutil.ReadAll(r.Body)
		c.Assert(err, IsNil)
		var resp struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		err = json.Unmarshal(result, &resp)
		c.Assert(err, IsNil)
		c.Assert(resp.Key, Equals, base64Encode(ddlAllSchemaVersionsPrefix+testKey))
		c.Assert(resp.Value, Equals, base64Encode(testValue))
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	c.Assert(err, IsNil)
	uArr := strings.Split(u.Host, ":")
	cmd := initCommand()
	args := []string{"etcd", "putkey", testKey, testValue, "-i", uArr[0], "-p", uArr[1]}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Assert(output, NotNil)
}
