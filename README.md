hdl-srv-api
-----------

Temporary rest api that directly

inserts data into mysql of handle server

# Build requirements

Golang v1.17

# Build

```
$ cd hdl-srv-api
$ go build
```

# Run

Live

```
$ cd hdl-srv-api
$ go run main.go
```

Or after build

```
$ ./hdl-srv-api
```

# Run options

**bind**

  Bind to host and address
  Default: `:3000`

**dsn**

  SQL DSN to the database of handle.net
  Default: `handle:handle@tcp(127.0.0.1:3306)/handle`

**auth-username**

  Username for basic auth
  Default: `handle`

**auth-password**

  Password for basic auth
  Default: `handle`
