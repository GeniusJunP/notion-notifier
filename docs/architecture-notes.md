# Architecture Notes

## Deferred Data-Structure Work

This note records structural issues found while fixing snooze. They are intentionally not part of the snooze-only implementation.

### Event date/time normalization

`models.Event` still stores `StartDate`, `StartTime`, `EndDate`, `EndTime`, and `IsAllDay` as separate compatibility fields. Scheduler and Google Calendar code rebuild time ranges from those strings independently.

Future direction:

- Introduce a read-only `EventTimeRange` value that is derived from the existing fields.
- Move scheduler and calendar calculations to that value.
- Keep the current DB columns until there is a separate migration plan.

### Upcoming schedule identity

`upcoming_schedules` still uses `rule_index`. This is acceptable for the current narrow snooze change, but it is unstable if notification rules are reordered or deleted.

Future direction:

- Add stable IDs to upcoming notification rules.
- Store `rule_id` in `upcoming_schedules`.
- Treat the table as a derived cache and rebuild it when the schema changes.

### Time value types

`HH:mm` values are still plain strings and validated at config boundaries. This avoids broad churn in the snooze-only change.

Future direction:

- Introduce a `ClockTime` value type for config times.
- Introduce a `Timestamp` value type if more API fields need RFC3339/browser datetime conversion.

### UI configuration boundaries

Snooze now has a dedicated API, but several other UI controls still save the whole config. That is workable today but can create stale-write risks as the UI grows.

Future direction:

- Add narrow PATCH endpoints for independent settings groups.
- Keep advanced webhook payload settings separated from day-to-day operational controls.
