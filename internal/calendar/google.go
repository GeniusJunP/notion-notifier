package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"notion-notifier/internal/models"
)

type Client struct {
	srv        *calendarapi.Service
	calendarID string
}

type ClientOptions struct {
	CalendarID        string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthRefreshToken string
}

func (o ClientOptions) normalize() ClientOptions {
	o.CalendarID = strings.TrimSpace(o.CalendarID)
	o.OAuthClientID = strings.TrimSpace(o.OAuthClientID)
	o.OAuthClientSecret = strings.TrimSpace(o.OAuthClientSecret)
	o.OAuthRefreshToken = strings.TrimSpace(o.OAuthRefreshToken)
	return o
}

func (o ClientOptions) Validate() error {
	o = o.normalize()
	if o.CalendarID == "" {
		return errors.New("google calendar id is empty")
	}
	if o.OAuthClientID == "" || o.OAuthClientSecret == "" || o.OAuthRefreshToken == "" {
		return errors.New("google oauth credentials are incomplete")
	}
	return nil
}

func (o ClientOptions) IsConfigured() bool {
	return o.Validate() == nil
}

func (o ClientOptions) Fingerprint() string {
	o = o.normalize()
	sum := sha256.Sum256([]byte(strings.Join([]string{
		o.CalendarID,
		o.OAuthClientID,
		o.OAuthClientSecret,
		o.OAuthRefreshToken,
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

func NewClient(ctx context.Context, opts ClientOptions) (*Client, error) {
	opts = opts.normalize()
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	oauthCfg := &oauth2.Config{
		ClientID:     opts.OAuthClientID,
		ClientSecret: opts.OAuthClientSecret,
		Endpoint:     googleoauth.Endpoint,
		Scopes:       []string{calendarapi.CalendarScope},
	}
	tokenSource := oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: opts.OAuthRefreshToken})
	srv, err := calendarapi.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}
	return &Client{srv: srv, calendarID: opts.CalendarID}, nil
}

func (c *Client) UpsertEvent(ctx context.Context, ev models.Event, existingID string, tz *time.Location) (string, string, error) {
	gevent := mapEvent(ev, tz)
	if existingID != "" {
		updated, err := c.srv.Events.Update(c.calendarID, existingID, gevent).Context(ctx).Do()
		if err != nil {
			return "", "", err
		}
		return updated.Id, updated.Updated, nil
	}
	created, err := c.srv.Events.Insert(c.calendarID, gevent).Context(ctx).Do()
	if err != nil {
		return "", "", err
	}
	return created.Id, created.Updated, nil
}

func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	return c.srv.Events.Delete(c.calendarID, eventID).Context(ctx).Do()
}

// CalendarEvent is a simplified representation of a Google Calendar event with Notion metadata.
type CalendarEvent struct {
	ID            string
	NotionPageID  string // from extendedProperties.private.notion_page_id
	Summary       string
	Location      string
	Description   string
	StartDate     string
	StartDateTime string
	EndDate       string
	EndDateTime   string
	Attendees     []string
	Updated       string
}

// ListEvents fetches all events from Google Calendar in the given time range
// and returns those that have a notion_page_id in their extended properties.
func (c *Client) ListEvents(ctx context.Context, from, to time.Time) ([]CalendarEvent, error) {
	var result []CalendarEvent
	pageToken := ""
	for {
		call := c.srv.Events.List(c.calendarID).
			Context(ctx).
			TimeMin(from.Format(time.RFC3339)).
			TimeMax(to.Format(time.RFC3339)).
			SingleEvents(true).
			MaxResults(250).
			Fields("items(id,summary,location,description,start(date,dateTime),end(date,dateTime),attendees(email),updated,extendedProperties),nextPageToken")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("calendar list events: %w", err)
		}
		for _, item := range resp.Items {
			ce := CalendarEvent{
				ID:          item.Id,
				Summary:     item.Summary,
				Location:    item.Location,
				Description: item.Description,
				Updated:     item.Updated,
			}
			if item.Start != nil {
				ce.StartDate = item.Start.Date
				ce.StartDateTime = item.Start.DateTime
			}
			if item.End != nil {
				ce.EndDate = item.End.Date
				ce.EndDateTime = item.End.DateTime
			}
			ce.Attendees = extractEmails(item.Attendees)
			if item.ExtendedProperties != nil && item.ExtendedProperties.Private != nil {
				ce.NotionPageID = item.ExtendedProperties.Private["notion_page_id"]
			}
			// Only include events that were created by this tool
			if ce.NotionPageID != "" {
				result = append(result, ce)
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return result, nil
}

// EventMatchesNotion returns true when the Calendar event already matches
// the canonical event generated from Notion data.
func EventMatchesNotion(calEvent CalendarEvent, ev models.Event, tz *time.Location) bool {
	want := mapEvent(ev, tz)
	if want == nil || want.Start == nil || want.End == nil {
		return false
	}
	if calEvent.Summary != want.Summary {
		return false
	}
	if calEvent.Location != want.Location {
		return false
	}
	if calEvent.Description != want.Description {
		return false
	}
	if !sameDateOrDateTime(calEvent.StartDate, calEvent.StartDateTime, want.Start.Date, want.Start.DateTime) {
		return false
	}
	if !sameDateOrDateTime(calEvent.EndDate, calEvent.EndDateTime, want.End.Date, want.End.DateTime) {
		return false
	}
	return equalEmails(calEvent.Attendees, extractEmails(want.Attendees))
}

func sameDateOrDateTime(currentDate, currentDateTime, desiredDate, desiredDateTime string) bool {
	curDate := strings.TrimSpace(currentDate)
	desDate := strings.TrimSpace(desiredDate)
	curDateTime := normalizeDateTime(currentDateTime)
	desDateTime := normalizeDateTime(desiredDateTime)
	if curDate != "" || desDate != "" {
		return curDate == desDate && curDateTime == "" && desDateTime == ""
	}
	return curDateTime == desDateTime
}

func normalizeDateTime(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return t.UTC().Format(time.RFC3339)
}

func equalEmails(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func extractEmails(attendees []*calendarapi.EventAttendee) []string {
	if len(attendees) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(attendees))
	emails := make([]string, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee == nil {
			continue
		}
		email := strings.ToLower(strings.TrimSpace(attendee.Email))
		if email == "" {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		emails = append(emails, email)
	}
	sort.Strings(emails)
	if len(emails) == 0 {
		return nil
	}
	return emails
}

func mapEvent(ev models.Event, tz *time.Location) *calendarapi.Event {
	start, end := buildStartEnd(ev, tz)
	description := ev.Content
	if ev.URL != "" {
		if description != "" {
			description += "\n\n"
		}
		description += fmt.Sprintf("Notion: %s", ev.URL)
	}
	gevent := &calendarapi.Event{
		Summary:     ev.Title,
		Location:    ev.Location,
		Description: description,
		Start:       start,
		End:         end,
		ExtendedProperties: &calendarapi.EventExtendedProperties{
			Private: map[string]string{"notion_page_id": ev.NotionPageID},
		},
	}
	// Add attendees from Notion people property emails
	for _, email := range ev.Attendees {
		gevent.Attendees = append(gevent.Attendees, &calendarapi.EventAttendee{Email: email})
	}
	return gevent
}

func buildStartEnd(ev models.Event, tz *time.Location) (*calendarapi.EventDateTime, *calendarapi.EventDateTime) {
	loc := tz
	if loc == nil {
		loc = time.Local
	}
	if ev.IsAllDay || ev.StartTime == "" {
		startDate := ev.StartDate
		endDate := ev.EndDate
		if endDate == "" {
			endDate = startDate
		}
		startDay, _ := time.ParseInLocation("2006-01-02", startDate, loc)
		endDay, _ := time.ParseInLocation("2006-01-02", endDate, loc)
		if endDay.Before(startDay) {
			endDay = startDay
		}
		endDay = endDay.AddDate(0, 0, 1)
		return &calendarapi.EventDateTime{Date: startDay.Format("2006-01-02")}, &calendarapi.EventDateTime{Date: endDay.Format("2006-01-02")}
	}
	startTime, _ := time.ParseInLocation("2006-01-02 15:04", ev.StartDate+" "+ev.StartTime, loc)
	var endTime time.Time
	if ev.EndTime != "" && ev.EndDate != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04", ev.EndDate+" "+ev.EndTime, loc)
	} else {
		endTime = startTime.Add(1 * time.Hour)
	}
	return &calendarapi.EventDateTime{DateTime: startTime.Format(time.RFC3339), TimeZone: loc.String()},
		&calendarapi.EventDateTime{DateTime: endTime.Format(time.RFC3339), TimeZone: loc.String()}
}
