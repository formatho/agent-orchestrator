# go-cron-agents

A powerful cron-based scheduler for AI agents with timezone support, human-readable aliases, and persistent run history tracking via SQLite.

## Features

- **Standard Cron Syntax** - Full support for 5-field cron expressions (`0 9 * * *`)
- **Human-Readable Aliases** - `@daily`, `@hourly`, `@weekly`, `@monthly`, `@yearly`
- **Interval Syntax** - `@every 5m`, `@every 1h30m`
- **Timezone Support** - Schedule jobs in any timezone (IANA format: `America/New_York`)
- **Missed Run Handling** - Configurable behavior for missed runs on startup
- **Retry on Failure** - Automatic retry with configurable max attempts
- **Run History Tracking** - SQLite-based persistent history
- **Thread-Safe Operations** - Safe for concurrent use
- **Job Management** - Add, remove, pause, resume, and update jobs
- **Manual Triggering** - Run jobs on-demand with `RunJob()`
- **Statistics API** - Get scheduler and history statistics

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/cron-agents
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    cron "github.com/formatho/agent-orchestrator/packages/cron-agents"
)

func main() {
    // Create scheduler with configuration
    scheduler, err := cron.NewScheduler(cron.Config{
        DBPath:            "./cron.db",
        DefaultTimezone:   "America/New_York",
        RetryOnFailure:    true,
        MaxRetries:        2,
        MissedRunBehavior: "skip", // or "run" to execute missed jobs
    })
    if err != nil {
        panic(err)
    }
    defer scheduler.Close()

    // Set executor (the function that runs jobs)
    scheduler.SetExecutor("default", func(ctx context.Context, job *cron.Job) (interface{}, error) {
        fmt.Printf("Executing job: %s\n", job.ID)
        // Your job logic here - call AI agent, etc.
        return map[string]string{"status": "completed"}, nil
    })

    // Add a job that runs daily at 9 AM
    err = scheduler.AddJob(cron.Job{
        ID:       "daily-report",
        Name:     "Daily Report Generator",
        Schedule: "0 9 * * *", // 9 AM daily
        Timezone: "America/New_York",
        AgentID:  "agent-1",
        TODO: map[string]interface{}{
            "description": "Generate daily report",
            "template":    "standard",
        },
        Enabled: true,
    })
    if err != nil {
        panic(err)
    }

    // Add a job using alias
    scheduler.AddJob(cron.Job{
        ID:       "hourly-sync",
        Schedule: "@hourly",
        AgentID:  "agent-2",
        TODO: map[string]interface{}{
            "task": "sync-data",
        },
    })

    // Start the scheduler
    if err := scheduler.Start(); err != nil {
        panic(err)
    }

    // Keep running...
    select {}
}
```

## Schedule Syntax

### Standard Cron (5 fields)

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday = 0)
│ │ │ │ │
* * * * *
```

Examples:
- `0 9 * * *` - Every day at 9:00 AM
- `30 14 * * 1-5` - Weekdays at 2:30 PM
- `0 0 1 * *` - First day of every month at midnight
- `*/15 * * * *` - Every 15 minutes

### Human-Readable Aliases

| Alias | Equivalent | Description |
|-------|------------|-------------|
| `@yearly` | `0 0 1 1 *` | Once a year (Jan 1 at midnight) |
| `@annually` | `0 0 1 1 *` | Same as @yearly |
| `@monthly` | `0 0 1 * *` | First of every month at midnight |
| `@weekly` | `0 0 * * 0` | Every Sunday at midnight |
| `@daily` | `0 0 * * *` | Every day at midnight |
| `@midnight` | `0 0 * * *` | Same as @daily |
| `@hourly` | `0 * * * *` | Every hour at minute 0 |

### Interval Syntax

```
@every <duration>
```

Examples:
- `@every 5m` - Every 5 minutes
- `@every 1h30m` - Every 1 hour 30 minutes
- `@every 24h` - Every 24 hours

