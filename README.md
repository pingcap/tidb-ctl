# tidb-ctl

TiDB Controller (tidb-ctl) is a command line tool for TiDB Server (tidb-server).

## Build

1. Make sure [*Go*](https://golang.org/) (version 1.9+) is installed.
1. Use `make` in repo root path. `tidb-ctl` will build in `bin` directory.

## Usage

Run:

    ./tidb-ctl store -h 127.0.0.1:10080

show all stores status. '-h' specify the tidb-server address, it can be overwritten by setting the environment variable `TIDB_CTL_ADDR`. Such as `export TIDB_CTL_ADDR=127.0.0.1:10080`

### Globe Flags

#### --config, -c

* TiDB Controller config file
* default: $HOME/.tidb-ctl.yaml

#### --host

* The TiDB server host IP
* default: 127.0.0.1

#### --port

* The TiDB server port
* default: 10080

### Command

#### schema

* : If no argument specified, show all databases schema info.
* --database, -d : Show all tables schema info of specified database.
* --table, -t : Get schema info of a specified table, database name must included, e.g. `mydb.mytable`.
* --tid : Get schema info of a specified table id.

#### region

* --table, -t : Get regions info of a specified table, database name must included, e.g. `mydb.mytable`.
* --meta, -m : Get meta data of all regions.
* --rid, -r : Get region info of a specified region id.

#### mvcc

* --table, -t : Combine with --hid or --start-ts to locate a specified table, database name must included, e.g. `mydb.mytable`.
* --hid : Get MVCC info of the key with a specified handle ID, must combine with --table.
* --start-ts : Get MVCC info of the primary key, or get MVCC info of the first key in the table (with --table) with a specified start ts.
* --index-name : Index Name of a specified index key.
* --index-values: Get MVCC info of a specified index key, argument example: `{column_name_1: column_value_1, column_name_2: column_value2...}`
