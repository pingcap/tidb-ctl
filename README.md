# tidb-ctl

TiDB Controller (tidb-ctl) is a command line tool for TiDB Server (tidb-server).

## Build

1. Make sure [*Go*](https://golang.org/) (version 1.9+) is installed.
1. Use `make` in repo root path. `tidb-ctl` will build in `bin` directory.

## Usage

Run:

    ./tidb-ctl store -h 127.0.0.1:10080

show all stores status. '-h' specify the tidb-server address, it can be overwritten by setting the environment variable `TIDB_CTL_ADDR`. Such as `export TIDB_CTL_ADDR=127.0.0.1:10080`

### Flags

#### --Host, -h

* The TiDB server address
* default: <http://127.0.0.1:10080>
* env variable: `TIDB_CTL_ADDR`

### Command

#### schema

* : If no argument specified, show all databases schema info.
* --database, -d : Show all tables schema info of specified database.
* --table, -t : Get schema info of a specified table, database name must included, e.g. `mysql.user`.
* --tid : Get schema info of a sepcified table id.

#### region

* --table, -t : Get regions info of a specified table, database name must included, e.g. `mysql.user`.
* --meta, -m : Get meta data of all regions.
* --rid, -r : Get region info of a sepcified region id.

#### mvcc

* --table, -t : Combine with --hid or --start-ts to locate a specified table, database name must included, e.g. `mysql.user`.
* --hid : Get MVCC info of the key with a specified handle ID, must combine with --table.
* --start-ts : -m : Get MVCC info of the primary key, or get MVCC info of the first key in the table (with --table) with a specified start ts.
* --index : Get MVCC info of a specified index key, argument example: `index_name:{column_name_1: column_value_1, column_name_2: column_value...}`