## API Reference

### Configuration

```go
type Config struct {
    DBPath            string // SQLite database path (empty = in-memory)
    DefaultTimezone   string // Default timezone (default: UTC)
    RetryOnFailure    bool   // Enable retry on failure
    MaxRetries        int    // Max retry attempts (default: 3)
    MissedRunBehavior string // "run", "skip", or "ignore" missed jobs
    Logger            Logger // Optional logger
}
```

### Job Structure

```go
type Job struct {
    ID        string                 // Unique identifier
    Name      string                 // Human-readable name
    Schedule  string                 // Cron expression
    Timezone  string                 // IANA timezone (e.g., "America/New_York")
    AgentID   string                 // Target agent identifier
    TODO      interface{}            // Task definition (any serializable data)
    Enabled   bool                   // Job enabled status
    LastRun   time.Time              // Last execution time
    NextRun   time.Time              // Next scheduled time
    CreatedAt time.Time              // Creation timestamp
    UpdatedAt time.Time              // Last update timestamp
    Metadata  map[string]interface{} // Additional metadata
}
```

### Scheduler Methods

```go
// Create scheduler
scheduler, err := cron.NewScheduler(config, opts...)

// Job management
err = scheduler.AddJob(job)
err = scheduler.RemoveJob(id)
err = scheduler.PauseJob(id)
err = scheduler.ResumeJob(id)
err = scheduler.UpdateJob(job)
job, err := scheduler.GetJob(id)
jobs := scheduler.GetJobs()
jobs := scheduler.GetEnabledJobs()

// Scheduler control
err = scheduler.Start()
err = scheduler.Stop()
err = scheduler.Close()

// Execution
err = scheduler.RunJob(id) // Manual trigger
scheduler.SetExecutor("default", executor)
scheduler.SetExecutor(jobID, executor) // Job-specific executor

// History
history, err := scheduler.GetHistory(jobID, limit)
allHistory, err := scheduler.GetAllHistory(limit)
stats, err := scheduler.GetStats()
status, err := scheduler.GetJobStatus(jobID)

// Scheduling info
nextRun, err := scheduler.GetNextRun(jobID)
```

### Run History Structure

```go
type RunHistory struct {
    ID           int64                  // Unique record ID
    JobID        string                 // Job identifier
    ScheduledFor time.Time              // When it was scheduled
    StartedAt    time.Time              // When it started
    CompletedAt  time.Time              // When it finished
    Status       RunStatus              // success, failed, retry, cancelled
    Result       interface{}            // Job result data
    Error        string                 // Error message if failed
    RetryCount   int                    // Number of retries
    Duration     int64                  // Execution time in ms
    Metadata     map[string]interface{} // Additional data
}
```

### Schedule Builder

Fluent interface for creating cron expressions:

