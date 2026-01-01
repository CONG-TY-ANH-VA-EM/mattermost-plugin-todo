package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
)

const (
	// MyListKey is the key used to store the list of the owned todos
	MyListKey = ""
	// InListKey is the key used to store the list of received todos
	InListKey = "_in"
	// OutListKey is the key used to store the list of sent todos
	OutListKey = "_out"
)

// ListStore represents the KVStore operations for lists
type ListStore interface {
	// Issue related function
	SaveIssue(issue *Issue) error
	GetIssue(issueID string) (*Issue, error)
	RemoveIssue(issueID string) error
	GetAndRemoveIssue(issueID string) (*Issue, error)

	// Issue References related functions

	// AddReference creates a new IssueRef with the issueID, foreignUSerID and foreignIssueID, and stores it
	// on the listID for userID.
	AddReference(userID, issueID, listID, foreignUserID, foreignIssueID string) error
	// RemoveReference removes the IssueRef for issueID in listID for userID
	RemoveReference(userID, issueID, listID string) error
	// PopReference removes the first IssueRef in listID for userID and returns it
	PopReference(userID, listID string) (*IssueRef, error)
	// BumpReference moves the Issue reference for issueID in listID for userID to the beginning of the list
	BumpReference(userID, issueID, listID string) error
	// GetIssueReference gets the IssueRef and position of the issue issueID on user userID's list listID
	GetIssueReference(userID, issueID, listID string) (*IssueRef, int, error)
	// GetIssueListAndReference gets the issue list, IssueRef and position for user userID
	GetIssueListAndReference(userID, issueID string) (string, *IssueRef, int)
	// GetList returns the list of IssueRef in listID for userID
	GetList(userID, listID string) ([]*IssueRef, error)

	// Preferences
	SetReminderPreference(userID string, enabled bool) error
	GetReminderPreference(userID string) bool
	SetLastReminderTime(userID string, time int64) error
	GetLastReminderTime(userID string) (int64, error)
	SetAllowIncomingTaskPreference(userID string, enabled bool) error
	GetAllowIncomingTaskPreference(userID string) (bool, error)

	// Comments
	SaveComment(comment *Comment) error
	GetComments(todoID string) ([]*Comment, error)
	DeleteComment(commentID string) error
	GetComment(commentID string) (*Comment, error)

	// Audit Log
	AddAuditLog(log *AuditLog) error
	GetAuditLogs(todoID string) ([]*AuditLog, error)
}

type listManager struct {
	store ListStore
	api   plugin.API
}

// NewListManager creates a new listManager
func NewListManager(api plugin.API, store ListStore) ListManager {
	return &listManager{
		store: store,
		api:   api,
	}
}

func (l *listManager) AddIssue(userID, message, postPermalink, description, postID string, dueAt int64, priority int) (*Issue, error) {
	message = SanitizeInput(message)
	description = SanitizeMultiline(description)
	issue := newIssue(message, postPermalink, description, postID, userID, userID, "open", dueAt, priority)

	if err := l.store.SaveIssue(issue); err != nil {
		return nil, err
	}

	if err := l.store.AddReference(userID, issue.ID, MyListKey, "", ""); err != nil {
		if rollbackError := l.store.RemoveIssue(issue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback issue after add error, Err=", err.Error())
		}
		return nil, err
	}
	l.recordAuditLog(issue.ID, userID, "create", "")

	return issue, nil
}

