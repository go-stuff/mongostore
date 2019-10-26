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

The [github.com/go-stuff/web](https://github.com/go-stuff/web) repository uses [github.com/go-stuff/mongostore](https://github.com/go-stuff/mongostore) you can browse this example code to see how it is used.

```go
import "github.com/go-stuff/mongostore"
```

```go
// create a session store using rotating keys
mongoStore, err := mongostore.NewStore(
    client.Database(DatabaseName).Collection("sessions"),
    http.Cookie{
        Path:     "/",
        Domain:   "",
        MaxAge:   20 * 60, // 20 mins
        Secure:   false,
        HttpOnly: true,
        SameSite: http.SameSiteStrictMode,
    },
    []byte("new-authentication-key"),
    []byte("new-encryption-key"),
    []byte("old-authentication-key"),
    []byte("old-encryption-key"),
)
if err != nil {
    return fmt.Errorf("[ERROR] creating mongo store: %w", err)
}
```

```go
const CookieName = "session-id"

// new sessions
session, err := s.Store.New(r, CookieName)
if err != nil {
    log.Printf("[ERROR] new session: %s", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

// add values to the session
session.Values["username"] = r.FormValue("username")

// save session
err = session.Save(r, w)
if err != nil {
    log.Printf("[ERROR] saving session: %s", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

```go
const CookieName = "session-id"

// existing sessions
session, err := s.Store.Get(r, CookieName)
if err != nil {
    log.Printf("[ERROR] getting session: %s", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

// add values to the session
session.Values["username"] = r.FormValue("username")

// save session
err = session.Save(r, w)
if err != nil {
    log.Printf("[ERROR] saving session: %s", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

## Example MongoDB Entries

```bash
> db.sessions.find().pretty()
{
	"_id" : ObjectId("5db388c21256d9f65e6ccf7e"),
	"data" : {
		"username" : "test"
	},
	"modified_at" : ISODate("2019-10-25T23:44:03.558Z"),
	"expires_at" : ISODate("2019-10-26T00:04:03.558Z"),
	"ttl" : ISODate("2019-10-25T23:44:03.558Z")
}
{
	"_id" : ObjectId("5db388cb1256d9f65e6ccf7f"),
	"data" : {
		"username" : "user1"
	},
	"modified_at" : ISODate("2019-10-25T23:44:11.485Z"),
	"expires_at" : ISODate("2019-10-26T00:04:11.485Z"),
	"ttl" : ISODate("2019-10-25T23:44:11.485Z")
}

```

## License

[MIT License](LICENSE)
