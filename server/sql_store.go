package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
)

type SQLStore struct {
	db         *sql.DB
	api        plugin.API
	driverName string
}

func NewSQLStore(api plugin.API) (*SQLStore, error) {
	config := api.GetUnsanitizedConfig()
	if config == nil {
		return nil, fmt.Errorf("failed to get config")
	}

	driverName := model.DatabaseDriverPostgres
	if config.SqlSettings.DriverName != nil {
		driverName = *config.SqlSettings.DriverName
	}

	if config.SqlSettings.DataSource == nil {
		return nil, fmt.Errorf("sql data source is nil")
	}

	db, err := sql.Open(driverName, *config.SqlSettings.DataSource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	s := &SQLStore{
		db:         db,
		api:        api,
		driverName: driverName,
	}

	if err := s.RunMigrations(); err != nil {
		return nil, errors.Wrap(err, "failed to run migrations")
	}

	return s, nil
}

func (s *SQLStore) RunMigrations() error {
	// Simple migration runner
	migrations := []struct {
		Name string
		SQL  string
	}{
		{
			Name: "000001_create_todos",
			SQL: `
				CREATE TABLE IF NOT EXISTS todos (
					id VARCHAR(26) PRIMARY KEY,
					message TEXT,
					description TEXT,
					creator_id VARCHAR(26),
					assignee_id VARCHAR(26),
					post_id VARCHAR(26),
					priority INTEGER DEFAULT 0,
					due_at BIGINT DEFAULT 0,
					status VARCHAR(20) DEFAULT 'open',
					created_at BIGINT,
					updated_at BIGINT,
					post_permalink TEXT,
					foreign_issue_id VARCHAR(26),
					foreign_user_id VARCHAR(26)
				);
			`,
		},
		{
			Name: "000002_create_comments",
			SQL: `
				CREATE TABLE IF NOT EXISTS todo_comments (
					id VARCHAR(26) PRIMARY KEY,
					todo_id VARCHAR(26),
					user_id VARCHAR(26),
					message TEXT,
					created_at BIGINT
				);
			`,
		},
		{
			Name: "000003_create_audit_log",
			SQL: `
				CREATE TABLE IF NOT EXISTS todo_audit_log (
					id VARCHAR(26) PRIMARY KEY,
					todo_id VARCHAR(26),
					user_id VARCHAR(26),
					action VARCHAR(50),
					metadata TEXT,
					created_at BIGINT
				);
			`,
		},
		{
			Name: "000004_create_preferences",
			SQL: `
				CREATE TABLE IF NOT EXISTS todo_preferences (
					user_id VARCHAR(26) PRIMARY KEY,
					reminder_enabled BOOLEAN DEFAULT TRUE,
					last_reminder_at BIGINT DEFAULT 0,
					allow_incoming_task BOOLEAN DEFAULT TRUE
				);
			`,
		},
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m.SQL); err != nil {
			return errors.Wrapf(err, "failed to run migration %s", m.Name)
		}
	}

	// Add indexes (separate to handle IF NOT EXISTS variations or failure gracefully)
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todos_creator ON todos (creator_id);")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todos_assignee ON todos (assignee_id, status);")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todo_comments_todo_id ON todo_comments (todo_id, created_at);")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todo_audit_log_todo_id ON todo_audit_log (todo_id, created_at);")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todos_due_at ON todos (due_at) WHERE due_at > 0;")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_todos_status ON todos (status);")

	return nil
}

// ListStore Implementation

