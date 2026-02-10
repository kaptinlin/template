package template

import (
	"github.com/kaptinlin/filter"
)

// registerDateFilters registers all date-related filters.
func registerDateFilters() {
	filters := map[string]FilterFunc{
		"date":       dateFilter,
		"month":      monthFilter,
		"monthFull":  monthFullFilter,
		"month_full": monthFullFilter, // Alias for backward compatibility
		"year":       yearFilter,
		"day":        dayFilter,
		"week":       weekFilter,
		"weekday":    weekdayFilter,
		"timeAgo":    timeAgoFilter,
		"timeago":    timeAgoFilter, // Alias for backward compatibility
	}

	for name, fn := range filters {
		RegisterFilter(name, fn)
	}
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
