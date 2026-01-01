package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/telemetry"
	"github.com/pkg/errors"
)

const (
	// WSEventRefresh is the WebSocket event for refreshing the Todo list
	WSEventRefresh = "refresh"

	// WSEventConfigUpdate is the WebSocket event to update the Todo list's configurations on webapp
	WSEventConfigUpdate = "config_update"

	ErrorMsgAddIssue = "Unable to add issue"
)

// ListManager represents the logic on the lists
type ListManager interface {
	// AddIssue adds a todo to userID's myList with the message
	AddIssue(userID, message, postPermalink, description, postID string, dueAt int64, priority int) (*Issue, error)
	// SendIssue sends the todo with the message from senderID to receiverID and returns the receiver's issueID
	SendIssue(senderID, receiverID, message, postPermalink, description, postID string, dueAt int64, priority int) (string, error)
	// GetIssueList gets the todos on listID for userID
	GetIssueList(userID, listID string) ([]*ExtendedIssue, error)
	// GetAllList get all issues
	GetAllList(userID string) (*ListsIssue, error)
	// CompleteIssue completes the todo issueID for userID, and returns the issue and the foreign ID if any
	CompleteIssue(userID, issueID string) (issue *Issue, foreignID string, listToUpdate string, err error)
	// AcceptIssue moves one the todo issueID of userID from inbox to myList, and returns the message and the foreignUserID if any
	AcceptIssue(userID, issueID string) (todoMessage string, foreignUserID string, err error)
	// RemoveIssue removes the todo issueID for userID and returns the issue, the foreign ID if any and whether the user sent the todo to someone else
	RemoveIssue(userID, issueID string) (issue *Issue, foreignID string, isSender bool, listToUpdate string, err error)
	// PopIssue the first element of myList for userID and returns the issue and the foreign ID if any
	PopIssue(userID string) (issue *Issue, foreignID string, err error)
	// BumpIssue moves a issueID sent by userID to the top of its receiver inbox list
	BumpIssue(userID string, issueID string) (todo *Issue, receiver string, foreignIssueID string, err error)
	// Comments
	AddComment(todoID, userID, message string) (*Comment, error)
	GetIssueComments(todoID string) ([]*ExtendedComment, error)
	DeleteComment(commentID, userID string) error
	// EditIssue updates the message on an issue
	EditIssue(userID string, issueID string, newMessage string, newDescription string, newDueAt int64, newPriority int) (foreignUserID string, list string, oldMessage string, err error)
	// ChangeAssignment updates an issue to assign a different person
	ChangeAssignment(issueID string, userID string, sendTo string) (issue *Issue, oldOwner string, err error)
	// GetUserName returns the readable username from userID
	GetUserName(userID string) string
	// IsAuthorized checks if the user has access to the todo
	IsAuthorized(todoID string, userID string) (bool, error)
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
	client *pluginapi.Client

	BotUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	router *mux.Router

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	listManager ListManager
	store       ListStore

	telemetryClient telemetry.Client
	tracker         telemetry.Tracker
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	if err := p.loadI18n(); err != nil {
		return errors.Wrap(err, "failed to load i18n")
	}

	if p.client == nil {
		p.client = pluginapi.NewClient(p.API, p.Driver)
	}

	botID, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "todo",
		DisplayName: "Todo Bot",
		Description: "Created by the Todo plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure todo bot")
	}
	p.BotUserID = botID

	sqlStore, err := NewSQLStore(p.API)
	if err != nil {
		return errors.Wrap(err, "failed to initialize SQL store")
	}
	p.store = sqlStore

	p.listManager = NewListManager(p.API, sqlStore)

	p.initializeAPI()

	p.telemetryClient, err = telemetry.NewRudderClient()
	if err != nil {
		p.API.LogWarn("telemetry client not started", "error", err.Error())
	}

	return p.API.RegisterCommand(getCommand())
}

