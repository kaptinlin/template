package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test time for consistency across tests.
func testTime() time.Time {
	return time.Date(2024, 3, 30, 15, 4, 5, 0, time.UTC)
}

func TestDateFilters(t *testing.T) {
	// Mock current time for consistent testing
	currentTime := testTime()

	// Define test cases with template string and expected output
	cases := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "FormattedDate",
			template: "Formatted date: {{ current | date:'F j, Y' }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Formatted date: March 30, 2024",
		},
		{
			name:     "DayAndMonthNames",
			template: "Day and month: {{ current | date:'l, F' }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Day and month: Saturday, March",
		},
		{
			name:     "TimeWithAmPm",
			template: "Time: {{ current | date:'g:i A' }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Time: 3:04 PM",
		},
		{
			name:     "DateFormatWithoutFormat",
			template: "Default date format: {{ current | date }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Default date format: 2024-03-30 15:04:05",
		},
		{
			name:     "WeekOfYearAndDayOfWeek",
			template: "Week of year and day of week: {{ current | date:'W, N' }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Week of year and day of week: 13, 6",
		},
		{
			name:     "UnixTimestamp",
			template: "Unix timestamp: {{ current | date:'S' }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Unix timestamp: 1711811045",
		},
		{
			name:     "TimeAgo",
			template: "Time ago: {{ past | timeago }}",
			context: map[string]interface{}{
				"past": time.Now().Add(-5 * 24 * time.Hour),
			},
			expected: "Time ago: 5 days ago",
		},
		{
			name:     "MonthAsNumber",
			template: "Current month: {{ current | month }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Current month: 3",
		},
		{
			name:     "FullMonthName",
			template: "Current month: {{ current | month_full }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Current month: March",
		},
		{
			name:     "Year",
			template: "Current year: {{ current | year }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Current year: 2024",
		},
		{
			name:     "Day",
			template: "Current day: {{ current | day }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Current day: 30",
		},
		{
			name:     "Weekday",
			template: "Current day of the week: {{ current | weekday }}",
			context:  map[string]interface{}{"current": currentTime},
			expected: "Current day of the week: Saturday",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the template
			tpl, err := Parse(tc.template)
			require.NoError(t, err, "Failed to parse template")

			// Create a context and add variables
			context := NewContext()
			for k, v := range tc.context {
				context.Set(k, v)
			}

			// Execute the template
			output, err := Execute(tpl, context)
			require.NoError(t, err, "Failed to execute template")

			// Verify the output
			assert.Equal(t, tc.expected, output)
		})
	}
}
