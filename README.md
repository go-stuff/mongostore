[![Build Status](https://travis-ci.com/go-stuff/mongostore.svg?branch=master)](https://travis-ci.com/go-stuff/mongostore)

# mongostore

An implementation of the [Gorilla web toolkit sessions](https://github.com/gorilla/sessions) for the [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)

## Requirements

- Gorilla web toolkit sessions [https://github.com/gorilla/sessions](https://github.com/gorilla/sessions)
- MongoDB Go Driver [https://github.com/mongodb/mongo-go-driver/](https://github.com/mongodb/mongo-go-driver/)

## Installation

The recommended way to get started using [mongostore](https://github.com/go-stuff/mongostore) is by using 'go get' to install the dependency in your project.

```go
go get "github.com/go-stuff/mongostore"
```

### Linux

Configure some environment variables for authentication and encryption:

```bash
openssl rand -hex 32
openssl rand -hex 16
```

Copy the results of the above two lines into environment variables:

```bash
export SESSION_AUTH_KEY=ResultOfOpenSslRandHex32
export SESSION_ENC_KEY=ResultOfOpenSslRandHex16
```

## Usage

```go
import (
    "github.com/go-stuff/mongostore"
)
```

## License

[MIT License](LICENSE)