// Package cron provides a cron-based scheduler for AI agents.
// It supports standard cron syntax, human-readable aliases, timezone handling,
// and persistent run history tracking via SQLite.
package cron

import (
	"errors"
	"time"
)

// Config holds the scheduler configuration.
type Config struct {
	// DBPath is the path to the SQLite database file for storing run history.
	// If empty, an in-memory database will be used.
	DBPath string

	// DefaultTimezone is the default timezone for jobs that don't specify one.
	// Defaults to UTC if empty or invalid.
	DefaultTimezone string

	// RetryOnFailure enables automatic retry for failed job executions.
	RetryOnFailure bool

	// MaxRetries is the maximum number of retry attempts for failed jobs.
	// Only used when RetryOnFailure is true.
	MaxRetries int

	// MissedRunBehavior defines how to handle missed runs on startup.
	// Options: "run" (run missed jobs), "skip" (skip missed jobs), "ignore"
	MissedRunBehavior string

	// Logger is an optional logger for debugging and monitoring.
	// If nil, no logging is performed.
	Logger Logger
}

// Logger interface for custom logging implementations.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a log field key-value pair.
type Field struct {
	Key   string
	Value interface{}
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DBPath:            "./cron.db",
		DefaultTimezone:   "UTC",
		RetryOnFailure:    false,
		MaxRetries:        3,
		MissedRunBehavior: "skip",
		Logger:            nil,
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.MaxRetries < 0 {
		return errors.New("MaxRetries cannot be negative")
	}

	validMissedRunBehaviors := map[string]bool{
		"run":    true,
		"skip":   true,
		"ignore": true,
		"":       true, // empty means use default
	}
	if !validMissedRunBehaviors[c.MissedRunBehavior] {
		return errors.New("invalid MissedRunBehavior: must be 'run', 'skip', or 'ignore'")
	}

	return nil
}

// GetTimezone returns the configured timezone as a *time.Location.
// Falls back to UTC if the timezone is invalid or not specified.
func (c *Config) GetTimezone() *time.Location {
	if c.DefaultTimezone == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(c.DefaultTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
