package slogleveloverride

import (
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/thejerf/slogassert"
)

// dynamicLevel is a test Leveler that can change its level at runtime
type dynamicLevel struct {
	level atomic.Int64
}

func newDynamicLevel(initial slog.Level) *dynamicLevel {
	dl := &dynamicLevel{}
	dl.level.Store(int64(initial))
	return dl
}

func (d *dynamicLevel) Level() slog.Level {
	return slog.Level(d.level.Load())
}

func (d *dynamicLevel) SetLevel(level slog.Level) {
	d.level.Store(int64(level))
}

// TestNew verifies that New creates a handler without level override
func TestNew(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := New(assertHandler)
	logger := slog.New(handler)

	// Without level override, should use underlying handler's level
	logger.Debug("debug message")
	logger.Info("info message")

	// slogassert default level is Info, so debug should not appear
	assertHandler.AssertMessage("info message")
}

// TestNewWithLevel verifies that NewWithLevel creates a handler with initial level
func TestNewWithLevel(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := NewWithLevel(assertHandler, slog.LevelWarn)
	logger := slog.New(handler)

	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Only warn and error should appear
	assertHandler.AssertMessage("warn message")
	assertHandler.AssertMessage("error message")
}

// TestNewLoggerWithLevel verifies that NewLoggerWithLevel wraps a logger
func TestNewLoggerWithLevel(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	baseLogger := slog.New(assertHandler)
	logger := NewLoggerWithLevel(baseLogger, slog.LevelError)

	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Only error should appear
	assertHandler.AssertMessage("error message")
}

// TestSetLevelMethod verifies that the SetLevel method changes the level
func TestSetLevelMethod(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := New(assertHandler)
	logger := slog.New(handler)

	// Start with Warn level
	handler.SetLevel(slog.LevelWarn)
	logger.Info("info 1")
	logger.Warn("warn 1")

	// Change to Info level
	handler.SetLevel(slog.LevelInfo)
	logger.Info("info 2")
	logger.Warn("warn 2")

	// Should see warn 1, info 2, and warn 2
	assertHandler.AssertMessage("warn 1")
	assertHandler.AssertMessage("info 2")
	assertHandler.AssertMessage("warn 2")
}

// TestSetLevelFunction verifies that the standalone SetLevel function works
func TestSetLevelFunction(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := New(assertHandler)
	logger := slog.New(handler)

	// Use standalone function
	result := SetLevel(handler, slog.LevelWarn)
	if !result {
		t.Fatal("SetLevel should return true for OverrideHandler")
	}

	logger.Info("info message")
	logger.Warn("warn message")

	assertHandler.AssertMessage("warn message")
}

// TestSetLevelFunctionReturnsFalse verifies that SetLevel returns false for non-OverrideHandler
func TestSetLevelFunctionReturnsFalse(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	// Try to set level on a non-OverrideHandler
	result := SetLevel(assertHandler, slog.LevelWarn)
	if result {
		t.Fatal("SetLevel should return false for non-OverrideHandler")
	}

	// Verify we can still log normally
	logger := slog.New(assertHandler)
	logger.Info("info message")
	assertHandler.AssertMessage("info message")
}

// TestSetLevelFunctionWithNilLevel verifies that SetLevel returns false with nil level
func TestSetLevelFunctionWithNilLevel(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := New(assertHandler)
	result := SetLevel(handler, nil)
	if result {
		t.Fatal("SetLevel should return false when newLevel is nil")
	}
}

// TestDynamicLeveler verifies that dynamic level changes are reflected immediately
func TestDynamicLeveler(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	dynamicLvl := newDynamicLevel(slog.LevelWarn)
	handler := NewWithLevel(assertHandler, dynamicLvl)
	logger := slog.New(handler)

	// Start at Warn level
	logger.Info("info 1")
	logger.Warn("warn 1")

	// Change to Info level
	dynamicLvl.SetLevel(slog.LevelInfo)
	logger.Info("info 2")
	logger.Warn("warn 2")

	// Change to Error level
	dynamicLvl.SetLevel(slog.LevelError)
	logger.Info("info 3")
	logger.Warn("warn 3")
	logger.Error("error 1")

	// Should see: warn 1, info 2, warn 2, error 1
	assertHandler.AssertMessage("warn 1")
	assertHandler.AssertMessage("info 2")
	assertHandler.AssertMessage("warn 2")
	assertHandler.AssertMessage("error 1")
}

