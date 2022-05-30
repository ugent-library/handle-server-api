handle-server-api
-----------

Temporary rest api that directly

inserts data into mysql of handle server

# Build requirements

Golang v1.17

# Build

```
$ cd handle-server-api
$ go build
```

# Run

Live

```
$ cd handle-server-api
$ go run main.go
```

Or after build

```
$ ./handle-server-api
```

# Run options

The application uses the following run options in order:

* option from command line flag

* option from environment variable (prefix `HDL_`, and all capital letters)

* option from internal default

**prefix**

  required prefix.

  Required: true

  Environment variable: `HDL_PREFIX`

  Default: there is no default

**bind**

  Bind to host and address

  Environment variable: `HDL_BIND`

  Internal default: `:3000`

**dsn**

  SQL DSN to the database of handle.net

  Environment variable: `HDL_DSN`

  Internal default: `handle:handle@tcp(127.0.0.1:3306)/handle`

**auth-username**

  Username for basic auth

  Environment variable: `HDL_AUTH_USERNAME`

  Internal default: `handle`

**auth-password**

  Password for basic auth

  Environment variable: `HDL_AUTH_PASSWORD`

  Internal default: `handle`

# API

I've tried to be as close to the original API as possible.

See [Technical Manual Version 9](http://www.handle.net/tech_manual/HN_Tech_Manual_9.pdf)

All endpoint require basic http authentication.

See run options `-auth-username` and `auth-password` above

## GET /handles/{prefix}/{local_id}

Retrieve all handle values of type URL

for handle `{prefix}/{local_id}`

Response:

```
{
   "values" : [
      {
         "ttl" : 86400,
         "type" : "URL",
         "timestamp" : "2022-05-04T06:45:34Z",
         "data" : {
            "format" : "string",
            "url" : "https://biblio.ugent.be/publication/1000117"
         },
         "index" : 1
      }
   ],
   "handle" : "1854/LU-1000117",
   "responseCode" : 1
}
```

`responseCode: 1` means "successfull"

`responseCode: 100` means "not found"

but http status codes also help, although a 404

may also mean that you've hit the wrong controller

## DELETE /handles/{prefix}/{local_id}

Delete all handle values for handle `{prefix}/{local_id}`

## PUT /handles/{prefix}/{local_id}

Insert/replace all handle value for handle `{prefix}/{local_id}`

Only body parameter `url` is supported

Returns same response as the GET controller
