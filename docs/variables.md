## Variables

Variables are placeholders within templates that are dynamically filled with data when the template is rendered. They are defined within double curly braces, like `{{ variableName }}`. Variable names can include alphanumeric characters and underscores. A dot (`.`) is used to access properties of variables, enabling you to retrieve nested data.

### Accessing Properties

To access a property of a variable, use the dot notation:

```
{{ object.property }}
```

This notation can be extended to access nested objects:

```
{{ object.nestedObject.property }}
```

If a variable or its property does not exist, the template will render an empty string for that variable.

### Example 1: Basic Variable Rendering

Consider a scenario where you want to personalize a greeting in an email template. Your data source provides a `name` variable.

**Template:**
```
Hello, {{ name }}!
```

**Context Data:**
```json
{
  "name": "Alice"
}
```

**Rendered Output:**
```
Hello, Alice!
```

This example shows how a simple variable `{{ name }}` is replaced with the value `"Alice"` from the context data.

### Example 2: Accessing Object Properties

In a blog application, you might have a `post` object that contains various properties such as `title` and `author`. Here's how you can access these properties:

**Template:**
```
Title: {{ post.title }}
Author: {{ post.author }}
```

**Context Data:**
```json
{
  "post": {
    "title": "Introduction to Templates",
    "author": "Bob"
  }
}
```

**Rendered Output:**
```
Title: Introduction to Templates
Author: Bob
```

This demonstrates accessing properties of an object within a template.

### Example 3: Nested Object Access

For a user profile page, consider displaying a user's address. The address is nested within the `user` object.

**Template:**
```
Address: {{ user.address.street }}, {{ user.address.city }}
```

**Context Data:**
```json
{
  "user": {
    "address": {
      "street": "123 Maple Street",
      "city": "Springfield"
    }
  }
}
```

**Rendered Output:**
```
Address: 123 Maple Street, Springfield
```

This example shows how to navigate through nested objects to retrieve specific pieces of data.

### Example 4: Handling Missing Data

Templates gracefully handle missing data by rendering an empty string when a variable is not found.

**Template:**
```
Welcome, {{ name }}!
```

**Context Data:**
```json
{}
```

**Rendered Output:**
```
Welcome, !
```

Since `name` is not provided in the context data, the output defaults to an empty space after "Welcome,".

### Example 5: Using Lists

If you're listing items from a collection, such as product names, you can also use variables to iterate over lists (though the iteration would be managed by the template's logic outside the scope of simple variable replacement).

**Template:**
```
Products:
- {{ products.0 }}
- {{ products.1 }}
```

**Context Data:**
```json
{
  "products": ["Coffee Maker", "Toaster"]
}
```

**Rendered Output:**
```
Products:
- Coffee Maker
- Toaster
```

This shows how to access elements in a list by their index.