func (p *Plugin) OnDeactivate() error {
	if p.telemetryClient != nil {
		err := p.telemetryClient.Close()
		if err != nil {
			p.API.LogWarn("OnDeactivate: failed to close telemetryClient", "error", err.Error())
		}
	}

	return nil
}

func (p *Plugin) initializeAPI() {
	p.router = mux.NewRouter()
	p.router.Use(p.withRecovery)

	p.router.Handle("/add", p.checkAuth(http.HandlerFunc(p.handleAdd))).Methods(http.MethodPost)
	p.router.Handle("/lists", p.checkAuth(http.HandlerFunc(p.handleLists))).Methods(http.MethodGet)
	p.router.Handle("/remove", p.checkAuth(http.HandlerFunc(p.handleRemove))).Methods(http.MethodPost)
	p.router.Handle("/complete", p.checkAuth(http.HandlerFunc(p.handleComplete))).Methods(http.MethodPost)
	p.router.Handle("/accept", p.checkAuth(http.HandlerFunc(p.handleAccept))).Methods(http.MethodPost)
	p.router.Handle("/bump", p.checkAuth(http.HandlerFunc(p.handleBump))).Methods(http.MethodPost)
	p.router.Handle("/telemetry", p.checkAuth(http.HandlerFunc(p.handleTelemetry))).Methods(http.MethodPost)
	p.router.Handle("/config", p.checkAuth(http.HandlerFunc(p.handleConfig))).Methods(http.MethodGet)
	p.router.Handle("/edit", p.checkAuth(http.HandlerFunc(p.handleEdit))).Methods(http.MethodPut)
	p.router.Handle("/change_assignment", p.checkAuth(http.HandlerFunc(p.handleChangeAssignment))).Methods(http.MethodPost)

	commentsRouter := p.router.PathPrefix("/comments").Subrouter()
	commentsRouter.Use(p.checkAuth) // Apply checkAuth to all comments routes

	commentsRouter.HandleFunc("/get", p.handleGetComments).Methods(http.MethodGet)
	commentsRouter.HandleFunc("/add", p.handleAddComment).Methods(http.MethodPost)
	commentsRouter.HandleFunc("/delete", p.handleDeleteComment).Methods(http.MethodPost)

	// 404 handler
	p.router.Handle("{anything:.*}", http.NotFoundHandler())
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p.router.ServeHTTP(w, r)
}

func (p *Plugin) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				p.API.LogWarn("Recovered from a panic",
					"url", r.URL.String(),
					"error", x,
					"stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	telemetryRequest, err := GetTelemetryPayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get telemetry payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = telemetryRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate telemetry payload.", err)
		return
	}

	if telemetryRequest.Event != "" {
		p.trackFrontend(userID, telemetryRequest.Event, telemetryRequest.Properties)
	}
}

