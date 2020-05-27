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
	"fmt"
	"io"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/codec"
	"github.com/spf13/cobra"
)

// decoderCmd represents the key-decoder command
var decoderCmd = &cobra.Command{
	Use:   "decoder",
	Short: "decode key",
	Long: `decode "key"
	currently support:
	table_row:   key format like 'txxx_rxxx'
	table_index: key format like 'txxx_ixxx'
	value:       base64 encoded value`,
	RunE: decodeKeyFunc,
}

type indexValue struct {
	typename, valueStr string
}

func decodeKey(text string) (string, error) {
	var buf []byte
	r := bytes.NewBuffer([]byte(text))
	for {
		c, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		if c != '\\' {
			buf = append(buf, c)
			continue
		}
		n := r.Next(1)
		if len(n) == 0 {
			return "", io.EOF
		}
		// See: https://golang.org/ref/spec#Rune_literals
		if idx := strings.IndexByte(`abfnrtv\'"`, n[0]); idx != -1 {
			buf = append(buf, []byte("\a\b\f\n\r\t\v\\'\"")[idx])
			continue
		}

		switch n[0] {
		case 'x':
			_, err := fmt.Sscanf(string(r.Next(2)), "%02x", &c)
			if err != nil {
				return "", err
			}
			buf = append(buf, c)
		default:
			n = append(n, r.Next(2)...)
			_, err := fmt.Sscanf(string(n), "%03o", &c)
			if err != nil {
				return "", err
			}
			buf = append(buf, c)
		}
	}
	return string(buf), nil
}

func decodeIndexValue(buf []byte) ([]indexValue, error) {
	key := buf
	values := make([]indexValue, 0, 10)
	for len(key) > 0 {
		remain, d, err := codec.DecodeOne(key)
		if err != nil {
			break
		}
		s, err := d.ToString()
		if err != nil {
			return nil, err
		}
		typeStr := types.KindStr(d.Kind())
		values = append(values, indexValue{typename: typeStr, valueStr: s})
		key = remain
	}
	return values, nil
}

func decodeTableIndex(buf []byte) (int64, int64, []indexValue, error) {
	if len(buf) >= 19 && buf[0] == 't' && buf[9] == '_' && buf[10] == 'i' {
		tableid, rowid, indexValue := buf[1:9], buf[11:19], buf[19:]
		_, tableID, err := codec.DecodeInt(tableid)
		if err != nil {
			return 0, 0, nil, err
		}
		_, rowID, err := codec.DecodeInt(rowid)
		if err != nil {
			return 0, 0, nil, err
		}
		values, err := decodeIndexValue(indexValue)
		if err != nil {
			return 0, 0, nil, err
		}
		return tableID, rowID, values, nil
	} else if len(buf) >= 22 && buf[0] == 't' && buf[10] == '_' && buf[11] == 'i' {
		tmp := make([]byte, 0)
		for i, val := range buf {
			if (i+1)%9 != 0 {
				tmp = append(tmp, val)
			}
		}
		pad := int(255 - buf[len(buf)-1])
		tmp = tmp[:len(tmp)-pad]
		tableid, rowid, indexValue := tmp[1:9], tmp[11:19], tmp[19:]
		_, tableID, err := codec.DecodeInt(tableid)
		if err != nil {
			return 0, 0, nil, err
		}
		_, rowID, err := codec.DecodeInt(rowid)
		if err != nil {
			return 0, 0, nil, err
		}
		values, err := decodeIndexValue(indexValue)
		if err != nil {
			return 0, 0, nil, err
		}
		return tableID, rowID, values, nil
	}
	return 0, 0, nil, errors.Errorf("illegal code format")
}

func decodeTableRow(buf []byte) (int64, int64, error) {
	if len(buf) >= 19 && buf[0] == 't' && buf[9] == '_' && buf[10] == 'r' {
		tableid, rowid := buf[1:9], buf[11:19]
		_, tableID, err := codec.DecodeInt(tableid)
		if err != nil {
			return 0, 0, err
		}
		_, rowID, err := codec.DecodeInt(rowid)
		if err != nil {
			return 0, 0, err
		}
		return tableID, rowID, nil
	} else if len(buf) >= 22 && buf[0] == 't' && buf[10] == '_' && buf[11] == 'r' {
		tmp := buf[:22]
		tableid, rowid := make([]byte, 0, 8), make([]byte, 0, 8)
		for i, val := range tmp {
			if i == 8 || i == 17 {
				continue
			}
			if i > 0 && i < 10 {
				tableid = append(tableid, val)
			} else if i > 11 {
				rowid = append(rowid, val)
			}
		}
		_, tableID, err := codec.DecodeInt(tableid)
		if err != nil {
			return 0, 0, err
		}
		_, rowID, err := codec.DecodeInt(rowid)
		if err != nil {
			return 0, 0, err
		}
		return tableID, rowID, nil
	}
	return 0, 0, errors.Errorf("illegal code format")
}

func decodeKeyFunc(c *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}
	keyValue := args[0]
	raw, err := decodeKey(keyValue)
	if err != nil {
		return err
	}
	// Try to decode using table_row format.
	tableID, rowID, err := decodeTableRow([]byte(raw))
	if err == nil {
		c.Printf("format: table_row\ntable_id: %v\nrow_id: %v\n", tableID, rowID)
		return nil
	}
	// Try to decode using table_index format.
	tableID, rowID, indexvalues, err := decodeTableIndex([]byte(raw))
	if err == nil {
		c.Printf("format: table_index\ntable_id: %v\nindex_id: %v\n", tableID, rowID)
		for i, iv := range indexvalues {
			c.Printf("index_value[%v]: {type: %v, value: %v}\n", i, iv.typename, iv.valueStr)
		}
		return nil
	}
	// Try to decode base64 format key.
	b64decode, err := base64.StdEncoding.DecodeString(keyValue)
	if err != nil {
		return err
	}
	tableID, rowID, err = decodeTableRow(b64decode)
	if err == nil {
		c.Printf("format: table_row\ntable_id: %v\nrow_id: %v\n", tableID, rowID)
		return nil
	}
	tableID, rowID, indexvalues, err = decodeTableIndex(b64decode)
	if err == nil {
		c.Printf("format: table_index\ntable_id: %v\nindex_id: %v\n", tableID, rowID)
		for i, iv := range indexvalues {
			c.Printf("index_value[%v]: {type: %v, value: %v}\n", i, iv.typename, iv.valueStr)
		}
		return nil
	}
	// Try to decode base64 format index_value.
	indexvalues, err = decodeIndexValue(b64decode)
	if err != nil {
		return err
	}
	c.Printf("format: index_value\n")
	for i, iv := range indexvalues {
		c.Printf("index_value[%v]: {type: %v, value: %v}\n", i, iv.typename, iv.valueStr)
	}
	return nil
}
