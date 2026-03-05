package todoqueue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store provides SQLite-based persistence for TODO items.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStore creates a new Store with the given database path.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for SQLite
	db.SetMaxOpenConns(1) // SQLite works best with a single connection
	db.SetMaxIdleConns(1)

	store := &Store{db: db}
	return store, nil
}

// RunMigrations creates the necessary database tables.
func (s *Store) RunMigrations() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS todo_items (
		id TEXT PRIMARY KEY,
		priority INTEGER NOT NULL DEFAULT 0,
		description TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		dependencies TEXT,
		skills_required TEXT,
		result TEXT,
		error TEXT,
		retry_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		started_at DATETIME,
		completed_at DATETIME,
		updated_at DATETIME NOT NULL,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_status ON todo_items(status);
	CREATE INDEX IF NOT EXISTS idx_priority ON todo_items(priority DESC);
	CREATE INDEX IF NOT EXISTS idx_status_priority ON todo_items(status, priority DESC);
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Save inserts or updates an item in the database.
func (s *Store) Save(item *Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deps, err := item.MarshalDependencies()
	if err != nil {
		return fmt.Errorf("failed to marshal dependencies: %w", err)
	}

	skills, err := item.MarshalSkillsRequired()
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	metadata, err := item.MarshalMetadata()
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
	INSERT OR REPLACE INTO todo_items (
		id, priority, description, status, dependencies, skills_required,
		result, error, retry_count, created_at, started_at, completed_at,
		updated_at, metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		item.ID,
		item.Priority,
		item.Description,
		item.Status,
		deps,
		skills,
		item.Result,
		item.Error,
		item.RetryCount,
		item.CreatedAt,
		item.StartedAt,
		item.CompletedAt,
		item.UpdatedAt,
		metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}

	return nil
}

// Get retrieves an item by ID.
func (s *Store) Get(id string) (*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, priority, description, status, dependencies, skills_required,
		   result, error, retry_count, created_at, started_at, completed_at,
		   updated_at, metadata
	FROM todo_items WHERE id = ?
	`

	item, err := s.scanItem(s.db.QueryRow(query, id))
	if err != nil {
		return nil, err
	}

	return item, nil
}

// Update updates specific fields of an item.
func (s *Store) Update(id string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(updates) == 0 {
		return nil
	}

	// Always update updated_at
	updates["updated_at"] = time.Now()

	// Build the UPDATE query dynamically
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE todo_items SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// Delete removes an item from the database.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM todo_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// Query retrieves items based on the provided filter.
func (s *Store) Query(filter Filter) ([]*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT id, priority, description, status, dependencies, skills_required, result, error, retry_count, created_at, started_at, completed_at, updated_at, metadata FROM todo_items"
	where := make([]string, 0)
	args := make([]interface{}, 0)

	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.MinPriority > 0 {
		where = append(where, "priority >= ?")
		args = append(args, filter.MinPriority)
	}

	if filter.MaxPriority > 0 {
		where = append(where, "priority <= ?")
		args = append(args, filter.MaxPriority)
	}

	if filter.HasDependencies != nil {
		if *filter.HasDependencies {
			where = append(where, "dependencies IS NOT NULL AND dependencies != 'null' AND dependencies != '[]'")
		} else {
			where = append(where, "(dependencies IS NULL OR dependencies = 'null' OR dependencies = '[]')")
		}
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// Order by priority (descending) then by created_at (ascending)
	query += " ORDER BY priority DESC, created_at ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	items := make([]*Item, 0)
	for rows.Next() {
		item, err := s.scanItemFromRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Filter by skills if specified (in-memory filtering for JSON field)
	if len(filter.Skills) > 0 {
		items = s.filterBySkills(items, filter.Skills)
	}

	return items, nil
}

// GetNextPending returns the highest priority pending item.
func (s *Store) GetNextPending() (*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, priority, description, status, dependencies, skills_required,
		   result, error, retry_count, created_at, started_at, completed_at,
		   updated_at, metadata
	FROM todo_items
	WHERE status = ?
	ORDER BY priority DESC, created_at ASC
	LIMIT 1
	`

	item, err := s.scanItem(s.db.QueryRow(query, StatusPending))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return item, nil
}

// GetDependencies retrieves all dependencies for an item.
func (s *Store) GetDependencies(itemID string) ([]*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First get the item to find its dependencies
	var depsJSON []byte
	err := s.db.QueryRow("SELECT dependencies FROM todo_items WHERE id = ?", itemID).Scan(&depsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get item dependencies: %w", err)
	}

	var depIDs []string
	if len(depsJSON) > 0 {
		if err := decodeJSON(depsJSON, &depIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
		}
	}

	if len(depIDs) == 0 {
		return nil, nil
	}

	// Fetch all dependency items
	placeholders := make([]string, len(depIDs))
	args := make([]interface{}, len(depIDs))
	for i, id := range depIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
	SELECT id, priority, description, status, dependencies, skills_required,
		   result, error, retry_count, created_at, started_at, completed_at,
		   updated_at, metadata
	FROM todo_items WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	items := make([]*Item, 0)
	for rows.Next() {
		item, err := s.scanItemFromRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// Count returns the count of items matching the filter.
func (s *Store) Count(filter Filter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT COUNT(*) FROM todo_items"
	where := make([]string, 0)
	args := make([]interface{}, 0)

	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.MinPriority > 0 {
		where = append(where, "priority >= ?")
		args = append(args, filter.MinPriority)
	}

	if filter.MaxPriority > 0 {
		where = append(where, "priority <= ?")
		args = append(args, filter.MaxPriority)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}

	return count, nil
}

// scanItem scans an item from a single row query.
func (s *Store) scanItem(row *sql.Row) (*Item, error) {
	var item Item
	var deps, skills, metadata []byte
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.Priority,
		&item.Description,
		&item.Status,
		&deps,
		&skills,
		&item.Result,
		&item.Error,
		&item.RetryCount,
		&item.CreatedAt,
		&startedAt,
		&completedAt,
		&item.UpdatedAt,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}

	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}

	if err := item.UnmarshalDependencies(deps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
	}
	if err := item.UnmarshalSkillsRequired(skills); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}
	if err := item.UnmarshalMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &item, nil
}