func (p *Plugin) handleAdd(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	addRequest, err := GetAddIssuePayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get add issue payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = addRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate add issue payload.", err)
		return
	}

	senderName := p.listManager.GetUserName(userID)

	if addRequest.SendTo == "" {
		_, err = p.listManager.AddIssue(userID, addRequest.Message, addRequest.PostPermalink, addRequest.Description, addRequest.PostID, addRequest.DueAt, addRequest.Priority)
		if err != nil {
			p.API.LogError(ErrorMsgAddIssue, "err", err.Error())
			p.handleErrorWithCode(w, http.StatusInternalServerError, ErrorMsgAddIssue, err)
			return
		}

		p.trackAddIssue(userID, sourceWebapp, addRequest.PostID != "")

		p.sendRefreshEvent(userID, []string{MyListKey})

		replyMessage := fmt.Sprintf("@%s attached a todo to this thread", senderName)
		p.postReplyIfNeeded(addRequest.PostID, replyMessage, addRequest.Message, addRequest.PostPermalink)

		return
	}

	receiver, appErr := p.API.GetUserByUsername(addRequest.SendTo)
	if appErr != nil {
		msg := "Unable to find user"
		p.API.LogError(msg, "err", appErr.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, appErr)
		return
	}

	if receiver.Id == userID {
		_, err = p.listManager.AddIssue(userID, addRequest.Message, addRequest.PostPermalink, addRequest.Description, addRequest.PostID, addRequest.DueAt, addRequest.Priority)
		if err != nil {
			p.API.LogError(ErrorMsgAddIssue, "err", err.Error())
			p.handleErrorWithCode(w, http.StatusInternalServerError, ErrorMsgAddIssue, err)
			return
		}

		p.trackAddIssue(userID, sourceWebapp, addRequest.PostID != "")

		p.sendRefreshEvent(userID, []string{MyListKey})

		replyMessage := fmt.Sprintf("@%s attached a todo to this thread", senderName)
		p.postReplyIfNeeded(addRequest.PostID, replyMessage, addRequest.Message, addRequest.PostPermalink)
		return
	}

	receiverAllowIncomingTaskRequestsPreference, err := p.getAllowIncomingTaskRequestsPreference(receiver.Id)
	if err != nil {
		p.API.LogError("Error when getting allow incoming task request preference, err=", err)
		receiverAllowIncomingTaskRequestsPreference = true
	}
	if !receiverAllowIncomingTaskRequestsPreference {
		replyMessage := fmt.Sprintf("@%s has blocked Todo requests", receiver.Username)
		p.PostBotDM(userID, replyMessage)
		return
	}

	issueID, err := p.listManager.SendIssue(userID, receiver.Id, addRequest.Message, addRequest.PostPermalink, addRequest.Description, addRequest.PostID, addRequest.DueAt, addRequest.Priority)
	if err != nil {
		msg := "Unable to send issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.trackSendIssue(userID, sourceWebapp, addRequest.PostID != "")

	p.sendRefreshEvent(userID, []string{OutListKey})
	p.sendRefreshEvent(receiver.Id, []string{InListKey})

	receiverMessage := fmt.Sprintf("You have received a new Todo from @%s", senderName)
	p.PostBotCustomDM(receiver.Id, receiverMessage, addRequest.Message, addRequest.PostPermalink, issueID)

	replyMessage := fmt.Sprintf("@%s sent @%s a todo attached to this thread", senderName, addRequest.SendTo)
	p.postReplyIfNeeded(addRequest.PostID, replyMessage, addRequest.Message, addRequest.PostPermalink)
}

func (p *Plugin) postReplyIfNeeded(postID, message, todo, postPermalink string) {
	if postID != "" {
		err := p.ReplyPostBot(postID, message, todo, postPermalink)
		if err != nil {
			p.API.LogError(err.Error())
		}
	}
}

