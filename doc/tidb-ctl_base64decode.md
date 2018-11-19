## tidb-ctl base64decode

deocde base64 value to uint64

### Options

```
  -h, --help   help for base64decode
  -v  --value  the base64 value you want decode
```

### Usage
If `--value` can not be decoded to `uint64` value, it will return nothing.
else it will return `uint64` value and `hex` value.


