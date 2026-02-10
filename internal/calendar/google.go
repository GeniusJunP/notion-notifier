package calendar

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"notion-notifier/internal/models"
)

type Client struct {
	srv        *calendarapi.Service
	calendarID string
}

func NewClient(ctx context.Context, calendarID string, serviceAccountKey string) (*Client, error) {
	if calendarID == "" {
		return nil, errors.New("google calendar id is empty")
	}
	creds, err := loadCredentials(serviceAccountKey)
	if err != nil {
		return nil, err
	}
	srv, err := calendarapi.NewService(ctx, option.WithCredentialsJSON(creds), option.WithScopes(calendarapi.CalendarScope))
	if err != nil {
		return nil, err
	}
	return &Client{srv: srv, calendarID: calendarID}, nil
}

func loadCredentials(value string) ([]byte, error) {
	if value == "" {
		return nil, errors.New("google service account key is empty")
	}
	if strings.HasPrefix(strings.TrimSpace(value), "{") {
		return []byte(value), nil
	}
	data, err := os.ReadFile(value)
	if err != nil {
		return nil, fmt.Errorf("read service account key: %w", err)
	}
	return data, nil
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

func (c *Client) GetEvent(ctx context.Context, eventID string) (*calendarapi.Event, error) {
	return c.srv.Events.Get(c.calendarID, eventID).Context(ctx).Do()
}

func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	return c.srv.Events.Delete(c.calendarID, eventID).Context(ctx).Do()
}

// CalendarEvent is a simplified representation of a Google Calendar event with Notion metadata.
type CalendarEvent struct {
	ID           string
	NotionPageID string // from extendedProperties.private.notion_page_id
	Summary      string
	Updated      string
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
			Fields("items(id,summary,updated,extendedProperties),nextPageToken")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("calendar list events: %w", err)
		}
		for _, item := range resp.Items {
			ce := CalendarEvent{
				ID:      item.Id,
				Summary: item.Summary,
				Updated: item.Updated,
			}
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

func ValidateServiceAccountKey(value string) error {
	creds, err := loadCredentials(value)
	if err != nil {
		return err
	}
	_, err = google.JWTConfigFromJSON(creds, calendarapi.CalendarScope)
	return err
}