func (s *SQLStore) SaveIssue(issue *Issue) error {
	var query string
	if s.driverName == "postgres" {
		query = `
			INSERT INTO todos (id, message, description, post_permalink, created_at, updated_at, post_id, creator_id, assignee_id, priority, due_at, status, foreign_issue_id, foreign_user_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				message = EXCLUDED.message,
				description = EXCLUDED.description,
				post_permalink = EXCLUDED.post_permalink,
				updated_at = EXCLUDED.updated_at,
				assignee_id = EXCLUDED.assignee_id,
				priority = EXCLUDED.priority,
				due_at = EXCLUDED.due_at,
				status = EXCLUDED.status,
				foreign_issue_id = EXCLUDED.foreign_issue_id,
				foreign_user_id = EXCLUDED.foreign_user_id;
		`
	} else { // Assuming MySQL for other drivers
		query = `
			INSERT INTO todos (id, message, description, post_permalink, created_at, updated_at, post_id, creator_id, assignee_id, priority, due_at, status, foreign_issue_id, foreign_user_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				message = VALUES(message),
				description = VALUES(description),
				post_permalink = VALUES(post_permalink),
				updated_at = VALUES(updated_at),
				assignee_id = VALUES(assignee_id),
				priority = VALUES(priority),
				due_at = VALUES(due_at),
				status = VALUES(status),
				foreign_issue_id = VALUES(foreign_issue_id),
				foreign_user_id = VALUES(foreign_user_id);
		`
	}

	_, err := s.db.Exec(s.replacePlaceholders(query), 
		issue.ID, issue.Message, issue.Description, issue.PostPermalink, issue.CreateAt, issue.UpdateAt, issue.PostID, issue.CreatorID, issue.AssigneeID, issue.Priority, issue.DueAt, issue.Status, issue.ForeignIssueID, issue.ForeignUserID)
	return err
}

func (s *SQLStore) replacePlaceholders(query string) string {
	if s.driverName != model.DatabaseDriverPostgres {
		return query
	}

	n := 1
	for strings.Contains(query, "?") {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", n), 1)
		n++
	}
	return query
}

func (s *SQLStore) GetIssue(issueID string) (*Issue, error) {
	issue := &Issue{}
	err := s.db.QueryRow(s.replacePlaceholders("SELECT id, message, description, post_permalink, created_at, updated_at, post_id, creator_id, assignee_id, priority, due_at, status, foreign_issue_id, foreign_user_id FROM todos WHERE id = ?"), issueID).
		Scan(&issue.ID, &issue.Message, &issue.Description, &issue.PostPermalink, &issue.CreateAt, &issue.UpdateAt, &issue.PostID, &issue.CreatorID, &issue.AssigneeID, &issue.Priority, &issue.DueAt, &issue.Status, &issue.ForeignIssueID, &issue.ForeignUserID)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *SQLStore) RemoveIssue(issueID string) error {
	_, err := s.db.Exec(s.replacePlaceholders("DELETE FROM todos WHERE id = ?"), issueID)
	return err
}

func (s *SQLStore) GetAndRemoveIssue(issueID string) (*Issue, error) {
	issue, err := s.GetIssue(issueID)
	if err != nil {
		return nil, err
	}
	err = s.RemoveIssue(issueID)
	return issue, err
}

func (s *SQLStore) AddReference(userID, issueID, listID, foreignUserID, foreignIssueID string) error {
    // In SQL world, "Adding a reference" means updating the creator/assignee/status.
    // listID "" (My), "_in" (In), "_out" (Out)
    
    status := "open"
    if listID == "_in" {
        status = "pending"
    }
    
    _, err := s.db.Exec(s.replacePlaceholders("UPDATE todos SET assignee_id = ?, status = ?, foreign_user_id = ?, foreign_issue_id = ? WHERE id = ?"), 
        userID, status, foreignUserID, foreignIssueID, issueID)
    return err
}

func (s *SQLStore) RemoveReference(userID, issueID, listID string) error {
	// For SQL, removing reference means archiving it so it doesn't show in active lists.
	_, err := s.db.Exec(s.replacePlaceholders("UPDATE todos SET status = 'archived', updated_at = ? WHERE id = ? AND (assignee_id = ? OR creator_id = ?) AND status IN ('open', 'pending')"),
		model.GetMillis(), issueID, userID, userID)
	return err
}

