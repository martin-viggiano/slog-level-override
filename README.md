# slog-level-override

A Go library that provides dynamic log level override for `slog.Handler`.

## Overview

This package allows you to override the log level of any `slog.Handler` at runtime. The level is evaluated dynamically on each logging call, enabling you to change log levels on-the-fly without restarting your application.

## Installation

```bash
go get github.com/martin-viggiano/slog-level-override
```

## Usage

```go
package main

import (
    "log/slog"
    "os"
    slogleveloverride "github.com/martin-viggiano/slog-level-override"
)

func main() {
    // Wrap an existing handler
    handler := slogleveloverride.New(slog.NewJSONHandler(os.Stdout, nil))
    logger := slog.New(handler)

    // Change the level at runtime
    handler.SetLevel(slog.LevelWarn)
    
    // Or create with an initial level
    handler2 := slogleveloverride.NewWithLevel(slog.NewJSONHandler(os.Stdout, nil), slog.LevelInfo)
    logger2 := slog.New(handler2)
}
```

## Inspiration

This project was inspired by [gekatateam/dynamic-level-handler](https://github.com/gekatateam/dynamic-level-handler).
