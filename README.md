# mongostore

[![Build Status](https://travis-ci.com/go-stuff/mongostore.svg?branch=master)](https://travis-ci.com/go-stuff/mongostore)
[![codecov](https://codecov.io/gh/go-stuff/mongostore/branch/master/graph/badge.svg)](https://codecov.io/gh/go-stuff/mongostore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![Gopher Share](https://github.com/go-stuff/images/blob/master/GOPHER_SHARE_640x320.png)

An implementation of Store using [Gorilla web toolkit sessions](https://github.com/gorilla/sessions) and the [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)

## Requirements

- Gorilla web toolkit sessions [https://github.com/gorilla/sessions](https://github.com/gorilla/sessions)
- MongoDB Go Driver [https://github.com/mongodb/mongo-go-driver/](https://github.com/mongodb/mongo-go-driver/)

## Installation

The recommended way to get started using [mongostore](https://github.com/go-stuff/mongostore) is by using 'go get' to install the dependency in your project.

```go
go get "github.com/go-stuff/mongostore"
```

## Usage

```go
import (
    "github.com/go-stuff/mongostore"
)
```

## License

[MIT License](LICENSE)