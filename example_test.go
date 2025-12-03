package slogleveloverride_test

import (
	"log/slog"
	"os"

	slogleveloverride "github.com/martin-viggiano/slog-level-override"
)

// Example demonstrates basic usage of OverrideHandler
func Example() {
	// Create a base handler without timestamp
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: removeTime,
	})

	// Wrap it with OverrideHandler
	handler := slogleveloverride.New(baseHandler)
	logger := slog.New(handler)

	// Initially uses underlying handler's level (Info)
	logger.Debug("This won't appear")
	logger.Info("Initial info message")

	// Override to Warn level
	handler.SetLevel(slog.LevelWarn)
	logger.Info("This won't appear either")
	logger.Warn("This will appear")

	// Output:
	// level=INFO msg="Initial info message"
	// level=WARN msg="This will appear"
}

// ExampleNewWithLevel demonstrates creating a handler with an initial level
func ExampleNewWithLevel() {
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelDebug, // Base handler accepts all levels
		ReplaceAttr: removeTime,
	})

	// Create handler with Warn level from the start
	handler := slogleveloverride.NewWithLevel(baseHandler, slog.LevelWarn)
	logger := slog.New(handler)

	logger.Info("This won't appear")
	logger.Warn("This will appear")

	// Output:
	// level=WARN msg="This will appear"
}

// ExampleNewLoggerWithLevel demonstrates wrapping an existing logger
func ExampleNewLoggerWithLevel() {
	baseLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: removeTime,
	}))

	// Wrap the existing logger with level override
	logger := slogleveloverride.NewLoggerWithLevel(baseLogger, slog.LevelError)

	logger.Info("This won't appear")
	logger.Warn("This won't appear either")
	logger.Error("This will appear")

	// Output:
	// level=ERROR msg="This will appear"
}

// ExampleSetLevel demonstrates using the standalone SetLevel function
func ExampleSetLevel() {
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: removeTime,
	})
	handler := slogleveloverride.New(baseHandler)
	logger := slog.New(handler)

	// Use standalone function to set level
	success := slogleveloverride.SetLevel(handler, slog.LevelWarn)
	if success {
		logger.Info("This won't appear")
		logger.Warn("This will appear")
	}

	// Output:
	// level=WARN msg="This will appear"
}

// ExampleOverrideHandler_WithAttrs demonstrates that derived loggers share the level override
func ExampleOverrideHandler_WithAttrs() {
	baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: removeTime,
	})
	handler := slogleveloverride.NewWithLevel(baseHandler, slog.LevelWarn)
	logger := slog.New(handler)

	// Create derived logger with attributes
	componentLogger := logger.With("component", "database")

	// Both respect the same level
	logger.Info("Won't appear")
	logger.Warn("Parent warning")
	componentLogger.Warn("Component warning")

	// Output:
	// level=WARN msg="Parent warning"
	// level=WARN msg="Component warning" component=database
}

// removeTime is a helper function that removes the time attribute from log output
func removeTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
