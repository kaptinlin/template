package template

import "github.com/kaptinlin/filter"

// registerDateFilters registers all date-related filters.
func registerDateFilters() {
	defaultRegistry.MustRegister("date", dateFilter)
	defaultRegistry.MustRegister("month", monthFilter)
	defaultRegistry.MustRegister("month_full", monthFullFilter)
	defaultRegistry.MustRegister("year", yearFilter)
	defaultRegistry.MustRegister("day", dayFilter)
	defaultRegistry.MustRegister("week", weekFilter)
	defaultRegistry.MustRegister("weekday", weekdayFilter)
	defaultRegistry.MustRegister("time_ago", timeAgoFilter)

	// Aliases
	defaultRegistry.MustRegister("timeago", timeAgoFilter)
}

// dateFilter formats a timestamp into a specified format using PHP-style format strings.
// An empty format string uses the default "2006-01-02 15:04:05" layout.
func dateFilter(value any, args ...any) (any, error) {
	var format string
	if len(args) > 0 {
		format = toString(args[0])
	}
	return filter.Date(value, format)
}

// dayFilter extracts and returns the day of the month.
func dayFilter(value any, _ ...any) (any, error) {
	return filter.Day(value)
}

// monthFilter extracts and returns the month number.
func monthFilter(value any, _ ...any) (any, error) {
	return filter.Month(value)
}

// monthFullFilter returns the full month name.
func monthFullFilter(value any, _ ...any) (any, error) {
	return filter.MonthFull(value)
}

// yearFilter extracts and returns the year.
func yearFilter(value any, _ ...any) (any, error) {
	return filter.Year(value)
}

// weekFilter returns the ISO week number.
func weekFilter(value any, _ ...any) (any, error) {
	return filter.Week(value)
}

// weekdayFilter returns the day of the week.
func weekdayFilter(value any, _ ...any) (any, error) {
	return filter.Weekday(value)
}

// timeAgoFilter returns a human-readable string representing the time difference.
func timeAgoFilter(value any, _ ...any) (any, error) {
	return filter.TimeAgo(value)
}
