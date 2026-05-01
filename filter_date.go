package template

import "github.com/kaptinlin/filter"

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

func dayFilter(value any, _ ...any) (any, error) {
	return filter.Day(value)
}

func monthFilter(value any, _ ...any) (any, error) {
	return filter.Month(value)
}

func monthFullFilter(value any, _ ...any) (any, error) {
	return filter.MonthFull(value)
}

func yearFilter(value any, _ ...any) (any, error) {
	return filter.Year(value)
}

func weekFilter(value any, _ ...any) (any, error) {
	return filter.Week(value)
}

func weekdayFilter(value any, _ ...any) (any, error) {
	return filter.Weekday(value)
}

func timeAgoFilter(value any, _ ...any) (any, error) {
	return filter.TimeAgo(value)
}
