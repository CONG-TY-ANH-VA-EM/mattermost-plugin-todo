package main

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// saveLastReminderTimeForUser saves the last time a user was reminded
func (p *Plugin) saveLastReminderTimeForUser(userID string) error {
	return p.store.SetLastReminderTime(userID, model.GetMillis())
}

// getLastReminderTimeForUser gets the last time a user was reminded
func (p *Plugin) getLastReminderTimeForUser(userID string) (int64, error) {
	return p.store.GetLastReminderTime(userID)
}

// saveReminderPreference saves user preference on reminder
func (p *Plugin) saveReminderPreference(userID string, preference bool) error {
	return p.store.SetReminderPreference(userID, preference)
}

// getReminderPreference gets user preference on reminder
func (p *Plugin) getReminderPreference(userID string) bool {
	return p.store.GetReminderPreference(userID)
}

// saveAllowIncomingTaskRequestsPreference saves user preference on allowing incoming task requests
func (p *Plugin) saveAllowIncomingTaskRequestsPreference(userID string, preference bool) error {
	return p.store.SetAllowIncomingTaskPreference(userID, preference)
}

// getAllowIncomingTaskRequestsPreference gets user preference on allowing incoming task requests
func (p *Plugin) getAllowIncomingTaskRequestsPreference(userID string) (bool, error) {
	return p.store.GetAllowIncomingTaskPreference(userID)
}
