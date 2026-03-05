// Package todoqueue provides a persistent, thread-safe priority queue for TODO items
// with SQLite storage, dependency resolution, and retry logic.
package todoqueue

import "time"

// Config holds the configuration for creating a new Queue.
type Config struct {
	// DBPath is the path to the SQLite database file.
	// If the file doesn't exist, it will be created.
	DBPath string

	// MaxRetries is the maximum number of times a failed item
	// will be automatically retried. Set to 0 to disable retries.
	MaxRetries int

	// RetryDelay is the duration to wait before retrying a failed item.
	// If not set, defaults to 5 minutes.
	RetryDelay time.Duration

	// AutoMigrate indicates whether to automatically run database
	// migrations when creating the queue. Defaults to true.
	AutoMigrate bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DBPath:      "./todos.db",
		MaxRetries:  3,
		RetryDelay:  5 * time.Minute,
		AutoMigrate: true,
	}
}

// Filter is used to query items from the queue.
type Filter struct {
	// Status filters by item status. Empty means all statuses.
	Status Status

	// MinPriority filters items with priority >= MinPriority.
	// Use 0 to include all priorities.
	MinPriority int

	// MaxPriority filters items with priority <= MaxPriority.
	// Use 0 to include all priorities (if MinPriority is also 0).
	MaxPriority int

	// HasDependencies filters items that have (or don't have) dependencies.
	// nil means don't filter by dependencies.
	HasDependencies *bool

	// Skills filters items that require specific skills.
	// Only items that require ALL specified skills will be returned.
	Skills []string

	// Limit limits the number of results returned.
	// 0 means no limit.
	Limit int

	// Offset skips the first N results.
	Offset int
}

// ListOptions is an alias for Filter for API clarity.
type ListOptions = Filter
