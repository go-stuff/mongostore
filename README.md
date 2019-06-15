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

## Example Use

The [github.com/go-stuff/web](https://github.com/go-stuff/web) repository uses [github.com/go-stuff/mongostore](https://github.com/go-stuff/mongostore) you can browse this example code to see how it is used.

Initialize necessary environmemt variables, initialize the database, store and server.

```go
package main

import (
    ...
    "github.com/go-stuff/mongostore"
    ...
)

func main() {
    ...
    // init database
    client, ctx, err := initMongoClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    // set a default session ttl to 20 minutes
    if os.Getenv("MONGOSTORE_SESSION_TTL") == "" {
        os.Setenv("MONGOSTORE_SESSION_TTL", strconv.Itoa(20*60))
    }

    // get the ttl from an environment variable
    ttl, err := strconv.Atoi(os.Getenv("MONGOSTORE_SESSION_TTL"))
    if err != nil {
        log.Fatal(err)
    }

    // get database name from an environment variable
    if os.Getenv("MONGO_DB_NAME") == "" {
        os.Setenv("MONGO_DB_NAME", "test")
    }

    // init store
    store, err := initMongoStore(client.Database(os.Getenv("MONGO_DB_NAME")).Collection("sessions"), ttl)
    if err != nil {
        log.Fatal(err)
    }

    // init controllers
    router := controllers.Init(client, store)
    ...
    // init server
    server := &http.Server{
        Addr:           ":8080",
        Handler:        router,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20, // 1 MB
    }

    // start server
    log.Println("Listening and Serving @", server.Addr)
    err = server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }
}

func initMongoClient() (*mongo.Client, context.Context, error) {
    // a Context carries a deadline, cancelation signal, and request-scoped values
    // across API boundaries. Its methods are safe for simultaneous use by multiple
    // goroutines
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // use a default mongo uri if the MONGOURL environment variable is not set
    if os.Getenv("MONGO_URI") == "" {
        os.Setenv("MONGO_URI", "mongodb://localhost:27017")
    }

    // connect does not do server discovery, use ping
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
    if err != nil {
        return nil, nil, err
    }

    // ping for server discovery
    err = client.Ping(ctx, readpref.Primary())
    if err != nil {
        return nil, nil, err
    }

    log.Println("Connected to MongoDB @", os.Getenv("MONGO_URI"))
    return client, ctx, nil
}

func initMongoStore(col *mongo.Collection, age int) (*mongostore.MongoStore, error) {
    // generate an authentication key to use if the GORILLA_SESSION_AUTH_KEY environment
    // variable is not set
    if os.Getenv("GORILLA_SESSION_AUTH_KEY") == "" {
        os.Setenv(
            "GORILLA_SESSION_AUTH_KEY", 
            base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
        )
    }

    // generate an encryption key to use if the GORILLA_SESSION_ENC_KEY environment
    // variable is not set
    if os.Getenv("GORILLA_SESSION_ENC_KEY") == "" {
        os.Setenv(
            "GORILLA_SESSION_ENC_KEY", 
            base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16)),
        )
    }

    store := mongostore.NewMongoStore(
        col,
        age,
        []byte(os.Getenv("GORILLA_SESSION_AUTH_KEY")),
        []byte(os.Getenv("GORILLA_SESSION_ENC_KEY")),
    )

    return store, nil
}
```

Use some form of authentication before allowing a new session to be saved.

```go
package controllers

import (
    ...
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
    ...
    switch r.Method {
    case "POST":
        // start a new session
        session, err := store.New(r, "session")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // this is just an example, you can swap out authentication
        // with AD, LDAP, oAuth, etc...
        authenticatedUser := make(map[string]string)
        authenticatedUser["test"] = "test"
        authenticatedUser["user1"] = "password"
        authenticatedUser["user2"] = "password"
        authenticatedUser["user3"] = "password"

        user := models.User{}

        var found bool
        for k, v := range authenticatedUser {
            if r.FormValue("username") == k && r.FormValue("password") == v {
                user.Username = k
                found = true
            }
        }

        // user not found
        if !found {
            render(w, r, "login.html",
                struct {
                    CSRF     template.HTML
                    Username string
                    Error    error
                }{
                    CSRF:     csrf.TemplateField(r),
                    Username: r.FormValue("username"),
                    Error:    errors.New("credential error"),
            })
            return
        }

        // add important values to the session
        session.Values["remoteaddr"] = r.RemoteAddr
        session.Values["host"] = r.Host
        session.Values["username"] = user.Username

        // save the session
        err = session.Save(r, w)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
    ...
}
```

Get the session at the begining of each handler and save by then end. You always want to save the session even if it is unchanged, saving the session extends the life of the session.

```go
package controllers

import (
    "log"
    "net/http"
    "github.com/go-stuff/web/models"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
    // get session
    session, err := store.Get(r, "session")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // get data from session.Values
    user := &models.User{
        Username: session.Values["username"].(string),
    }

    // save session
    err = session.Save(r, w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // render to template
    render(w, r, "home.html",
        struct {
            User  *models.User
            Error error
        }{
            User:  user,
            Error: nil,
        })
}
```

## Example MongoDB Entries

```bash
> db.sessions.find().pretty()
{
    "_id" : "5d04eb4d92dc27bec40d7138",
    "remoteaddr" : "127.0.0.1:43690",
    "host" : "127.0.0.1:8080",
    "username" : "test",
    "createdAt" : ISODate("2019-06-15T12:57:49.458Z"),
    "modifiedAt" : ISODate("2019-06-15T12:57:49.526Z"),
    "expiresAt" : ISODate("2019-06-15T13:24:29.526Z")
}
{
    "_id" : "5d04eb8292dc27bec40d7139",
    "remoteaddr" : "127.0.0.1:43702",
    "host" : "127.0.0.1:8080",
    "username" : "user1",
    "createdAt" : ISODate("2019-06-15T12:58:42.241Z"),
    "modifiedAt" : ISODate("2019-06-15T12:58:42.252Z"),
    "expiresAt" : ISODate("2019-06-15T13:25:22.252Z")
}
```

## License

[MIT License](LICENSE)
