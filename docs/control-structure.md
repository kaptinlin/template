# Control Structures

Control structures enable dynamic content generation in templates by adding logical flow control, such as loops (`for`) and conditionals (`if`). These structures can be nested to create complex templates.

## **Loop Structure (`for`)**

The `for` loop iterates over arrays, maps, or other iterable objects.

### **Basic Syntax**
```plaintext
{% for key in iterable %}
    {{ key }}
{% endfor %}
```

### **Example 1: Iterating Over a Map**
**Template:**
```plaintext
{% for key in simple.strmap %}
    Key: {{ key }}
{% endfor %}
```

**Context Data:**
```json
{
    "simple": {
        "strmap": {
            "key1": "value1",
            "key2": "value2"
        }
    }
}
```

**Output:**
```plaintext
Key: key1
Key: key2
```

### **Example 2: Iterating Over a List**
**Template:**
```plaintext
{% for item in products %}
    Product: {{ item }}
{% endfor %}
```

**Context Data:**
```json
{
    "products": ["Coffee Maker", "Toaster"]
}
```

**Output:**
```plaintext
Product: Coffee Maker
Product: Toaster
```


## **Conditional Structure (`if`)**

The `if` statement conditionally renders content based on data values.

### **Basic Syntax**
```plaintext
{% if condition %}
    Content to render if condition is true
{% endif %}
```

### **Example 1: Basic `if` Condition**
**Template:**
```plaintext
{% if simple.float %}
    Float value is: {{ simple.float }}
{% endif %}
```

**Context Data:**
```json
{
    "simple": {
        "float": 3.14
    }
}
```

**Output:**
```plaintext
Float value is: 3.14
```

### **Example 2: `if-else` Condition**
**Template:**
```plaintext
{% if !simple %}
    false
{% else %}
    !simple
{% endif %}
```

**Context Data:**
```json
{
    "simple": null
}
```

**Output:**
```plaintext
!simple
```

## **Nested Control Structures**

Control structures can be nested for complex logical flows.

### **Example 1: `for` Loop with Nested `if` Condition**
**Template:**
```plaintext
{% for key in simple.strmap %}
    {% if simple.float %}
        {{ key }}: {{ simple.float }}
    {% endif %}
{% endfor %}
```

**Context Data:**
```json
{
    "simple": {
        "strmap": {
            "key1": "value1",
            "key2": "value2"
        },
        "float": 3.14
    }
}
```

**Output:**
```plaintext
key1: 3.14
key2: 3.14
```

### **Example 2: `if` Condition with Nested `for` Loop**
**Template:**
```plaintext
{% if simple.float %}
    {% for key in simple.strmap %}
        {{ key }}
    {% endfor %}
{% endif %}
```

**Context Data:**
```json
{
    "simple": {
        "strmap": {
            "key1": "value1",
            "key2": "value2"
        },
        "float": 3.14
    }
}
```

**Output:**
```plaintext
key1
key2
```

## **Expression Syntax**

Control structures support rich expression syntax, including:

### **1. Basic Data Types**
- Numbers: `123`, `3.14`
- Strings: `"hello"`, `'world'`
- Booleans: `true`, `false`
- Variables: `user.name`, `product.price`

### **2. Operators**
- **Arithmetic**: `+`, `-`, `*`, `/`, `%`
- **Comparison**: `==`, `!=`, `<`, `>`, `<=`, `>=`
- **Logical**: `&&`, `||`, `!`

### **3. Expression Examples**
```plaintext
{% if user.age >= 18 && user.verified %}
    Adult and verified
{% endif %}

{% if price > 100 || quantity >= 5 %}
    Eligible for discount
{% endif %}

{% if !(user.blocked) %}
    User is not blocked
{% endif %}
```

### **4. Filters**
Filters transform values within expressions:
```plaintext
{% if user.name|length > 0 %}
    Username is not empty
{% endif %}

```

## **Lexical and Syntax Analysis**

### **Lexical Analysis**
The template engine breaks expressions into tokens:
- **Identifiers**: Variable names, property paths
- **Literals**: Numbers, strings, booleans
- **Operators**: Arithmetic, comparison, logical
- **Others**: Parentheses, pipes, filters

### **Syntax Analysis**
The parser uses recursive descent to process expressions, following precedence rules:
```plaintext
Parse -> parseExpression
parseExpression -> parseLogicalOr
parseLogicalOr -> parseLogicalAnd ('||' parseLogicalAnd)
parseLogicalAnd -> parseComparison ('&&' parseComparison)
parseComparison -> parseAdditive (CompOp parseAdditive)（CompOp = '==' | '!=' | '<' | '>' | '<=' | '>='）
parseAdditive -> parseMultiplicative ([+-] parseMultiplicative)
parseMultiplicative -> parseUnary ([/%] parseUnary)
parseUnary -> ('!') parseUnary | parsePrimary
parsePrimary -> parseBasicPrimary ('|' FilterName)
parseBasicPrimary -> Number | String | Boolean | Variable | '(' parseExpression ')'
```

## **Implementation Details**

1. **Expression Evaluation**
   - Evaluated at runtime with precedence rules.
   - Short-circuit evaluation for `&&` and `||`.

2. **Error Handling**
   - Syntax errors during parsing.
   - Runtime errors (e.g., type mismatches) during execution.

3. **Variable Scope**
   - Outer scope variables are accessible within control structures.
   - `for` loops create new scopes for iteration variables.

4. **Type Safety**
   - Operators require matching operand types.
   - No implicit type conversion.

5. **Performance**
   - Expressions are parsed once and cached.
   - Minimal memory allocation during evaluation.


## **Advanced Usage**

### **1. Complex Expressions**
```plaintext
{% if (price * quantity > 1000) && (user.level|upper == 'VIP') %}
    Apply VIP discount
{% endif %}
```

### **2. Nested Operations**
```plaintext
{% if !(user.age < 18 || user.restricted) && user.verified %}
    Access granted
{% endif %}
```

### **3. Filter Chaining**
```plaintext
{% if product.name|trim|length > 0 %}
    Product has valid name
{% endif %}
```

## **Best Practices**

1. **Proper Closing**
   - Always close control structures with `{% endfor %}` or `{% endif %}`.

2. **Indentation**
   - Use consistent indentation for readability in nested structures.

3. **Negation**
   - Use `!` for logical negation.

4. **Variable Access**
   - Outer scope variables are accessible within loops and conditions.

5. **Missing Data**
   - If a variable or property is missing, the template renders an empty string.
