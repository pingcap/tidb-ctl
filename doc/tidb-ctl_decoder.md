## tidb-ctl decoder

deocde tidb key

### Synopsis

decoder is a simple tool to make key readable.

### Options

```
  -h, --help   help for decoder
  -f, --format target the format of key you want decode
  -k, --key    target the key you want decode
```

### Usage

When you use `tidb-ctl decoder`, `--format` and `--key` can not be empty.
`tidb-ctl decoder` provide three kind of format.

The first is `table_row` format, it a kind of key which has `tablePrefix`, `rowPrefix`, `tableID`, `rowID`. 
`tablePrefix` is `t`, `rowPrefix` is `_r`, `tableID` and `rowID` are hex strings.
Decode result of `table_row` key are `tableID` and `rowID`.

For example:
```
./tidb-ctl decoder -f=table_row -k "t\x80\x00\x00\x00\x00\x00\x07\x8f_r\x80\x00\x00\x00\x00\x08\x3b\xba"
table_id: 1935
row_id: 539578
```

The second is `table_index` format, it a kind of key which has `tablePrefix`, `indexPrefix`, `tableID`, `indexID` and `indexValue`.
`tablePrefix` is `t`, `indexPrefix` is `_i`, `tableID`, `indexID` and `indexValue` are hex strings.
Decode result of `table_index` key are `tableID`, `indexID`.
`indexValue` will be decoded as pairs of type and value.

For example:
```
./tidb-ctl decoder -f=table_index -k "t\x80\x00\x00\x00\x00\x00\x00\x5f_i\x80\x00\x00\x00\x00\x00\x00\x01\x03\x80\x00\x00\x00\x00\x00\x00\x02\x03\x80\x00\x00\x00\x00\x00\x00\x02"
table_id: 95
index_id: 1
type: Int64, value: 2
type: Int64, value: 2
```

The third is `value`, it is a base64 encoded byte array. 
It will be decoded as following format:
```
type of column1, columnID1
value of column1
type of column2, columnID2
value of column2
...
```

For example:
```
./tidb-ctl decoder -f=value -k "CAQCBmFiYw=="
type: Int64, value: 2
type: Bytes, value: abc
```
