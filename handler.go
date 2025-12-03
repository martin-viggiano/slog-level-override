package slogleveloverride

import (
	"context"
	"log/slog"
	"sync/atomic"
)

var _ slog.Handler = (*OverrideHandler)(nil)

// New creates a new [OverrideHandler] wrapping the provided handler.
//
// Initially, no level override is set, and the underlying handler's
// Enabled method will be used to determine if logging is enabled.
func New(h slog.Handler) *OverrideHandler {
	return &OverrideHandler{
		basic:         h,
		assignedLevel: &atomic.Value{},
	}
}

// NewWithLevel creates a new [OverrideHandler] wrapping the provided handler
// with the specified [slog.Leveler] already set.
//
// The level is evaluated dynamically, allowing for runtime level changes.
func NewWithLevel(h slog.Handler, level slog.Leveler) *OverrideHandler {
	dynamicHandler := New(h)
	dynamicHandler.SetLevel(level)
	return dynamicHandler
}

// NewLoggerWithLevel wraps an existing [slog.Logger] with an [OverrideHandler]
// that overrides the level with the specified [slog.Leveler].
//
// The level is evaluated dynamically, allowing for runtime level changes.
// Returns a new [slog.Logger] with the wrapped handler.
func NewLoggerWithLevel(logger *slog.Logger, level slog.Leveler) *slog.Logger {
	handler := NewWithLevel(logger.Handler(), level)
	return slog.New(handler)
}

// OverrideHandler is an [slog.Handler] that wraps another handler and allows
// dynamic override of its log level filtering.
//
// When a level override is set via [SetLevel], the handler will evaluate the
// [slog.Leveler] on each logging operation, enabling runtime level changes.
// If no override is set, the handler delegates to the wrapped handler's Enabled method.
type OverrideHandler struct {
	basic         slog.Handler
	assignedLevel *atomic.Value
}

// SetLevel sets the level of an [slog.Handler] with the provided [slog.Leveler].
//
// The provided [slog.Leveler] is evaluated dynamically on each logging call,
// allowing for runtime level changes. This supports dynamic level implementations
// where the Level() method may return different values over time.
//
// Returns true if the operation was successful and false if the provided
// [slog.Handler] is not an [OverrideHandler] or if newLevel is nil.
func SetLevel(h slog.Handler, newLevel slog.Leveler) bool {
	if dlh, ok := h.(*OverrideHandler); ok && newLevel != nil {
		dlh.SetLevel(newLevel)
		return true
	}
	return false
}

// SetLevel sets a new level override for this handler.
//
// The provided [slog.Leveler] is stored and evaluated dynamically on each
// logging call, allowing the level to change at runtime. This method is
// thread-safe and can be called concurrently.
func (h *OverrideHandler) SetLevel(newLevel slog.Leveler) {
	h.assignedLevel.Store(newLevel)
}

// Level returns the current level of the handler by evaluating the assigned
// [slog.Leveler]. If no level override is set, it returns the level from the
// underlying handler if it implements the Level() method, otherwise returns 0.
func (h *OverrideHandler) Level() slog.Level {
	leveler := h.assignedLevel.Load()
	if leveler == nil {
		// Fallback to basic handler's level if it has one
		if l, ok := h.basic.(interface{ Level() slog.Level }); ok {
			return l.Level()
		}
		return 0
	}
	return leveler.(slog.Leveler).Level()
}

// Handle forwards the record to the underlying handler without modification.
func (h *OverrideHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.basic.Handle(ctx, record)
}

// Enabled determines if logging is enabled for the given level.
//
// If a level override is set, it evaluates the [slog.Leveler] dynamically
// to get the current threshold level. If no override is set, it delegates
// to the underlying handler's Enabled method.
func (h *OverrideHandler) Enabled(ctx context.Context, level slog.Level) bool {
	leveler := h.assignedLevel.Load()
	if leveler == nil {
		return h.basic.Enabled(ctx, level)
	}
	return level >= leveler.(slog.Leveler).Level()
}

// WithAttrs returns a new [OverrideHandler] with the given attributes added.
//
// The new handler shares the same level override as the parent handler,
// meaning changes to the level will be reflected in both handlers.
func (h *OverrideHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newLevel := &atomic.Value{}
	newLevel.Store(h.assignedLevel.Load())

	return &OverrideHandler{
		basic:         h.basic.WithAttrs(attrs),
		assignedLevel: newLevel,
	}
}

// WithGroup returns a new [OverrideHandler] with the given group name added.
//
// The new handler shares the same level override as the parent handler,
// meaning changes to the level will be reflected in both handlers.
func (h *OverrideHandler) WithGroup(name string) slog.Handler {
	newLevel := &atomic.Value{}
	newLevel.Store(h.assignedLevel.Load())

	return &OverrideHandler{
		basic:         h.basic.WithGroup(name),
		assignedLevel: newLevel,
	}
}