func (s *SQLStore) PopReference(userID, listID string) (*IssueRef, error) {
    // Not very SQL-friendly but needed for interface.
    // Get the first one and return it as a reference.
    var issueID string
    err := s.db.QueryRow(s.replacePlaceholders("SELECT id FROM todos WHERE assignee_id = ? AND status = 'open' ORDER BY created_at ASC LIMIT 1"), userID).Scan(&issueID)
    if err != nil {
        return nil, err
    }
    return &IssueRef{IssueID: issueID}, nil
}

func (s *SQLStore) BumpReference(userID, issueID, listID string) error {
    // Update updated_at to bring it to top
    _, err := s.db.Exec(s.replacePlaceholders("UPDATE todos SET updated_at = ? WHERE id = ?"), model.GetMillis(), issueID)
    return err
}

func (s *SQLStore) GetIssueReference(userID, issueID, listID string) (*IssueRef, int, error) {
    // Return reference if it belongs to user
    issue, err := s.GetIssue(issueID)
    if err != nil {
        return nil, 0, err
    }
    if issue.AssigneeID != userID {
        return nil, 0, errors.New("not assigned to this user")
    }
    return &IssueRef{IssueID: issueID}, 0, nil
}

func (s *SQLStore) GetIssueListAndReference(userID, issueID string) (string, *IssueRef, int) {
    issue, err := s.GetIssue(issueID)
    if err != nil {
        return "", nil, 0
    }
    if issue.AssigneeID == userID {
        if issue.Status == "pending" {
            return "_in", &IssueRef{IssueID: issueID}, 0
        }
        return "", &IssueRef{IssueID: issueID}, 0
    }
    if issue.CreatorID == userID && issue.AssigneeID != userID {
        return "_out", &IssueRef{IssueID: issueID}, 0
    }
    return "", nil, 0
}

func (s *SQLStore) GetList(userID, listID string) ([]*IssueRef, error) {
    var query string
    if listID == "" { // My
        query = "SELECT id FROM todos WHERE assignee_id = ? AND status = 'open' ORDER BY updated_at DESC"
    } else if listID == "_in" {
        query = "SELECT id FROM todos WHERE assignee_id = ? AND status = 'pending' ORDER BY updated_at DESC"
    } else if listID == "_out" {
        query = "SELECT id FROM todos WHERE creator_id = ? AND assignee_id != ? AND status = 'pending' ORDER BY updated_at DESC"
    }
    
    	var rows *sql.Rows
	var err error
	if listID == "_out" {
		rows, err = s.db.Query(s.replacePlaceholders(query), userID, userID)
	} else {
		rows, err = s.db.Query(s.replacePlaceholders(query), userID)
	}

	if err != nil {
		return nil, err
	}
    
    defer rows.Close()
    
    var refs []*IssueRef
    for rows.Next() {
        var id string
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        refs = append(refs, &IssueRef{IssueID: id})
    }
    return refs, nil
}

// Preferences implementation

func (s *SQLStore) SetReminderPreference(userID string, enabled bool) error {
	query := "INSERT INTO todo_preferences (user_id, reminder_enabled) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET reminder_enabled = EXCLUDED.reminder_enabled"
	if s.driverName == model.DatabaseDriverMysql {
		query = "INSERT INTO todo_preferences (user_id, reminder_enabled) VALUES (?, ?) ON DUPLICATE KEY UPDATE reminder_enabled = VALUES(reminder_enabled)"
	}
	_, err := s.db.Exec(s.replacePlaceholders(query), userID, enabled)
	return err
}

func (s *SQLStore) GetReminderPreference(userID string) bool {
	var enabled bool
	err := s.db.QueryRow(s.replacePlaceholders("SELECT reminder_enabled FROM todo_preferences WHERE user_id = ?"), userID).Scan(&enabled)
	if err != nil {
		return true // Default
	}
	return enabled
}

