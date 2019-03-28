## tidb-ctl base64decode

decode base64 value

### Synopsis


decode base64 value to hex and uint64

```
tidb-ctl base64decode [flags]
```

### Examples

#### Decode base64 value: `tidb-ctl base64decode [base64_data`

```shell
▶ tidb-ctl base64decode AAAAAAAAAAE=
hex: 0000000000000001
uint64: 1

```

#### Decode base64 with table schema.`tidb-ctl base64decode db_name.table_name [base64_data]`

   **prepare execute below sql**

```shell
use test;
create table t (a int, b varchar(20),c datetime default current_timestamp , d timestamp default current_timestamp);
insert into t (a,b,c) values(1,"哈哈 hello",NULL);
alter table t add column e varchar(20);
```

**then you can use http api to get MVCC data**

```shell
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
```

**then decode table base64 raw data**

```shell
▶ ./tidb-ctl base64decode test.t CAIIAggEAhjlk4jlk4ggaGVsbG8IBgAICAmAgICI0Yyr0Rk=
a:      1
b:      哈哈 hello
c is NULL
d:      2019-03-22 06:20:17
e not found in data


# if the table id of test.t is 60, you can also use below command to do the same thing.
▶ ./tidb-ctl base64decode 60 CAIIAggEAhjlk4jlk4ggaGVsbG8IBgAICAmAgICI0Yyr0Rk=
a:      1
b:      哈哈 hello
c is NULL
d:      2019-03-22 06:20:17
e not found in data
```

As you can see, data of column c is NULL, and data of column e is not found in data, because e is added latter, TiDB currently have not back fill data for added after column.

### Options

```
  -h, --help   help for base64decode
```

### SEE ALSO
* [tidb-ctl](tidb-ctl.md)	 - TiDB Controller

