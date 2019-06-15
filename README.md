# mongostore

[![GoDoc](https://godoc.org/github.com/go-stuff/mongostore?status.svg)](https://godoc.org/github.com/go-stuff/mongostore)
[![Build Status](https://cloud.drone.io/api/badges/go-stuff/mongostore/status.svg)](https://cloud.drone.io/go-stuff/mongostore)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-stuff/mongostore)](https://goreportcard.com/report/github.com/go-stuff/mongostore)
[![codecov](https://codecov.io/gh/go-stuff/mongostore/branch/master/graph/badge.svg)](https://codecov.io/gh/go-stuff/mongostore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![Gopher Share](https://github.com/go-stuff/images/blob/master/GOPHER_SHARE_640x320.png)

An implementation of Store using [Gorilla web toolkit](https://github.com/gorilla) and the [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)

## Packages Imported

- Gorilla web tookit [github.com/gorilla](https://github.com/gorilla)
  - [securecookie](https://github.com/gorilla/securecookie)
  - [sessions](https://github.com/gorilla/sessions)
- MongoDB Go Driver [github.com/mongodb/mongo-go-driver](https://github.com/mongodb/mongo-go-driver)
  - [bson](https://github.com/mongodb/mongo-go-driver/bson)
  - [bson/primitive](https://github.com/mongodb/mongo-go-driver/bson/primitive)
  - [mongo](https://github.com/mongodb/mongo-go-driver/mongo)
  - [mongo/options](https://github.com/mongodb/mongo-go-driver/mongo/options)
  - [x/bsonx](https://github.com/mongodb/mongo-go-driver/x/bsonx)

## Installation

The recommended way to get started using [github.com/go-stuff/mongostore](https://github.com/go-stuff/mongostore) is by using 'go get' to install the dependency in your project.

```bash
go get "github.com/go-stuff/mongostore"
```

## Usage

```go
import (
    "github.com/go-stuff/mongostore"
)
```

## Environment Variables

The `go-stuff\mongostore` package uses two environment variables for authentication and encryption. If those keys are not alerady part of the `environment`, they will be generated each time `NewMongoStore` is run.  

A better solution would be to add permanent keys to the `environment`. The [gorilla toolkit](https://www.gorillatoolkit.org/pkg/securecookie#GenerateRandomKey) provides a way to generate random keys using [func GenerateRandomKey](https://www.gorillatoolkit.org/pkg/securecookie#GenerateRandomKey).

```bash
GORILLA_SESSION_AUTH_KEY (32 bytes)
GORILLA_SESSION_ENC_KEY  (16 bytes)
```

Once the keys added to the `environment`, sessions and cookies will be maintained each time `NewMongoStore` is run. Messages like `securecookie: the value is not valid` will be avoided, this happens when previously created cookies, still in the browser, used different authentication and encryption keys.

## Example

The [github.com/go-stuff/web](https://github.com/go-stuff/web) repository uses [github.com/go-stuff/mongostore](https://github.com/go-stuff/mongostore) you can browse this example code to see how it is used.

## License

[MIT License](LICENSE)
