package main

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

// IssueRef denotes every element in any of the lists. Contains the issue that refers to,
// and may contain foreign ids of issue and user, denoting the user this element is related to
// and the issue on that user system.
type IssueRef struct {
	IssueID        string `json:"issue_id"`
	ForeignIssueID string `json:"foreign_issue_id"`
	ForeignUserID  string `json:"foreign_user_id"`
}

// Issue represents a Todo issue
type Issue struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	PostPermalink string `json:"postPermalink"`
	Description   string `json:"description,omitempty"`
	CreateAt      int64  `json:"create_at"`
	UpdateAt      int64  `json:"update_at"`
	PostID        string `json:"post_id"`
	CreatorID     string `json:"creator_id"`
	AssigneeID    string `json:"assignee_id"`
	ForeignUserID string `json:"foreign_user_id"`
	ForeignIssueID string `json:"foreign_issue_id"`
	Priority      int    `json:"priority"`
	DueAt         int64  `json:"due_at"`
	Status        string `json:"status"`
}

// ExtendedIssue extends the information on Issue to be used on the front-end
type ExtendedIssue struct {
	Issue
	ForeignUser     string `json:"user"`
	ForeignList     string `json:"list"`
	ForeignPosition int    `json:"position"`
}

// ListsIssue for all list issues
type ListsIssue struct {
	In  []*ExtendedIssue `json:"in"`
	My  []*ExtendedIssue `json:"my"`
	Out []*ExtendedIssue `json:"out"`
}

// Comment represents a comment on a Todo
type Comment struct {
	ID        string `json:"id"`
	TodoID    string `json:"todo_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"created_at"`
}

// ExtendedComment adds user info to Comment
type ExtendedComment struct {
	Comment
	UserName string `json:"username"`
}

// AuditLog represents an action on a Todo
type AuditLog struct {
	ID        string `json:"id"`
	TodoID    string `json:"todo_id"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	Metadata  string `json:"metadata"`
	CreatedAt int64  `json:"created_at"`
}

func newIssue(message, postPermalink, description, postID, creatorID, assigneeID, status string, dueAt int64, priority int) *Issue {
	now := model.GetMillis()
	return &Issue{
		ID:            model.NewId(),
		CreateAt:      now,
		UpdateAt:      now,
		Message:       message,
		PostPermalink: postPermalink,
		Description:   description,
		PostID:        postID,
		CreatorID:     creatorID,
		AssigneeID:    assigneeID,
		Status:        status,
		DueAt:         dueAt,
		Priority:      priority,
	}
}

func issuesListToString(issues []*ExtendedIssue) string {
	if len(issues) == 0 {
		return "Nothing to do!"
	}

	str := "\n\n"

	for _, issue := range issues {
		createAt := time.Unix(issue.CreateAt/1000, 0)
		str += fmt.Sprintf("* %s\n  * (%s)\n", issue.Message, createAt.Format("January 2, 2006 at 15:04"))
	}

	return str
}
