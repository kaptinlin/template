# Simple Go Template Engine

## Overview

This Go Template Engine library introduces dynamic templating for Go applications, enabling variable interpolation and data manipulation with filters. Drawing inspiration from the syntax of Liquid and Django, it simplifies the generation of dynamic content.

## Getting Started

### Installation

Ensure you have Go set up on your system. To add the Go Template Engine to your project, run:

```sh
go get github.com/kaptinlin/template
```

This command downloads the library and prepares it for use in your project.

### Basic Usage

#### Parsing and Executing a Template

Create and parse a template, then execute it with a context:

```go
package main

import (
	"fmt"
	"github.com/kaptinlin/template"
)

func main() {
    // Define your template
    source := "Hello, {{ name }}!"
    // Parse the template
    tpl, err := template.Parse(source)
    if err != nil {
        panic(err)
    }
    
    // Create a context and add variables
    context := template.NewContext()
    context.Set("name", "World")
    
    // Execute the template
    output, err := template.Execute(tpl, context)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(output) // Output: Hello, World!
}
```

#### Quick Parsing and Execution with Render

Directly parse and execute a template in one step:

```go
package main

import (
	"fmt"
	"github.com/kaptinlin/template"
)

func main() {
    // Define your template and context
    source := "Goodbye, {{ name }}!"
    context := template.NewContext()
    context.Set("name", "Mars")
    
    // Render the template
    output, err := template.Render(source, context)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(output) // Output: Goodbye, Mars!
}
```

#### Ignoring Errors with MustExecute
Execute a template and ignore any errors, useful for templates guaranteed not to fail:

```go
package main

import (
	"fmt"
	"github.com/kaptinlin/template"
)

func main() {
    // Define your template
    source := "Welcome, {{ name }}!"
    // Parse the template
    tpl, err := template.Parse(source)
    if err != nil {
        panic(err)
    }
    
    // Create a context and add variables
    context := template.NewContext()
    context.Set("name", "Universe")
    
    // MustExecute the template, ignoring errors
    output := template.MustExecute(tpl, context)
    
    fmt.Println(output) // Output: Welcome, Universe!
}
```

## Syntax and Features

### Variables

Enclose variables in `{{ }}` to embed dynamic content:

```go
{{ userName }}
```

For extended syntax, refer to the [documentation](docs/variables.md).

### Filters

Use the pipe `|` to apply filters to variables:

```go
Hello, {{ name|capitalize }}!
```

Detailed usage can be found in the [documentation](docs/filters.md).

### For Loops

Use `{% for item in collection %}` and `{% endfor %}` to iterate over arrays, maps:

```go
{% for item in items %}
    {{ item }}
{% endfor %}
```

### If Conditionals

Use `{% if condition %}`, `{% else %}` and `{% endif %}` to conditionally render content based on expressions:

```go
{% if user.age >= 18 %}
    Adult
{% else %}
    Minor
{% endif %}
```

Control structures can be nested and support complex expressions. For more details, see [Control Structures Documentation](docs/control-structure.md).


## Custom Filters

Easily extend functionality by adding custom filters. For example, a filter to capitalize a string:

```go
package main

import (
	"github.com/kaptinlin/template"
	"strings"
)

func capitalize(input interface{}, args ...string) (interface{}, error) {
	s, ok := input.(string)
	if !ok {
		return input, nil
	}
	return strings.Title(s), nil
}

func init() {
	template.RegisterFilter("capitalize", capitalize)
}
```

Use the custom filter like so:

```go
{{ "john doe"|capitalize }}
```

## Context Management

Contexts pass variables to templates. Here’s how to create and use one:

```go
context := template.NewContext()
context.Set("key", "value")
```

## How to Contribute

Contributions to the `template` package are welcome. If you'd like to contribute, please follow the [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.