func (l *listManager) SendIssue(senderID, receiverID, message, postPermalink, description, postID string, dueAt int64, priority int) (string, error) {
	message = SanitizeInput(message)
	description = SanitizeMultiline(description)
	senderIssue := newIssue(message, postPermalink, description, postID, senderID, receiverID, "pending", dueAt, priority)
	if err := l.store.SaveIssue(senderIssue); err != nil {
		return "", err
	}

	receiverIssue := newIssue(message, postPermalink, description, postID, senderID, receiverID, "pending", dueAt, priority)
	if err := l.store.SaveIssue(receiverIssue); err != nil {
		if rollbackError := l.store.RemoveIssue(senderIssue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback sender issue after send error, Err=", err.Error())
		}
		return "", err
	}

	if err := l.store.AddReference(senderID, senderIssue.ID, OutListKey, receiverID, receiverIssue.ID); err != nil {
		if rollbackError := l.store.RemoveIssue(senderIssue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback sender issue after send error, Err=", err.Error())
		}
		if rollbackError := l.store.RemoveIssue(receiverIssue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback receiver issue after send error, Err=", err.Error())
		}
		return "", err
	}

	if err := l.store.AddReference(receiverID, receiverIssue.ID, InListKey, senderID, senderIssue.ID); err != nil {
		if rollbackError := l.store.RemoveIssue(senderIssue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback sender issue after send error, Err=", err.Error())
		}
		if rollbackError := l.store.RemoveIssue(receiverIssue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback receiver issue after send error ,Err=", err.Error())
		}
		if rollbackError := l.store.RemoveReference(senderID, senderIssue.ID, OutListKey); rollbackError != nil {
			l.api.LogError("cannot rollback sender list after send error, Err=", err.Error())
		}
		return "", err
	}
	l.recordAuditLog(senderIssue.ID, senderID, "send", receiverID)
	l.recordAuditLog(receiverIssue.ID, receiverID, "receive", senderID)

	return receiverIssue.ID, nil
}

func (l *listManager) GetIssueList(userID, listID string) ([]*ExtendedIssue, error) {
	irs, err := l.store.GetList(userID, listID)
	if err != nil {
		return nil, err
	}

	extendedIssues := []*ExtendedIssue{}
	for _, ir := range irs {
		issue, err := l.store.GetIssue(ir.IssueID)
		if err != nil {
			continue
		}

		extendedIssue := l.extendIssueInfo(issue, ir)
		extendedIssues = append(extendedIssues, extendedIssue)
	}

	return extendedIssues, nil
}

func (l *listManager) GetAllList(userID string) (listsIssue *ListsIssue, err error) {
	inListIssue, err := l.GetIssueList(userID, InListKey)
	if err != nil {
		return nil, err
	}
	myListIssue, err := l.GetIssueList(userID, MyListKey)
	if err != nil {
		return nil, err
	}
	outListIssue, err := l.GetIssueList(userID, OutListKey)
	if err != nil {
		return nil, err
	}
	return &ListsIssue{
		In:  inListIssue,
		My:  myListIssue,
		Out: outListIssue,
	}, nil
}

func (l *listManager) CompleteIssue(userID, issueID string) (issue *Issue, foreignID string, listToUpdate string, err error) {
	issueList, ir, _ := l.store.GetIssueListAndReference(userID, issueID)
	if ir == nil {
		return nil, "", issueList, fmt.Errorf("cannot find element")
	}

	if err = l.store.RemoveReference(userID, issueID, issueList); err != nil {
		return nil, "", issueList, err
	}

	issue, err = l.store.GetIssue(issueID)
	if err != nil {
		return nil, "", issueList, err
	}
	issue.Status = "completed"
	issue.UpdateAt = model.GetMillis()
	if err := l.store.SaveIssue(issue); err != nil {
		l.api.LogError("cannot update issue status, Err=", err.Error())
	}
	l.recordAuditLog(issueID, userID, "complete", "")

	if ir.ForeignUserID == "" {
		return issue, "", issueList, nil
	}

	err = l.store.RemoveReference(ir.ForeignUserID, ir.ForeignIssueID, OutListKey)
	if err != nil {
		l.api.LogError("cannot clean foreigner list after complete, Err=", err.Error())
	}

	foreignIssue, err := l.store.GetIssue(ir.ForeignIssueID)
	if err == nil {
		foreignIssue.Status = "completed"
		foreignIssue.UpdateAt = model.GetMillis()
		_ = l.store.SaveIssue(foreignIssue)
	}

	return issue, ir.ForeignUserID, issueList, nil
}
func (l *listManager) EditIssue(userID string, issueID string, newMessage string, newDescription string, newDueAt int64, newPriority int) (string, string, string, error) {
	issue, err := l.store.GetIssue(issueID)
	if err != nil {
		return "", "", "", err
	}

	list, _, _ := l.store.GetIssueListAndReference(userID, issueID)

	oldMessage := issue.Message
	message := SanitizeInput(newMessage)
	description := SanitizeMultiline(newDescription)

	if issue.ForeignIssueID != "" {
		foreignIssue, foreignErr := l.store.GetIssue(issue.ForeignIssueID)
		if foreignErr == nil {
			foreignIssue.Message = message
			foreignIssue.Description = description
			foreignIssue.DueAt = newDueAt
			foreignIssue.Priority = newPriority
			foreignIssue.UpdateAt = model.GetMillis()
			_ = l.store.SaveIssue(foreignIssue)
		}
	}

	issue.Message = message
	issue.Description = description
	issue.DueAt = newDueAt
	issue.Priority = newPriority
	issue.UpdateAt = model.GetMillis()

	if err := l.store.SaveIssue(issue); err != nil {
		return "", "", "", err
	}
	l.recordAuditLog(issue.ID, userID, "edit", "")

	return issue.ForeignUserID, list, oldMessage, nil
}

