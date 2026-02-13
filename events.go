package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventEmitter provides methods to emit events to the frontend
type EventEmitter struct {
	ctx context.Context
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(ctx context.Context) *EventEmitter {
	return &EventEmitter{ctx: ctx}
}

// NotificationData represents a notification event
type NotificationData struct {
	Type     string `json:"type"`     // success, error, info, warning
	Title    string `json:"title"`    // Main message
	Message  string `json:"message"`  // Optional details
	Duration int    `json:"duration"` // Auto-dismiss duration in ms (0 = manual)
}

// ProgressData represents a progress event
type ProgressData struct {
	ID        string `json:"id"`        // Unique operation ID
	Operation string `json:"operation"` // Operation name
	Current   int64  `json:"current"`   // Current progress
	Total     int64  `json:"total"`     // Total items/bytes
	Filename  string `json:"filename"`  // Optional filename being processed
}

// EmitNotification sends a notification to the frontend
func (e *EventEmitter) EmitNotification(notif NotificationData) {
	runtime.EventsEmit(e.ctx, "notification", notif)
}

// EmitProgress sends a progress update to the frontend
func (e *EventEmitter) EmitProgress(progress ProgressData) {
	runtime.EventsEmit(e.ctx, "progress", progress)
}

// EmitProgressComplete signals that an operation is complete
func (e *EventEmitter) EmitProgressComplete(id string) {
	runtime.EventsEmit(e.ctx, "progress:complete", map[string]string{"id": id})
}

// Convenience methods for common notification types
func (e *EventEmitter) Success(title, message string) {
	e.EmitNotification(NotificationData{
		Type:     "success",
		Title:    title,
		Message:  message,
		Duration: 5000,
	})
}

func (e *EventEmitter) Error(title, message string) {
	e.EmitNotification(NotificationData{
		Type:     "error",
		Title:    title,
		Message:  message,
		Duration: 8000,
	})
}

func (e *EventEmitter) Info(title, message string) {
	e.EmitNotification(NotificationData{
		Type:     "info",
		Title:    title,
		Message:  message,
		Duration: 5000,
	})
}

func (e *EventEmitter) Warning(title, message string) {
	e.EmitNotification(NotificationData{
		Type:     "warning",
		Title:    title,
		Message:  message,
		Duration: 6000,
	})
}