func (p *Plugin) handleLists(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	allListIssue, err := p.listManager.GetAllList(userID)
	if err != nil {
		msg := "Unable to get issues for user"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	if allListIssue != nil && len(allListIssue.My) > 0 && r.URL.Query().Get("reminder") == "true" && p.getReminderPreference(userID) {
		var lastReminderAt int64
		lastReminderAt, err = p.getLastReminderTimeForUser(userID)
		if err != nil {
			msg := "Unable to send reminder"
			p.API.LogError(msg, "err", err.Error())
			p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
			return
		}

		var timezone *time.Location
		offset, _ := strconv.Atoi(r.Header.Get("X-Timezone-Offset"))
		timezone = time.FixedZone("local", -60*offset)

		// Post reminder message if it's the next day and been more than an hour since the last post
		now := model.GetMillis()
		nt := time.Unix(now/1000, 0).In(timezone)
		lt := time.Unix(lastReminderAt/1000, 0).In(timezone)
		if nt.Sub(lt).Hours() >= 1 && (nt.Day() != lt.Day() || nt.Month() != lt.Month() || nt.Year() != lt.Year()) {
			p.PostBotDM(userID, "Daily Reminder:\n\n"+issuesListToString(allListIssue.My))
			p.trackDailySummary(userID)
			err = p.saveLastReminderTimeForUser(userID)
			if err != nil {
				p.API.LogError("Unable to save last reminder for user err=" + err.Error())
			}
		}
	}

	allListIssueJSON, err := json.Marshal(allListIssue)
	if err != nil {
		msg := "Unable marhsal all lists issues to json"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	_, err = w.Write(allListIssueJSON)
	if err != nil {
		p.API.LogError("Unable to write json response while listing issues err=" + err.Error())
	}
}

func (p *Plugin) handleEdit(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	editRequest, err := GetEditIssuePayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get edit issue payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = editRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate edit issue payload.", err)
		return
	}

	if !p.checkAuthorization(w, editRequest.ID, userID) {
		return
	}

	foreignUserID, list, oldMessage, err := p.listManager.EditIssue(userID, editRequest.ID, editRequest.Message, editRequest.Description, editRequest.DueAt, editRequest.Priority)
	if err != nil {
		msg := "Unable to edit message"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.trackEditIssue(userID)
	p.sendRefreshEvent(userID, []string{list})

	if foreignUserID != "" {
		var lists []string
		if list == OutListKey {
			lists = []string{MyListKey, InListKey}
		} else {
			lists = []string{OutListKey}
		}
		p.sendRefreshEvent(foreignUserID, lists)

		userName := p.listManager.GetUserName(userID)
		message := fmt.Sprintf("@%s modified a Todo from:\n%s\nTo:\n%s", userName, oldMessage, editRequest.Message)
		p.PostBotDM(foreignUserID, message)
	}
}

func (p *Plugin) handleChangeAssignment(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	changeRequest, err := GetChangeAssignmentPayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get change request payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = changeRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate change assignment request payload.", err)
		return
	}

	if !p.checkAuthorization(w, changeRequest.ID, userID) {
		return
	}

	receiver, appErr := p.API.GetUserByUsername(changeRequest.SendTo)
	if appErr != nil {
		msg := "username not valid"
		p.API.LogError(msg, "err", appErr.Error())
		p.handleErrorWithCode(w, http.StatusNotFound, msg, appErr)
		return
	}

	issue, oldOwner, err := p.listManager.ChangeAssignment(changeRequest.ID, userID, receiver.Id)
	if err != nil {
		msg := "Unable to change the assignment of an issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.trackChangeAssignment(userID)

	p.sendRefreshEvent(userID, []string{MyListKey, OutListKey})

	userName := p.listManager.GetUserName(userID)
	if receiver.Id != userID {
		p.sendRefreshEvent(receiver.Id, []string{InListKey})
		receiverMessage := fmt.Sprintf("You have received a new Todo from @%s", userName)
		p.PostBotCustomDM(receiver.Id, receiverMessage, issue.Message, issue.PostPermalink, changeRequest.ID)
	}
	if oldOwner != "" {
		p.sendRefreshEvent(oldOwner, []string{InListKey, MyListKey})
		oldOwnerMessage := fmt.Sprintf("@%s removed you from Todo:\n%s", userName, issue.Message)
		p.PostBotDM(oldOwner, oldOwnerMessage)
	}
}

func (p *Plugin) handleAccept(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	acceptRequest, err := GetAcceptRequestPayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get accept request payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = acceptRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate accept request payload.", err)
		return
	}

	if !p.checkAuthorization(w, acceptRequest.ID, userID) {
		return
	}

	todoMessage, sender, err := p.listManager.AcceptIssue(userID, acceptRequest.ID)
	if err != nil {
		msg := "Unable to accept issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.trackAcceptIssue(userID)

	p.sendRefreshEvent(userID, []string{MyListKey, InListKey})
	p.sendRefreshEvent(sender, []string{OutListKey})

	userName := p.listManager.GetUserName(userID)
	message := fmt.Sprintf("@%s accepted a Todo you sent: %s", userName, todoMessage)
	p.PostBotDM(sender, message)
}

func (p *Plugin) handleComplete(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	completeRequest, err := GetCompleteIssuePayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get complete issue request payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = completeRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate complete issue request payload.", err)
		return
	}

	if !p.checkAuthorization(w, completeRequest.ID, userID) {
		return
	}

	issue, foreignID, listToUpdate, err := p.listManager.CompleteIssue(userID, completeRequest.ID)
	if err != nil {
		msg := "Unable to complete issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.sendRefreshEvent(userID, []string{listToUpdate})

	p.trackCompleteIssue(userID)

	userName := p.listManager.GetUserName(userID)
	replyMessage := fmt.Sprintf("@%s completed a todo attached to this thread", userName)
	p.postReplyIfNeeded(issue.PostID, replyMessage, issue.Message, issue.PostPermalink)

	if foreignID == "" {
		return
	}

	p.sendRefreshEvent(foreignID, []string{OutListKey})

	if issue.PostPermalink != "" {
		issue.Message = fmt.Sprintf("%s\n[Permalink](%s)", issue.Message, issue.PostPermalink)
	}

	message := fmt.Sprintf("@%s completed a Todo you sent: %s", userName, issue.Message)
	p.PostBotDM(foreignID, message)
}

func (p *Plugin) handleRemove(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	removeRequest, err := GetRemoveIssuePayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get remove issue request payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = removeRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate remove issue request payload.", err)
		return
	}

	if !p.checkAuthorization(w, removeRequest.ID, userID) {
		return
	}

	issue, foreignID, isSender, listToUpdate, err := p.listManager.RemoveIssue(userID, removeRequest.ID)
	if err != nil {
		msg := "Unable to remove issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}
	p.sendRefreshEvent(userID, []string{listToUpdate})

	p.trackRemoveIssue(userID)

	userName := p.listManager.GetUserName(userID)
	replyMessage := fmt.Sprintf("@%s removed a todo attached to this thread", userName)
	p.postReplyIfNeeded(issue.PostID, replyMessage, issue.Message, issue.PostPermalink)

	if foreignID == "" {
		return
	}

	list := InListKey

	message := fmt.Sprintf("@%s removed a Todo you received: %s", userName, issue.Message)
	if isSender {
		message = fmt.Sprintf("@%s declined a Todo you sent: %s", userName, issue.Message)
		list = OutListKey
	}

	if issue.PostPermalink != "" {
		message = fmt.Sprintf("%s\n[Permalink](%s)", message, issue.PostPermalink)
	}

	p.sendRefreshEvent(foreignID, []string{list})

	p.PostBotDM(foreignID, message)
}

func (p *Plugin) handleBump(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	bumpRequest, err := GetBumpIssuePayloadFromJSON(r.Body)
	if err != nil {
		msg := "Unable to get bump issue request payload from JSON"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, msg, err)
		return
	}

	if err = bumpRequest.IsValid(); err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to validate bump issue request payload.", err)
		return
	}

	if !p.checkAuthorization(w, bumpRequest.ID, userID) {
		return
	}

	todo, foreignUser, foreignIssueID, err := p.listManager.BumpIssue(userID, bumpRequest.ID)
	if err != nil {
		msg := "Unable to bump issue"
		p.API.LogError(msg, "err", err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
		return
	}

	p.trackBumpIssue(userID)

	if foreignUser == "" {
		return
	}

	p.sendRefreshEvent(foreignUser, []string{InListKey})

	userName := p.listManager.GetUserName(userID)
	message := fmt.Sprintf("@%s bumped a Todo you received.", userName)
	p.PostBotCustomDM(foreignUser, message, todo.Message, todo.PostPermalink, foreignIssueID)
}

// API endpoint to retrieve plugin configurations
func (p *Plugin) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if p.configuration != nil {
		// retrieve client only configurations
		clientConfig := struct {
			HideTeamSidebar bool `json:"hide_team_sidebar"`
		}{
			HideTeamSidebar: p.configuration.HideTeamSidebar,
		}

		configJSON, err := json.Marshal(clientConfig)
		if err != nil {
			msg := "Unable to marshal plugin configuration to json"
			p.API.LogError(msg, "err", err.Error())
			p.handleErrorWithCode(w, http.StatusInternalServerError, msg, err)
			return
		}

		_, err = w.Write(configJSON)
		if err != nil {
			p.API.LogError("Unable to write json response err=" + err.Error())
		}
	}
}

