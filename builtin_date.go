package template

import (
	"github.com/kaptinlin/filter"
)

func init() {
	mustRegisterFilters(map[string]FilterFunc{
		"date":       dateFilter,
		"day":        dayFilter,
		"month":      monthFilter,
		"month_full": monthFullFilter,
		"year":       yearFilter,
		"week":       weekFilter,
		"weekday":    weekdayFilter,
		"timeago":    timeAgoFilter,
	})
}

// dateFilter formats a timestamp into a specified format.
func dateFilter(value any, args ...string) (any, error) {
	format := ""
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
