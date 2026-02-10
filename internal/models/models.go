package models

import "time"

type Event struct {
	NotionPageID    string
	Title           string
	StartDate       string
	StartTime       string
	EndDate         string
	EndTime         string
	IsAllDay        bool
	Location        string
	URL             string
	Content         string
	Custom          map[string]string
	RawPropsJSON    string
	FetchedAt       time.Time
	NotionUpdatedAt string
	Attendees       []string // email addresses extracted from Notion people property
}

type NotificationHistory struct {
	ID           int64
	Type         string
	Status       string
	Message      string
	NotionPageID string
	Error        string
	SentAt       time.Time
}

type AdvanceSchedule struct {
	ID           int64
	NotionPageID string
	RuleIndex    int
	FireAt       time.Time
	Fired        bool
}

type SyncRecord struct {
	NotionPageID    string
	CalendarEventID string
	NotionUpdatedAt string
	Synced          bool
}

type TemplateEvent struct {
	Name          string
	Date          string
	Time          string
	EndTime       string
	IsAllDay      bool
	Location      string
	URL           string
	Content       string
	MinutesBefore int
	Custom        map[string]string
}

type TemplateContext struct {
	Events        []TemplateEvent
	MinutesBefore int
}

type WebhookPayloadContext struct {
	Type          string
	Message       string
	Events        []TemplateEvent
	Event         TemplateEvent
	MinutesBefore int
}