func (s *SQLStore) SetLastReminderTime(userID string, time int64) error {
	query := "INSERT INTO todo_preferences (user_id, last_reminder_at) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET last_reminder_at = EXCLUDED.last_reminder_at"
	if s.driverName == model.DatabaseDriverMysql {
		query = "INSERT INTO todo_preferences (user_id, last_reminder_at) VALUES (?, ?) ON DUPLICATE KEY UPDATE last_reminder_at = VALUES(last_reminder_at)"
	}
	_, err := s.db.Exec(s.replacePlaceholders(query), userID, time)
	return err
}

func (s *SQLStore) GetLastReminderTime(userID string) (int64, error) {
	var last int64
	err := s.db.QueryRow(s.replacePlaceholders("SELECT last_reminder_at FROM todo_preferences WHERE user_id = ?"), userID).Scan(&last)
	if err != nil {
		return 0, nil
	}
	return last, nil
}

func (s *SQLStore) SetAllowIncomingTaskPreference(userID string, enabled bool) error {
	query := "INSERT INTO todo_preferences (user_id, allow_incoming_task) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET allow_incoming_task = EXCLUDED.allow_incoming_task"
	if s.driverName == model.DatabaseDriverMysql {
		query = "INSERT INTO todo_preferences (user_id, allow_incoming_task) VALUES (?, ?) ON DUPLICATE KEY UPDATE allow_incoming_task = VALUES(allow_incoming_task)"
	}
	_, err := s.db.Exec(s.replacePlaceholders(query), userID, enabled)
	return err
}

func (s *SQLStore) GetAllowIncomingTaskPreference(userID string) (bool, error) {
	var enabled bool
	err := s.db.QueryRow(s.replacePlaceholders("SELECT allow_incoming_task FROM todo_preferences WHERE user_id = ?"), userID).Scan(&enabled)
	if err != nil {
		return true, nil
	}
	return enabled, nil
}

func (s *SQLStore) SaveComment(comment *Comment) error {
	if comment.ID == "" {
		comment.ID = model.NewId()
	}
	if comment.CreatedAt == 0 {
		comment.CreatedAt = model.GetMillis()
	}
	_, err := s.db.Exec(s.replacePlaceholders("INSERT INTO todo_comments (id, todo_id, user_id, message, created_at) VALUES (?, ?, ?, ?, ?)"),
		comment.ID, comment.TodoID, comment.UserID, comment.Message, comment.CreatedAt)
	return err
}

func (s *SQLStore) GetComments(todoID string) ([]*Comment, error) {
	rows, err := s.db.Query(s.replacePlaceholders("SELECT id, todo_id, user_id, message, created_at FROM todo_comments WHERE todo_id = ? ORDER BY created_at ASC"), todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.TodoID, &c.UserID, &c.Message, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *SQLStore) GetComment(commentID string) (*Comment, error) {
	c := &Comment{}
	err := s.db.QueryRow(s.replacePlaceholders("SELECT id, todo_id, user_id, message, created_at FROM todo_comments WHERE id = ?"), commentID).
		Scan(&c.ID, &c.TodoID, &c.UserID, &c.Message, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *SQLStore) DeleteComment(commentID string) error {
	_, err := s.db.Exec(s.replacePlaceholders("DELETE FROM todo_comments WHERE id = ?"), commentID)
	return err
}

func (s *SQLStore) AddAuditLog(log *AuditLog) error {
	if log.ID == "" {
		log.ID = model.NewId()
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = model.GetMillis()
	}
	_, err := s.db.Exec(s.replacePlaceholders("INSERT INTO todo_audit_log (id, todo_id, user_id, action, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)"),
		log.ID, log.TodoID, log.UserID, log.Action, log.Metadata, log.CreatedAt)
	return err
}

func (s *SQLStore) GetAuditLogs(todoID string) ([]*AuditLog, error) {
	rows, err := s.db.Query(s.replacePlaceholders("SELECT id, todo_id, user_id, action, metadata, created_at FROM todo_audit_log WHERE todo_id = ? ORDER BY created_at DESC"), todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		l := &AuditLog{}
		if err := rows.Scan(&l.ID, &l.TodoID, &l.UserID, &l.Action, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