func (l *listManager) ChangeAssignment(issueID string, userID string, sendTo string) (issue *Issue, oldOwner string, err error) {
	issue, err = l.store.GetIssue(issueID)
	if err != nil {
		return nil, "", err
	}

	list, ir, _ := l.store.GetIssueListAndReference(userID, issueID)
	if ir == nil {
		return nil, "", errors.New("reference not found")
	}

	if (list == InListKey) || (ir.ForeignIssueID != "" && list == MyListKey) {
		return nil, "", errors.New("trying to change the assignment of a todo not owned")
	}

	if ir.ForeignUserID != "" {
		// Remove reference from foreign user
		foreignList, foreignIR, _ := l.store.GetIssueListAndReference(ir.ForeignUserID, ir.ForeignIssueID)
		if foreignIR == nil {
			return nil, "", errors.New("reference not found")
		}

		if err := l.store.RemoveReference(ir.ForeignUserID, ir.ForeignIssueID, foreignList); err != nil {
			return nil, "", err
		}

		_, err := l.store.GetAndRemoveIssue(ir.ForeignIssueID)
		if err != nil {
			l.api.LogError("cannot remove issue", "err", err.Error())
		}
	}

	if userID == sendTo && list == OutListKey {
		if err := l.store.RemoveReference(userID, issueID, OutListKey); err != nil {
			return nil, "", err
		}

		if err := l.store.AddReference(userID, issueID, MyListKey, "", ""); err != nil {
			return nil, "", err
		}

		return issue, ir.ForeignUserID, nil
	}

	if userID != sendTo {
		if err := l.store.RemoveReference(userID, issueID, list); err != nil {
			return nil, "", err
		}
	}

	receiverIssue := newIssue(issue.Message, issue.PostPermalink, issue.Description, issue.PostID, userID, sendTo, "pending", issue.DueAt, issue.Priority)
	if err := l.store.SaveIssue(receiverIssue); err != nil {
		return nil, "", err
	}

	if err := l.store.AddReference(userID, issueID, OutListKey, sendTo, receiverIssue.ID); err != nil {
		return nil, "", err
	}

	l.recordAuditLog(receiverIssue.ID, sendTo, "receive", userID)
	l.recordAuditLog(issueID, userID, "reassign", sendTo)

	return issue, ir.ForeignUserID, nil
}