```go
builder := cron.NewScheduleBuilder()

// Daily at 9:30 AM
expr := builder.DailyAt(9, 30).Build()
// "30 9 * * *"

// Weekly on Monday at 10:00 AM
expr := builder.WeeklyOn(1, 10, 0).Build()
// "0 10 * * 1"

// Monthly on 15th at noon
expr := builder.MonthlyOn(15, 12, 0).Build()
// "0 12 15 * *"

// Every hour
expr := builder.EveryHour().Build()
// "0 * * * *"
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    cron "github.com/formatho/agent-orchestrator/packages/cron-agents"
)

func main() {
    // Create scheduler
    scheduler, err := cron.NewScheduler(cron.Config{
        DBPath:            "./data/cron.db",
        DefaultTimezone:   "UTC",
        RetryOnFailure:    true,
        MaxRetries:        2,
        MissedRunBehavior: "skip",
    })
    if err != nil {
        log.Fatalf("Failed to create scheduler: %v", err)
    }

    // Set default executor
    scheduler.SetExecutor("default", executeJob)

    // Define jobs
    jobs := []cron.Job{
        {
            ID:       "morning-briefing",
            Name:     "Morning Briefing",
            Schedule: "0 8 * * 1-5", // Weekdays at 8 AM
            Timezone: "America/New_York",
            AgentID:  "assistant",
            TODO: map[string]interface{}{
                "type":        "briefing",
                "recipients":  []string{"team@example.com"},
                "include_cal": true,
            },
        },
        {
            ID:       "daily-report",
            Name:     "Daily Report",
            Schedule: "@daily",
            AgentID:  "reporter",
            TODO: map[string]interface{}{
                "type":     "daily_summary",
                "channels": []string{"slack://general"},
            },
        },
        {
            ID:       "health-check",
            Name:     "Health Check",
            Schedule: "@every 5m",
            AgentID:  "monitor",
            TODO: map[string]interface{}{
                "type":     "health_check",
                "services": []string{"api", "db", "cache"},
            },
        },
    }

    // Add jobs
    for _, job := range jobs {
        if err := scheduler.AddJob(job); err != nil {
            log.Printf("Failed to add job %s: %v", job.ID, err)
        }
    }

    // Start scheduler
    if err := scheduler.Start(); err != nil {
        log.Fatalf("Failed to start scheduler: %v", err)
    }
    defer scheduler.Stop()

    fmt.Println("Scheduler started. Press Ctrl+C to stop.")

    // Print stats
    printStats(scheduler)

    // Wait for interrupt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    fmt.Println("\nShutting down...")
}

func executeJob(ctx context.Context, job *cron.Job) (interface{}, error) {
    start := time.Now()
    
    fmt.Printf("\n[%s] Executing job: %s (Agent: %s)\n", 
        start.Format(time.RFC3339), job.ID, job.AgentID)

    // Simulate work
    time.Sleep(100 * time.Millisecond)

    // Access TODO data
    if todo, ok := job.TODO.(map[string]interface{}); ok {
        if taskType, ok := todo["type"].(string); ok {
            fmt.Printf("  Task type: %s\n", taskType)
        }
    }

    duration := time.Since(start)
    fmt.Printf("  Completed in %v\n", duration)

    return map[string]interface{}{
        "duration_ms": duration.Milliseconds(),
        "status":      "success",
    }, nil
}

func printStats(s *cron.Scheduler) {
    stats, err := s.GetStats()
    if err != nil {
        log.Printf("Failed to get stats: %v", err)
        return
    }

    fmt.Println("\n=== Scheduler Stats ===")
    fmt.Printf("Total Jobs: %v\n", stats["total_jobs"])
    fmt.Printf("Enabled Jobs: %v\n", stats["enabled_jobs"])
    fmt.Printf("Running Jobs: %v\n", stats["running_jobs"])

    if history, ok := stats["history"].(map[string]interface{}); ok {
        fmt.Printf("Total Runs: %v\n", history["total_runs"])
    }
    fmt.Println("=======================")
}
```

## Best Practices

### Timezone Handling

```go
// Always specify timezone explicitly for time-sensitive jobs
Job{
    Schedule: "0 9 * * *", // 9 AM
    Timezone: "America/New_York", // Will run at 9 AM NY time
}

// For global services, use UTC
Job{
    Schedule: "0 0 * * *",
    Timezone: "UTC",
}
```

### Error Handling in Executors

```go
scheduler.SetExecutor("default", func(ctx context.Context, job *cron.Job) (interface{}, error) {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Do work with proper error handling
    result, err := doWork()
    if err != nil {
        return nil, fmt.Errorf("job %s failed: %w", job.ID, err)
    }

    return result, nil
})
```

### Cleanup

```go
// Always use defer for cleanup
scheduler, _ := cron.NewScheduler(config)
defer scheduler.Close()

// Or explicitly stop before close
scheduler.Stop()
scheduler.Close()
```

## License

MIT License