// TestWithAttrs verifies that WithAttrs creates a new handler that shares level override
func TestWithAttrs(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := NewWithLevel(assertHandler, slog.LevelWarn)
	logger := slog.New(handler)

	// Create a derived logger with attributes
	derivedLogger := logger.With("component", "test")

	logger.Info("info from parent")
	logger.Warn("warn from parent")
	derivedLogger.Info("info from derived")
	derivedLogger.Warn("warn from derived")

	// Both should respect the Warn level
	assertHandler.AssertMessage("warn from parent")
	assertHandler.AssertMessage("warn from derived")
}

// TestWithGroup verifies that WithGroup creates a new handler that shares level override
func TestWithGroup(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := NewWithLevel(assertHandler, slog.LevelWarn)
	logger := slog.New(handler)

	// Create a derived logger with group
	derivedLogger := logger.WithGroup("mygroup")

	logger.Info("info from parent")
	logger.Warn("warn from parent")
	derivedLogger.Info("info from derived")
	derivedLogger.Warn("warn from derived")

	// Both should respect the Warn level
	assertHandler.AssertMessage("warn from parent")
	assertHandler.AssertMessage("warn from derived")
}

// TestLevelChangesPropagateToDerivered verifies that level changes affect derived handlers
func TestLevelChangesPropagateToDerivered(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	dynamicLvl := newDynamicLevel(slog.LevelWarn)
	handler := NewWithLevel(assertHandler, dynamicLvl)
	logger := slog.New(handler)
	derivedLogger := logger.With("component", "derived")

	// Start at Warn
	logger.Info("info 1")
	derivedLogger.Warn("warn 1")

	// Change to Info
	dynamicLvl.SetLevel(slog.LevelInfo)
	logger.Info("info 2")
	derivedLogger.Info("info 3")

	// Should see warn 1, info 2, info 3
	assertHandler.AssertMessage("warn 1")
	assertHandler.AssertMessage("info 2")
	assertHandler.AssertMessage("info 3")
}

// TestHandleForwardsToUnderlying verifies that Handle() forwards to underlying handler
func TestHandleForwardsToUnderlying(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := NewWithLevel(assertHandler, slog.LevelInfo)
	logger := slog.New(handler)

	logger.Info("test message", "key", "value", "number", 42)

	// Verify the message and attributes made it through
	assertHandler.AssertMessage("test message")
}

// TestNoOverrideDelegatesToUnderlying verifies that without override, underlying handler's Enabled is used
func TestNoOverrideDelegatesToUnderlying(t *testing.T) {
	// Create a handler with Warn level
	assertHandler := slogassert.New(t, slog.LevelWarn, nil)
	defer assertHandler.AssertEmpty()

	// Wrap it without setting a level override
	handler := New(assertHandler)
	logger := slog.New(handler)

	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Should only see warn and error (underlying handler's level)
	assertHandler.AssertMessage("warn message")
	assertHandler.AssertMessage("error message")
}

// TestConcurrentSetLevel verifies thread-safety of concurrent SetLevel calls
func TestConcurrentSetLevel(t *testing.T) {
	assertHandler := slogassert.New(t, slog.LevelInfo, nil)
	defer assertHandler.AssertEmpty()

	handler := New(assertHandler)
	logger := slog.New(handler)

	// Set level concurrently from multiple goroutines
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(n int) {
			handler.SetLevel(slog.LevelInfo)
			logger.Info("concurrent message")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should have 3 messages
	assertHandler.AssertMessage("concurrent message")
	assertHandler.AssertMessage("concurrent message")
	assertHandler.AssertMessage("concurrent message")
}
