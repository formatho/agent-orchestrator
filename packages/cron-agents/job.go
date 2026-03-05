package cron

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled cron job.
type Job struct {
	// ID is the unique identifier for the job.
	ID string `json:"id"`

	// Name is a human-readable name for the job.
	Name string `json:"name"`

	// Schedule is the cron expression (e.g., "0 9 * * *" or "@daily").
	Schedule string `json:"schedule"`

	// Timezone specifies the timezone for the job schedule.
	// If empty, uses the scheduler's default timezone.
	Timezone string `json:"timezone,omitempty"`

	// AgentID identifies which agent should be executed.
	AgentID string `json:"agent_id"`

	// TODO is the task template to create when the job runs.
	// This is a generic interface to allow flexibility in task definitions.
	// It will be JSON-encoded when stored.
	// TODO can be any serializable structure representing the task.
	// Example: map[string]interface{}{"description": "Generate report"}
	TODO interface{} `json:"todo,omitempty"`

	// Enabled determines if the job is active.
	Enabled bool `json:"enabled"`

	// LastRun is the timestamp of the last successful execution.
	LastRun time.Time `json:"last_run,omitempty"`

	// NextRun is the timestamp of the next scheduled execution.
	NextRun time.Time `json:"next_run,omitempty"`

	// CreatedAt is the timestamp when the job was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the job was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata stores additional job-specific configuration.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// parsedSchedule holds the parsed cron expression.
	parsedSchedule *ParsedSchedule `json:"-"`

	// location caches the parsed timezone location.
	location *time.Location `json:"-"`

	// cronEntryID is the internal cron entry ID.
	cronEntryID cron.EntryID `json:"-"`
}

// Validate validates the job configuration.
func (j *Job) Validate() error {
	if j.ID == "" {
		return errors.New("job ID cannot be empty")
	}

	if j.Schedule == "" {
		return errors.New("job schedule cannot be empty")
	}

	if j.AgentID == "" {
		return errors.New("job AgentID cannot be empty")
	}

	// Parse and validate the schedule
	parser := NewParser()
	parsed, err := parser.Parse(j.Schedule)
	if err != nil {
		return err
	}
	j.parsedSchedule = parsed

	return nil
}

// GetLocation returns the job's timezone as a *time.Location.
// Falls back to UTC if not specified or invalid.
func (j *Job) GetLocation() *time.Location {
	if j.location != nil {
		return j.location
	}

	if j.Timezone == "" {
		j.location = time.UTC
		return j.location
	}

	loc, err := time.LoadLocation(j.Timezone)
	if err != nil {
		j.location = time.UTC
	} else {
		j.location = loc
	}

	return j.location
}

// CalculateNextRun calculates the next run time based on the schedule.
func (j *Job) CalculateNextRun(from time.Time) (time.Time, error) {
	if j.parsedSchedule == nil {
		parser := NewParser()
		parsed, err := parser.Parse(j.Schedule)
		if err != nil {
			return time.Time{}, err
		}
		j.parsedSchedule = parsed
	}

	loc := j.GetLocation()
	fromInLoc := from.In(loc)
	return j.parsedSchedule.Next(fromInLoc), nil
}

// MarshalTODO serializes the TODO field to JSON bytes.
func (j *Job) MarshalTODO() ([]byte, error) {
	if j.TODO == nil {
		return nil, nil
	}
	return json.Marshal(j.TODO)
}

// UnmarshalTODO deserializes JSON bytes into the TODO field.
func (j *Job) UnmarshalTODO(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &j.TODO)
}

// JobStatus represents the current status of a job.
type JobStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	LastRun   time.Time `json:"last_run"`
	NextRun   time.Time `json:"next_run"`
	IsRunning bool      `json:"is_running"`
}

// JobList is a collection of jobs with helper methods.
type JobList []Job

// FindByID finds a job by its ID.
func (jl JobList) FindByID(id string) *Job {
	for i := range jl {
		if jl[i].ID == id {
			return &jl[i]
		}
	}
	return nil
}

// FilterEnabled returns only enabled jobs.
func (jl JobList) FilterEnabled() JobList {
	var result JobList
	for _, job := range jl {
		if job.Enabled {
			result = append(result, job)
		}
	}
	return result
}

// FilterByAgent returns jobs for a specific agent.
func (jl JobList) FilterByAgent(agentID string) JobList {
	var result JobList
	for _, job := range jl {
		if job.AgentID == agentID {
			result = append(result, job)
		}
	}
	return result
}
