# go-todo-queue

A persistent, thread-safe priority queue for TODO items with SQLite storage, dependency resolution, and retry logic.

## Features

- **Priority Queue**: Higher priority items are processed first
- **SQLite Persistence**: All items are stored in SQLite for durability
- **Dependency Resolution**: Tasks can depend on other tasks
- **State Machine**: Proper state transitions (pending → in-progress → completed/failed)
- **Retry Logic**: Automatic retry for failed items with configurable max retries
- **Thread-Safe**: Safe for concurrent use with `sync.RWMutex`
- **Simple API**: Clean, intuitive interface

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/todo-queue
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    todo "github.com/formatho/agent-orchestrator/packages/todo-queue"
)

func main() {
    // Create a new queue with SQLite backend
    queue, err := todo.New(todo.Config{
        DBPath:     "./todos.db",
        MaxRetries: 3,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer queue.Close()

    // Add a new TODO item
    item, err := queue.Add("Implement user authentication")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created item: %s\n", item.ID)

    // Get the next pending item (highest priority)
    next, err := queue.Next()
    if err != nil {
        log.Fatal(err)
    }

    // Start processing
    if err := queue.Start(next.ID); err != nil {
        log.Fatal(err)
    }

    // Complete the item
    if err := queue.Complete(next.ID, "Authentication implemented"); err != nil {
        log.Fatal(err)
    }
}
```

## Usage Examples

### Creating Items with Options

```go
// Simple item
item, _ := queue.Add("Simple task")

// With priority (higher = more important)
item, _ := queue.Add("Urgent task", todo.WithPriority(10))

// With dependencies
parentTask, _ := queue.Add("Parent task")
childTask, _ := queue.Add("Child task", todo.WithDependencies(parentTask.ID))

// With required skills
item, _ := queue.Add("Code review", todo.WithSkills("go", "review"))

// With custom metadata
item, _ := queue.Add("Task", todo.WithMetadata(map[string]interface{}{
    "assignee": "alice",
    "tags":     []string{"backend", "urgent"},
}))

// Combine multiple options
item, _ := queue.Add("Complex task",
    todo.WithPriority(8),
    todo.WithSkills("go", "database"),
    todo.WithMetadata(map[string]interface{}{"project": "formatho"}),
)
```

### Processing Items

```go
// Get highest priority pending item
next, err := queue.Next()
if err != nil {
    log.Fatal(err)
}
if next == nil {
    fmt.Println("No items to process")
    return
}

// Mark as in-progress
if err := queue.Start(next.ID); err != nil {
    log.Fatal(err)
}

// Do the work...
result := doWork(next)

// Mark as completed
if err := queue.Complete(next.ID, result); err != nil {
    log.Fatal(err)
}

// Or if it fails
if err := queue.Fail(next.ID, "Something went wrong"); err != nil {
    log.Fatal(err)
}
```

### Working with Dependencies

```go
// Create a dependency chain
task1, _ := queue.Add("Setup database")
task2, _ := queue.Add("Create migrations", todo.WithDependencies(task1.ID))
task3, _ := queue.Add("Run migrations", todo.WithDependencies(task2.ID))

// Check if dependencies are met
ready, unmet, err := queue.CheckDependencies(task2.ID)
if err != nil {
    log.Fatal(err)
}
if !ready {
    fmt.Printf("Waiting for: %v\n", unmet)
}

// Next() respects dependencies - task2 won't be returned until task1 is done
next, _ := queue.Next() // Returns task1
queue.Start(next.ID)
queue.Complete(next.ID, "DB setup done")

next, _ = queue.Next() // Now returns task2
```

### Blocking Items

```go
// Block an item waiting for external resource
if err := queue.Block(item.ID, "Waiting for API key"); err != nil {
    log.Fatal(err)
}

// Unblock when ready
if err := queue.Unblock(item.ID); err != nil {
    log.Fatal(err)
}
```

### Retry Failed Items

```go
// Automatic retry (configured via MaxRetries)
// Items that fail are automatically moved back to pending status
// until MaxRetries is reached

// Manual retry
if err := queue.Retry(item.ID); err != nil {
    log.Fatal(err)
}
```

### Listing and Filtering

```go
// List all items
items, _ := queue.List(todo.Filter{})

// List only pending items
items, _ := queue.List(todo.Filter{Status: todo.StatusPending})

// List with priority range
items, _ := queue.List(todo.Filter{
    MinPriority: 5,
    MaxPriority: 10,
})

// List with pagination
items, _ := queue.List(todo.Filter{
    Limit:  20,
    Offset: 40,
})

// List items with dependencies
hasDeps := true
items, _ := queue.List(todo.Filter{HasDependencies: &hasDeps})
```

### Getting Statistics

```go
stats, err := queue.Stats()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total: %d\n", stats.Total)
fmt.Printf("Pending: %d\n", stats.Pending)
fmt.Printf("In Progress: %d\n", stats.InProgress)
fmt.Printf("Completed: %d\n", stats.Completed)
fmt.Printf("Failed: %d\n", stats.Failed)
fmt.Printf("Blocked: %d\n", stats.Blocked)
```

### Worker Pattern

```go
func worker(queue *todo.Queue, workerID int) {
    for {
        // Get next item
        item, err := queue.Next()
        if err != nil {
            log.Printf("Worker %d: error getting next: %v", workerID, err)
            time.Sleep(time.Second)
            continue
        }
        
        if item == nil {
            // No items available, wait and retry
            time.Sleep(5 * time.Second)
            continue
        }
        
        // Start processing
        if err := queue.Start(item.ID); err != nil {
            continue // Item may have been claimed by another worker
        }
        
        log.Printf("Worker %d: processing %s", workerID, item.Description)
        
        // Do the work
        result, err := processItem(item)
        if err != nil {
            queue.Fail(item.ID, err.Error())
        } else {
            queue.Complete(item.ID, result)
        }
    }
}

// Start multiple workers
for i := 0; i < 5; i++ {
    go worker(queue, i)
}
```

## API Reference

### Types

#### `Queue`

Main queue struct that manages TODO items.

#### `Item`

Represents a single TODO item:
- `ID` - Unique identifier
- `Priority` - Priority level (higher = more important)
- `Description` - Human-readable description
- `Status` - Current status (pending, in-progress, completed, failed, blocked)
- `Dependencies` - List of item IDs that must complete first
- `SkillsRequired` - Skills needed to process this item
- `Result` - Result data when completed
- `Error` - Error message when failed
- `RetryCount` - Number of times this has been retried
- `CreatedAt`, `StartedAt`, `CompletedAt`, `UpdatedAt` - Timestamps
- `Metadata` - Custom key-value data

#### `Status`

Item status constants:
- `StatusPending` - Waiting to be processed
- `StatusInProgress` - Currently being processed
- `StatusCompleted` - Successfully completed
- `StatusFailed` - Failed to complete
- `StatusBlocked` - Blocked by external factors

#### `Config`

Queue configuration:
- `DBPath` - Path to SQLite database file
- `MaxRetries` - Maximum retry attempts for failed items
- `RetryDelay` - Time to wait before retrying
- `AutoMigrate` - Automatically run database migrations

#### `Filter` / `ListOptions`

Query filter options:
- `Status` - Filter by status
- `MinPriority`, `MaxPriority` - Priority range
- `HasDependencies` - Filter by dependency presence
- `Skills` - Filter by required skills
- `Limit`, `Offset` - Pagination

### Queue Methods

| Method | Description |
|--------|-------------|
| `New(config)` | Create a new queue |
| `Close()` | Close the queue and database |
| `Add(description, opts...)` | Add a new TODO item |
| `Next()` | Get highest priority pending item |
| `Start(id)` | Mark item as in-progress |
| `Complete(id, result)` | Mark item as completed |
| `Fail(id, error)` | Mark item as failed |
| `Block(id, reason)` | Mark item as blocked |
| `Unblock(id)` | Unblock an item |
| `Retry(id)` | Manually retry a failed item |
| `List(filter)` | List items with filters |
| `Get(id)` | Get specific item |
| `Delete(id)` | Delete an item |
| `Update(id, updates)` | Update item fields |
| `CheckDependencies(id)` | Check if dependencies are met |
| `Stats()` | Get queue statistics |

### Functional Options

| Option | Description |
|--------|-------------|
| `WithPriority(n)` | Set item priority |
| `WithID(id)` | Set custom ID |
| `WithStatus(status)` | Set initial status |
| `WithDependencies(ids...)` | Set dependencies |
| `WithSkills(skills...)` | Set required skills |
| `WithMetadata(map)` | Set custom metadata |

## State Transitions

```
pending → in-progress → completed
    ↓          ↓
blocked      failed → pending (retry)
```

- `pending` → `in-progress`: When starting work
- `pending` → `blocked`: When blocking item
- `in-progress` → `completed`: On success
- `in-progress` → `failed`: On failure
- `failed` → `pending`: Automatic retry (if under max)
- `blocked` → `pending`: When unblocked

## Thread Safety

The queue is thread-safe and can be used from multiple goroutines simultaneously. All operations use appropriate read/write locks.

## Database Schema

```sql
CREATE TABLE todo_items (
    id TEXT PRIMARY KEY,
    priority INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    dependencies TEXT,          -- JSON array
    skills_required TEXT,       -- JSON array
    result TEXT,
    error TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    started_at DATETIME,
    completed_at DATETIME,
    updated_at DATETIME NOT NULL,
    metadata TEXT               -- JSON object
);
```

## License

MIT License
