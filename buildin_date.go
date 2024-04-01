package template

import (
	"github.com/kaptinlin/filter"
)

func init() {
	// Register all date filters
	RegisterFilter("date", dateFilter)
	RegisterFilter("day", dayFilter)
	RegisterFilter("month", monthFilter)
	RegisterFilter("month_full", monthFullFilter)
	RegisterFilter("year", yearFilter)
	RegisterFilter("week", weekFilter)
	RegisterFilter("weekday", weekdayFilter)
	RegisterFilter("timeago", timeAgoFilter)
}

// dateFilter formats a timestamp into a specified format.
func dateFilter(value interface{}, args ...string) (interface{}, error) {
	format := ""
	if len(args) > 0 {
		format = args[0]
	}
	return filter.Date(value, format)
}

// dayFilter extracts and returns the day of the month.
func dayFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Day(value)
}

// monthFilter extracts and returns the month number.
func monthFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Month(value)
}

// monthFullFilter returns the full month name.
func monthFullFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.MonthFull(value)
}

// yearFilter extracts and returns the year.
func yearFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Year(value)
}

// weekFilter returns the ISO week number.
func weekFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Week(value)
}

// weekdayFilter returns the day of the week.
func weekdayFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.Weekday(value)
}

// timeAgoFilter returns a human-readable string representing the time difference.
func timeAgoFilter(value interface{}, args ...string) (interface{}, error) {
	return filter.TimeAgo(value)
}