func (l *listManager) AcceptIssue(userID, issueID string) (todoMessage string, foreignUserID string, outErr error) {
	issue, err := l.store.GetIssue(issueID)
	if err != nil {
		return "", "", err
	}

	ir, _, err := l.store.GetIssueReference(userID, issueID, InListKey)
	if err != nil {
		return "", "", err
	}
	if ir == nil {
		return "", "", fmt.Errorf("element reference not found")
	}

	err = l.store.AddReference(userID, issueID, MyListKey, ir.ForeignUserID, ir.ForeignIssueID)
	if err != nil {
		return "", "", err
	}

	err = l.store.RemoveReference(userID, issueID, InListKey)
	if err != nil {
		if rollbackError := l.store.RemoveReference(userID, issueID, MyListKey); rollbackError != nil {
			l.api.LogError("cannot rollback accept operation, Err=", rollbackError.Error())
		}
		return "", "", err
	}

	l.recordAuditLog(issueID, userID, "accept", ir.ForeignUserID)

	if issue.PostPermalink != "" {
		issue.Message = fmt.Sprintf("%s\n[Permalink](%s)", issue.Message, issue.PostPermalink)
	}

	return issue.Message, ir.ForeignUserID, nil
}

func (l *listManager) RemoveIssue(userID, issueID string) (outIssue *Issue, foreignID string, isSender bool, listToUpdate string, outErr error) {
	issueList, ir, _ := l.store.GetIssueListAndReference(userID, issueID)
	if ir == nil {
		return nil, "", false, issueList, fmt.Errorf("cannot find element")
	}

	if err := l.store.RemoveReference(userID, issueID, issueList); err != nil {
		return nil, "", false, issueList, err
	}

	issue, err := l.store.GetIssue(issueID)
	if err != nil {
		return nil, "", false, issueList, err
	}
	issue.Status = "removed"
	issue.UpdateAt = model.GetMillis()
	if err := l.store.SaveIssue(issue); err != nil {
		l.api.LogError("cannot update issue status, Err=", err.Error())
	}

	if ir.ForeignUserID == "" {
		l.recordAuditLog(issueID, userID, "remove", "")
		return issue, "", false, issueList, nil
	}
	list, _, _ := l.store.GetIssueListAndReference(ir.ForeignUserID, ir.ForeignIssueID)

	err = l.store.RemoveReference(ir.ForeignUserID, ir.ForeignIssueID, list)
	if err != nil {
		l.api.LogError("cannot clean foreigner list after remove, Err=", err.Error())
	}

	issue, err = l.store.GetAndRemoveIssue(ir.ForeignIssueID)
	if err != nil {
		l.api.LogError("cannot clean foreigner issue after remove, Err=", err.Error())
	}

	l.recordAuditLog(issueID, userID, "remove", "")

	return issue, ir.ForeignUserID, list == OutListKey, issueList, nil
}

func (l *listManager) PopIssue(userID string) (issue *Issue, foreignID string, err error) {
	ir, err := l.store.PopReference(userID, MyListKey)
	if err != nil {
		return nil, "", err
	}

	if ir == nil {
		return nil, "", errors.New("unexpected nil for issue reference")
	}

	issue, err = l.store.GetIssue(ir.IssueID)
	if err != nil {
		l.api.LogError("cannot find issue after pop, Err=", err.Error())
		return nil, "", err
	}
	issue.Status = "completed"
	issue.UpdateAt = model.GetMillis()
	if err := l.store.SaveIssue(issue); err != nil {
		l.api.LogError("cannot update issue status after pop, Err=", err.Error())
	}
	if ir.ForeignUserID == "" {
		return issue, "", nil
	}

	err = l.store.RemoveReference(ir.ForeignUserID, ir.ForeignIssueID, OutListKey)
	if err != nil {
		l.api.LogError("cannot clean foreigner list after pop, Err=", err.Error())
	}
	foreignIssue, err := l.store.GetIssue(ir.ForeignIssueID)
	if err == nil {
		foreignIssue.Status = "completed"
		foreignIssue.UpdateAt = model.GetMillis()
		_ = l.store.SaveIssue(foreignIssue)
	}

	return issue, ir.ForeignUserID, nil
}