// scanItemFromRow scans an item from a rows iterator.
func (s *Store) scanItemFromRow(rows *sql.Rows) (*Item, error) {
	var item Item
	var deps, skills, metadata []byte
	var startedAt, completedAt sql.NullTime

	err := rows.Scan(
		&item.ID,
		&item.Priority,
		&item.Description,
		&item.Status,
		&deps,
		&skills,
		&item.Result,
		&item.Error,
		&item.RetryCount,
		&item.CreatedAt,
		&startedAt,
		&completedAt,
		&item.UpdatedAt,
		&metadata,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}

	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}

	if err := item.UnmarshalDependencies(deps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
	}
	if err := item.UnmarshalSkillsRequired(skills); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}
	if err := item.UnmarshalMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &item, nil
}

// filterBySkills filters items by required skills in-memory.
func (s *Store) filterBySkills(items []*Item, requiredSkills []string) []*Item {
	result := make([]*Item, 0)
	for _, item := range items {
		if s.hasAllSkills(item.SkillsRequired, requiredSkills) {
			result = append(result, item)
		}
	}
	return result
}

// hasAllSkills checks if all required skills are present in item skills.
func (s *Store) hasAllSkills(itemSkills, requiredSkills []string) bool {
	skillSet := make(map[string]bool)
	for _, skill := range itemSkills {
		skillSet[skill] = true
	}

	for _, required := range requiredSkills {
		if !skillSet[required] {
			return false
		}
	}

	return true
}

// decodeJSON is a helper to decode JSON bytes.
func decodeJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}
