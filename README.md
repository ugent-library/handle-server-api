handle-server-api
-----------

Temporary rest api that directly

inserts data into postgres of handle server

# Build requirements

Golang v1.19

# Build

```
$ cd handle-server-api
$ go build
```

# Run

Live

```
$ cd handle-server-api
$ go run main.go server
```

Or after build

```
$ ./handle-server-api server
```

# Runtime options

Runtime options can be set with the following environment variables.

* `HDL_PREFIX`

  required prefix.

  Required: true

  Default: there is no default

* `HDL_HOST`

  Host to bind to

  Default: `0.0.0.0`

* `HDL_PORT`

  Port to bind to

  Default: `3000`

* `HDL_DSN`

  SQL DSN to the database of handle.net

  Default: `handle:handle@tcp(127.0.0.1:5432)/handle?sslmode=disable`

* `HDL_AUTH_USERNAME`

  Username for basic auth

  Default: `handle`

* `HDL_AUTH_PASSWORD`

  Password for basic auth

  Default: `handle`

# API

I've tried to be as close to the original API as possible.

See [Technical Manual Version 9](http://www.handle.net/tech_manual/HN_Tech_Manual_9.pdf)

All endpoint require basic http authentication.

See environment variable `HDL_AUTH_USERNAME` and `HDL_AUTH_PASSWORD` above

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

# Notes

* Environment can also indirectly be set by placing a file called `.env` in the current directory, and adding all environment variables (but WITHOUT keyword `export `!).