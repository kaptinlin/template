package template

import "github.com/kaptinlin/filter"

// registerDateFilters registers all date-related filters.
func registerDateFilters() {
	RegisterFilter("date", dateFilter)
	RegisterFilter("month", monthFilter)
	RegisterFilter("monthFull", monthFullFilter)
	RegisterFilter("month_full", monthFullFilter)
	RegisterFilter("year", yearFilter)
	RegisterFilter("day", dayFilter)
	RegisterFilter("week", weekFilter)
	RegisterFilter("weekday", weekdayFilter)
	RegisterFilter("timeAgo", timeAgoFilter)
	RegisterFilter("timeago", timeAgoFilter)
}

// dateFilter formats a timestamp into a specified format using PHP-style format strings.
// An empty format string uses the default "2006-01-02 15:04:05" layout.
func dateFilter(value any, args ...string) (any, error) {
	var format string
	if len(args) > 0 {
		format = args[0]
	}
	return filter.Date(value, format)
}

// dayFilter extracts and returns the day of the month.
func dayFilter(value any, _ ...string) (any, error) {
	return filter.Day(value)
}

// monthFilter extracts and returns the month number.
func monthFilter(value any, _ ...string) (any, error) {
	return filter.Month(value)
}

// monthFullFilter returns the full month name.
func monthFullFilter(value any, _ ...string) (any, error) {
	return filter.MonthFull(value)
}

// yearFilter extracts and returns the year.
func yearFilter(value any, _ ...string) (any, error) {
	return filter.Year(value)
}

// weekFilter returns the ISO week number.
func weekFilter(value any, _ ...string) (any, error) {
	return filter.Week(value)
}

// weekdayFilter returns the day of the week.
func weekdayFilter(value any, _ ...string) (any, error) {
	return filter.Weekday(value)
}

// timeAgoFilter returns a human-readable string representing the time difference.
func timeAgoFilter(value any, _ ...string) (any, error) {
	return filter.TimeAgo(value)
}
