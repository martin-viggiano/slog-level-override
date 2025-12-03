# slog-level-override

A Go library that provides dynamic log level override for `slog.Handler`.

## Overview

This package allows you to override the log level of any `slog.Handler` at runtime. The level is evaluated dynamically on each logging call, enabling you to change log levels on-the-fly without restarting your application.

## Installation

```bash
go get github.com/martin-viggiano/slog-level-override
```

## Usage

### Basic Usage

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

    // Initially uses underlying handler's level
    logger.Info("This will appear")
    
    // Change the level at runtime
    handler.SetLevel(slog.LevelWarn)
    logger.Info("This won't appear")
    logger.Warn("This will appear")
}
```

### Create with Initial Level

```go
// Create handler with a level set from the start
handler := slogleveloverride.NewWithLevel(
    slog.NewJSONHandler(os.Stdout, nil),
    slog.LevelWarn,
)
logger := slog.New(handler)

logger.Info("This won't appear")
logger.Warn("This will appear")
```

### Wrap an Existing Logger

```go
// Wrap an existing logger
baseLogger := slog.Default()
logger := slogleveloverride.NewLoggerWithLevel(baseLogger, slog.LevelError)

logger.Info("This won't appear")
logger.Error("This will appear")
```

### Dynamic Level Changes

```go
// Create a custom Leveler that can change at runtime
type DynamicLevel struct {
    level atomic.Int64
}

func (d *DynamicLevel) Level() slog.Level {
    return slog.Level(d.level.Load())
}

func (d *DynamicLevel) SetLevel(level slog.Level) {
    d.level.Store(int64(level))
}

// Use it with the handler
dynamicLevel := &DynamicLevel{}
dynamicLevel.SetLevel(slog.LevelWarn)

handler := slogleveloverride.NewWithLevel(
    slog.NewJSONHandler(os.Stdout, nil),
    dynamicLevel,
)
logger := slog.New(handler)

// Change level at runtime
dynamicLevel.SetLevel(slog.LevelInfo)
logger.Info("Now this will appear")
```

### Using Standalone SetLevel Function

```go
handler := slogleveloverride.New(slog.NewJSONHandler(os.Stdout, nil))
logger := slog.New(handler)

// Use standalone function
if slogleveloverride.SetLevel(handler, slog.LevelWarn) {
    logger.Warn("Level was set successfully")
}
```

## ⚠️ Important: Handler Wrapping Order

When wrapping multiple `slog.Handler` implementations, **`OverrideHandler` must be the outermost (last) wrapper** for level overrides to work correctly.

### Correct Order

```go
// First wrap with other handlers, then OverrideHandler
baseHandler := slog.NewJSONHandler(os.Stdout, nil)
wrappedHandler := someOtherHandler.Wrap(baseHandler)
overrideHandler := slogleveloverride.New(wrappedHandler)
logger := slog.New(overrideHandler)
```

### Incorrect Order

```go
// DON'T wrap OverrideHandler with other handlers
baseHandler := slog.NewJSONHandler(os.Stdout, nil)
overrideHandler := slogleveloverride.New(baseHandler)
wrappedHandler := someOtherHandler.Wrap(overrideHandler) // Wrong!
logger := slog.New(wrappedHandler)
```

The `OverrideHandler` controls level filtering through its `Enabled()` method. If another handler wraps it, that handler's `Enabled()` method will be called first, potentially bypassing the level override.

## Inspiration

This project was inspired by [gekatateam/dynamic-level-handler](https://github.com/gekatateam/dynamic-level-handler).
