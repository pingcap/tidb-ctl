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
	. "github.com/pingcap/check"
)

var _ = Suite(&decoderTestSuite{})

type decoderTestSuite struct{}

func (s *decoderTestSuite) TestTableRowDecode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "t\x80\x00\x00\x00\x00\x00\x07\x8f_r\x80\x00\x00\x00\x00\x08\x3b\xba"}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\ntable_id: 1935\nrow_id: 539578\n")

	args = []string{"decoder", "t\200\000\000\000\000\000\007\217_r\200\000\000\000\000\010;\272"}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\ntable_id: 1935\nrow_id: 539578\n")

	args = []string{"decoder", "t\200\000\000\000\000\000\025\377\316_r\200\000\001j\331\377\357vI\000\000\000\000\000\372"}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\ntable_id: 5582\nrow_id: 1558434510409\n")
}

func (s *decoderTestSuite) TestTableIndexDecode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "t\x80\x00\x00\x00\x00\x00\x00\x5f_i\x80\x00\x00\x00\x00\x00\x00\x01\x03\x80\x00\x00\x00\x00\x00\x00\x02\x03\x80\x00\x00\x00\x00\x00\x00\x02"}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_index\n"+
		"table_id: 95\n"+
		"index_id: 1\n"+
		"index_value[0]: {type: bigint, value: 2}\n"+
		"index_value[1]: {type: bigint, value: 2}\n")

	args = []string{"decoder", "t\200\000\000\000\000\000\000\255_i\200\000\000\000\000\000\000\001\003\200\000\000\000\000e\221|\003\200\000\000\000\0008\307\024\003\200\000\000\000\0014\025\230\003\200\000\000\000"}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_index\n"+
		"table_id: 173\n"+
		"index_id: 1\n"+
		"index_value[0]: {type: bigint, value: 6656380}\n"+
		"index_value[1]: {type: bigint, value: 3720980}\n"+
		"index_value[2]: {type: bigint, value: 20190616}\n")
}

func (s *decoderTestSuite) TestBase64Decode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "CAQCBmFiYw=="}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: index_value\n"+
		"index_value[0]: {type: bigint, value: 2}\n"+
		"index_value[1]: {type: bytes, value: abc}\n")
	args = []string{"decoder", "dIAAAAAAAABAX3KAAAAAAAAAAQ=="}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\n"+
		"table_id: 64\n"+
		"row_id: 1\n")
}
