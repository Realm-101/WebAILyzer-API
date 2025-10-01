# Logging Package

This package provides a structured logging interface for the WebAILyzer API using logrus as the underlying implementation.

## Features

- Multiple log levels (trace, debug, info, warn, error, fatal)
- JSON and text formatting options
- Configurable output destinations (stdout, stderr, or file)
- Structured logging with fields
- Context support
- Global logger instance for convenience

## Usage

### Basic Usage

```go
package main

import (
    "github.com/projectdiscovery/wappalyzergo/internal/logging"
)

func main() {
    // Use default configuration
    logger, err := logging.NewLogger(logging.DefaultConfig())
    if err != nil {
        panic(err)
    }

    logger.Info("Application started")
    logger.WithField("user_id", 123).Info("User logged in")
}
```

### Custom Configuration

```go
config := &logging.Config{
    Level:      logging.LogLevelDebug,
    Format:     "json",
    Output:     "stdout",
    TimeFormat: "2006-01-02T15:04:05Z07:00",
}

logger, err := logging.NewLogger(config)
if err != nil {
    panic(err)
}
```

### Global Logger

```go
// Initialize global logger
err := logging.InitGlobalLogger(config)
if err != nil {
    panic(err)
}

// Use global functions
logging.Info("This uses the global logger")
logging.WithField("component", "auth").Warn("Authentication warning")
```

## Configuration Options

- `Level`: Log level (trace, debug, info, warn, error, fatal)
- `Format`: Output format ("json" or "text")
- `Output`: Output destination ("stdout", "stderr", or file path)
- `TimeFormat`: Time format for log entries (Go time format)

## Log Levels

- `LogLevelTrace`: Very detailed logs
- `LogLevelDebug`: Debug information
- `LogLevelInfo`: General information
- `LogLevelWarn`: Warning messages
- `LogLevelError`: Error messages
- `LogLevelFatal`: Fatal errors (will exit the program)