func (l *listManager) BumpIssue(userID, issueID string) (todo *Issue, receiver string, foreignIssueID string, outErr error) {
	ir, _, err := l.store.GetIssueReference(userID, issueID, OutListKey)
	if err != nil {
		return nil, "", "", err
	}

	if ir == nil {
		return nil, "", "", fmt.Errorf("cannot find sender issue")
	}

	err = l.store.BumpReference(ir.ForeignUserID, ir.ForeignIssueID, InListKey)
	if err != nil {
		return nil, "", "", err
	}

	issue, err := l.store.GetIssue(ir.ForeignIssueID)
	if err != nil {
		l.api.LogError("cannot find foreigner issue after bump, Err=", err.Error())
		return nil, "", "", nil
	}

	l.recordAuditLog(ir.ForeignIssueID, ir.ForeignUserID, "bumped_by", userID)

	return issue, ir.ForeignUserID, ir.ForeignIssueID, nil
}

func (l *listManager) GetUserName(userID string) string {
	user, err := l.api.GetUser(userID)
	if err != nil {
		return "Someone"
	}
	return user.Username
}

func (l *listManager) extendIssueInfo(issue *Issue, ir *IssueRef) *ExtendedIssue {
	if issue == nil || ir == nil {
		return nil
	}

	feIssue := &ExtendedIssue{
		Issue: *issue,
	}

	if ir.ForeignUserID == "" {
		return feIssue
	}

	list, _, n := l.store.GetIssueListAndReference(ir.ForeignUserID, ir.ForeignIssueID)

	var listName string
	switch list {
	case MyListKey:
		listName = MyListKey
	case InListKey:
		listName = InFlag
	case OutListKey:
		listName = OutFlag
	}

	userName := l.GetUserName(ir.ForeignUserID)

	feIssue.ForeignUser = userName
	feIssue.ForeignList = listName
	feIssue.ForeignPosition = n

	return feIssue
}

func (l *listManager) AddComment(todoID, userID, message string) (*Comment, error) {
	message = SanitizeMultiline(message)
	comment := &Comment{
		TodoID:  todoID,
		UserID:  userID,
		Message: message,
	}
	if err := l.store.SaveComment(comment); err != nil {
		return nil, err
	}
	l.recordAuditLog(todoID, userID, "add_comment", comment.ID)
	return comment, nil
}

func (l *listManager) GetIssueComments(todoID string) ([]*ExtendedComment, error) {
	comments, err := l.store.GetComments(todoID)
	if err != nil {
		return nil, err
	}

	extendedComments := make([]*ExtendedComment, 0, len(comments))
	for _, c := range comments {
		ec := &ExtendedComment{
			Comment:  *c,
			UserName: l.GetUserName(c.UserID),
		}
		extendedComments = append(extendedComments, ec)
	}

	return extendedComments, nil
}

func (l *listManager) DeleteComment(commentID, userID string) error {
	comment, err := l.store.GetComment(commentID)
	if err != nil {
		return err
	}

	if comment.UserID != userID {
		return errors.New("not authorized to delete this comment")
	}

	if err := l.store.DeleteComment(commentID); err != nil {
		return err
	}
	l.recordAuditLog(comment.TodoID, userID, "delete_comment", commentID)
	return nil
}

func (l *listManager) recordAuditLog(todoID, userID, action, metadata string) {
	log := &AuditLog{
		TodoID:   todoID,
		UserID:   userID,
		Action:   action,
		Metadata: metadata,
	}
	if err := l.store.AddAuditLog(log); err != nil {
		l.api.LogError("failed to record audit log", "error", err.Error())
	}
}

func (l *listManager) IsAuthorized(todoID, userID string) (bool, error) {
	// Root bypass
	user, err := l.api.GetUser(userID)
	if err == nil && user.IsSystemAdmin() {
		return true, nil
	}

	issue, err2 := l.store.GetIssue(todoID)
	if err2 != nil {
		return false, err2
	}

	if issue.CreatorID == userID || issue.AssigneeID == userID {
		return true, nil
	}
	return false, nil
}
