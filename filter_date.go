package template

import (
	"github.com/kaptinlin/filter"
)

// registerDateFilters registers all date-related filters.
func registerDateFilters() {
	RegisterFilter("date", dateFilter)
	RegisterFilter("month", monthFilter)
	RegisterFilter("monthFull", monthFullFilter)
	RegisterFilter("month_full", monthFullFilter) // Alias for backward compatibility
	RegisterFilter("year", yearFilter)
	RegisterFilter("day", dayFilter)
	RegisterFilter("week", weekFilter)
	RegisterFilter("weekday", weekdayFilter)
	RegisterFilter("timeAgo", timeAgoFilter)
	RegisterFilter("timeago", timeAgoFilter) // Alias for backward compatibility
}

// dateFilter formats a timestamp into a specified format using PHP-style format strings.
func dateFilter(value interface{}, args ...string) (interface{}, error) {
	format := ""
	if len(args) > 0 {
		format = args[0]
	}
	return filter.Date(value, format)
}

// dayFilter extracts and returns the day of the month.
func dayFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Day(value)
}

// monthFilter extracts and returns the month number.
func monthFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Month(value)
}

// monthFullFilter returns the full month name.
func monthFullFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.MonthFull(value)
}

// yearFilter extracts and returns the year.
func yearFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Year(value)
}

// weekFilter returns the ISO week number.
func weekFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Week(value)
}

// weekdayFilter returns the day of the week.
func weekdayFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.Weekday(value)
}

// timeAgoFilter returns a human-readable string representing the time difference.
func timeAgoFilter(value interface{}, _ ...string) (interface{}, error) {
	return filter.TimeAgo(value)
}