func (p *Plugin) sendRefreshEvent(userID string, lists []string) {
	p.API.PublishWebSocketEvent(
		WSEventRefresh,
		map[string]interface{}{"lists": lists},
		&model.WebsocketBroadcast{UserId: userID},
	)
}

// Publish a WebSocket event to update the client config of the plugin on the webapp end.
func (p *Plugin) sendConfigUpdateEvent() {
	clientConfigMap := map[string]interface{}{
		"hide_team_sidebar": p.configuration.HideTeamSidebar,
	}

	p.API.PublishWebSocketEvent(
		WSEventConfigUpdate,
		clientConfigMap,
		&model.WebsocketBroadcast{},
	)
}

func (p *Plugin) handleErrorWithCode(w http.ResponseWriter, code int, errTitle string, err error) {
	w.WriteHeader(code)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   errTitle,
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}

func (p *Plugin) handleGetComments(w http.ResponseWriter, r *http.Request) {
	todoID := r.URL.Query().Get("id")
	if todoID == "" {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Missing id parameter", nil)
		return
	}

	userID := r.Header.Get("Mattermost-User-ID")

	if !p.checkAuthorization(w, todoID, userID) {
		return
	}

	comments, err := p.listManager.GetIssueComments(todoID)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to get comments", err)
		return
	}

	b, _ := json.Marshal(comments)
	w.Write(b)
}

