# Filters

Filters are versatile tools in template systems, allowing for data transformation, formatting, and presentation directly within the template. They are applied to variables or values using the pipe symbol (`|`) and can take additional arguments to fine-tune their behavior. Filters can be chained, enabling a sequence of transformations.

For those interested in exploring more about available filters and extending their functionality, visit [github.com/kaptinlin/filter](https://github.com/kaptinlin/filter).

## Basic Usage

To apply a filter, use the syntax:

```plaintext
{{ variable | filterName }}
```

For filters requiring arguments, include them after the filter name, separated by commas (`,`). Use quotes (`"`) for string arguments:

```plaintext
{{ "text" | filterName:arg1,"arg2" }}
```

## Chaining Filters

Multiple filters can be applied in sequence, where the output of one serves as the input to the next:

```plaintext
{{ variable | filterOne | filterTwo:"arg" }}
```

## Filter List

### String Functions

**Default**
Sets a default value if the original is empty.

```plaintext
{{ userName | default:"Guest" }}
Output: Guest
```

**Trim**
Removes leading and trailing spaces.

```plaintext
{{ "  Hello World  " | trim }}
Output: Hello World
```

**Split**
Splits a string into an array using the specified delimiter.

```plaintext
{{ "apple,banana,orange" | split:"," }}
Output: ["apple", "banana", "orange"]
```

**Replace**
Replaces occurrences of a substring with another string.

```plaintext
{{ "Hello World" | replace:"World","There" }}
Output: Hello There
```

**Remove**
Removes all occurrences of a specified substring.

```plaintext
{{ "Hello World" | remove:"World" }}
Output: Hello 
```

**Append**
Adds characters to the end of a string.

```plaintext
{{ "Hello" | append:". Goodbye." }}
Output: Hello. Goodbye.
```

**Prepend**
Adds characters to the beginning of a string.

```plaintext
{{ "World" | prepend:"Hello, " }}
Output: Hello, World
```

**Length**
Returns the length of the string.

```plaintext
{{ "Hello" | length }}
Output: 5
```

**Upper**
Converts the string to uppercase.

```plaintext
{{ "Hello World" | upper }}
Output: HELLO WORLD
```

**Lower**
Converts the string to lowercase.

```plaintext
{{ "Hello World" | lower }}
Output: hello world
```

**Titleize**
Capitalizes the first letter of each word.

```plaintext
{{ "hello world" | titleize }}
Output: Hello World
```

**Capitalize**
Capitalizes the first word in a string.

```plaintext
{{ "hello world" | capitalize }}
Output: Hello world
```

**Camelize**
Converts a string to camelCase.

```plaintext
{{ "hello-world" | camelize }}
Output: helloWorld
```

**Pascalize**
Converts a string to PascalCase.

```plaintext
{{ "hello-world" | pascalize }}
Output: HelloWorld
```

**Dasherize**
Converts a string to a lowercase, dash-separated format.

```plaintext
{{ "Hello World" | dasherize }}
Output: hello-world
```

**Slugify**
Generates a URL-friendly "slug" from a string.

```plaintext
{{ "Hello World & Friends" | slugify }}
Output: hello-world-and-friends
```

**Pluralize**
Outputs the singular or plural form of a word based on a numeric value.

```plaintext
{{ 1 | pluralize:"item","items" }}
Output: item
{{ 2 | pluralize:"item","items" }}
Output: items
```

**Ordinalize**
Converts a number to its ordinal English form.

```plaintext
{{ 1 | ordinalize }}
Output: 1st
{{ 2 | ordinalize }}
Output: 2nd
```

**Truncate**
Shortens a string to a specified length, appending "..." if truncated.

```plaintext
{{ "Hello World" | truncate:5 }}
Output: Hello...
```

**TruncateWords**
Truncates a string to a specified number of words.

```plaintext
{{ "Hello World Friends" | truncatewords:2 }}
Output: Hello World...
```

---

### Array Functions

Array functions allow manipulation and querying of arrays within templates, providing a powerful means to handle list data effectively.

**Unique**
Removes duplicate elements from an array.

```plaintext
{{ [1, 2, 2, 3] | unique | join:"," }}
Output: 1,2,3
```

**Join**
Concatenates the elements of an array into a single string, using a given separator.

```plaintext
{{ ["apple", "banana", "cherry"] | join:" - " }}
Output: apple - banana - cherry
```

**First**
Retrieves the first element from an array.

```plaintext
{{ ["first", "second", "third"] | first }}
Output: first
```

**Last**
Obtains the last element of an array.

```plaintext
{{ ["first", "second", "last"] | last }}
Output: last
```

**Random**
Selects a random element from the array. Due to its randomness, the output might vary, so this example uses a fixed scenario for illustration.

```plaintext
{{ [42] | random }}
Output: 42
```

**Reverse**
Reverses the order of elements in an array.

```plaintext
{{ [1, 2, 3] | reverse | join:"," }}
Output: 3,2,1
```

**Shuffle**
Randomly rearranges the elements within the array. The output is unpredictable and varies with each execution.

```plaintext
{{ [1, 2, 3, 4] | shuffle }}
Output: Varies
```

**Size**
Returns the number of elements in the array.

```plaintext
{{ ["one", "two", "three"] | size }}
Output: 3
```

**Max**
Finds the maximum value in a numerical array.

```plaintext
{{ [1, 2, 3, 4, 5] | max }}
Output: 5
```

**Min**
Identifies the minimum value in a numerical array.

```plaintext
{{ [1, 2, 3, 4, 5] | min }}
Output: 1
```

**Sum**
Calculates the sum of all numerical elements in the array.

```plaintext
{{ [1, 2, 3] | sum }}
Output: 6
```

**Average**
Computes the average of numerical values in the array.

```plaintext
{{ [1, 2, 3, 4] | average }}
Output: 2.5
```

**Map**
Extracts values associated with a specified key from each object in an array of maps.

```plaintext
{{ [{"name":"John"}, {"name":"Jane"}] | map:"name" | join:", " }}
Output: John, Jane
```

---

### Date Functions

Date functions provide capabilities for formatting, parsing, and computing differences with dates and times, essential for displaying dates in user-preferred formats or calculating time intervals.

**Date**
Formats a timestamp into a specified format. If no format is provided, a default datetime string is returned.

```plaintext
{{ currentTime | date:"Y-m-d" }}
Output: 2024-03-30
```

**Day**
Extracts and returns the day of the month from a given date.

```plaintext
{{ currentTime | day }}
Output: 30
```

**Month**
Retrieves the month number from a specified date.

```plaintext
{{ currentTime | month }}
Output: 3
```

**MonthFull**
Returns the full name of the month from a specified date.

```plaintext
{{ currentTime | monthFull }}
Output: March
```

**Year**
Extracts and returns the year from a specified date.

```plaintext
{{ currentTime | year }}
Output: 2024
```

**Week**
Returns the ISO week number of a given date.

```plaintext
{{ currentTime | week }}
Output: 13
```

**Weekday**
Determines the day of the week from a specified date, returning its full name.

```plaintext
{{ currentTime | weekday }}
Output: Saturday
```

**TimeAgo**
Generates a human-readable string representing the time difference between the current time and the provided date, often used for displaying how long ago something happened.

```plaintext
{{ pastDate | timeAgo }}
Output: 2 days ago
```

### Number Functions

Number functions are designed to format numeric values, aiding in their presentation and readability within templates.

**Number**
Formats any numeric value according to the specified format string, making it highly adaptable for displaying numbers in various formats.

```plaintext
{{ 1234567.89 | number:"#,###.##" }}
Output: 1,234,567.89
```

**Bytes**
Converts a numeric value into a human-readable format representing bytes, automatically selecting the appropriate unit (KB, MB, GB, etc.) based on the magnitude of the input. This function is particularly useful for displaying file sizes.

```plaintext
{{ 2048 | bytes }}
Output: 2.0 KB
```

---

### Math Functions

Math functions facilitate the execution of mathematical operations on numeric data, enhancing the template's ability to perform calculations and numerical transformations.

**Abs**
Calculates the absolute value of a given number.

```plaintext
{{ -5 | abs }}
Output: 5
```

**AtLeast (at_least)**
Ensures that a number is at least as large as a specified minimum value.

```plaintext
{{ 3 | atLeast:5 }}
Output: 5
```

**AtMost (at_most)**
Ensures that a number does not exceed a specified maximum value.

```plaintext
{{ 10 | atMost:8 }}
Output: 8
```

**Round**
Rounds a number to the nearest whole number or to a specified number of decimal places.

```plaintext
{{ 3.142 | round:2 }}
Output: 3.14
```

**Floor**
Rounds a number down to the nearest whole number.

```plaintext
{{ 3.99 | floor }}
Output: 3
```

**Ceil**
Rounds a number up to the nearest whole number.

```plaintext
{{ 3.01 | ceil }}
Output: 4
```

**Plus**
Adds two numbers together.

```plaintext
{{ 5 | plus:3 }}
Output: 8
```

**Minus**
Subtracts one number from another.

```plaintext
{{ 10 | minus:4 }}
Output: 6
```

**Times**
Multiplies two numbers.

```plaintext
{{ 5 | times:4 }}
Output: 20
```

**Divide**
Divides one number by another. Includes handling for division by zero.

```plaintext
{{ 20 | divide:5 }}
Output: 4
```

**Modulo**
Finds the remainder of division of one number by another.

```plaintext
{{ 10 | modulo:3 }}
Output: 1
```

---

### Map Functions

Map functions provide the capability to interact with and manipulate data stored in maps, enabling more complex data extraction and transformation.

**Extract**
Retrieves a nested value from a map, slice, or array using a dot-separated key path, simplifying access to deeply nested data.

```plaintext
{{ data | extract:"user.profile.age" }}
Output: 30
```