func (p *Plugin) handleAddComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	req, err := GetCommentPayloadFromJSON(r.Body)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to parse comment payload", err)
		return
	}

	if req.TodoID == "" || req.Message == "" {
		p.handleErrorWithCode(w, http.StatusBadRequest, "TodoID and Message are required", nil)
		return
	}

	if !p.checkAuthorization(w, req.TodoID, userID) {
		return
	}

	comment, err := p.listManager.AddComment(req.TodoID, userID, req.Message)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to add comment", err)
		return
	}

	p.sendRefreshEvent(userID, []string{MyListKey, InListKey, OutListKey})
	// Also notify other involved user if any
	issue, _ := p.store.GetIssue(req.TodoID)
	if issue != nil {
		otherUser := issue.CreatorID
		if userID == issue.CreatorID {
			otherUser = issue.AssigneeID
		}
		if otherUser != userID {
			p.sendRefreshEvent(otherUser, []string{MyListKey, InListKey, OutListKey})
		}
	}

	b, _ := json.Marshal(comment)
	w.Write(b)
}

func (p *Plugin) checkAuthorization(w http.ResponseWriter, todoID, userID string) bool {
	authorized, err := p.listManager.IsAuthorized(todoID, userID)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to check authorization", err)
		return false
	}
	if !authorized {
		p.handleErrorWithCode(w, http.StatusForbidden, "Not authorized for this todo", nil)
		return false
	}
	return true
}

func (p *Plugin) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")

	req, err := GetCommentPayloadFromJSON(r.Body)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to parse comment payload", err)
		return
	}

	if req.ID == "" {
		p.handleErrorWithCode(w, http.StatusBadRequest, "Comment ID is required", nil)
		return
	}

	comment, err := p.store.GetComment(req.ID)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to fetch comment", err)
		return
	}

	err = p.listManager.DeleteComment(req.ID, userID)
	if err != nil {
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to delete comment", err)
		return
	}

	p.sendRefreshEvent(userID, []string{MyListKey, InListKey, OutListKey})
	issue, _ := p.store.GetIssue(comment.TodoID)
	if issue != nil {
		otherUser := issue.CreatorID
		if userID == issue.CreatorID {
			otherUser = issue.AssigneeID
		}
		if otherUser != userID {
			p.sendRefreshEvent(otherUser, []string{MyListKey, InListKey, OutListKey})
		}
	}

	w.Write([]byte(`{"status": "OK"}`))
}
