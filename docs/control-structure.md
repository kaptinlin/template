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
- Booleans: `true`, `false`, `True`, `False` (case-insensitive)
- Null: `null`, `Null`, `none`, `None` (case-insensitive)
- Variables: `user.name`, `product.price`

### **2. Operators**
- **Arithmetic**: `+`, `-`, `*`, `/`, `%`
- **Comparison**: `==`, `!=`, `<`, `>`, `<=`, `>=`
- **Logical**: `&&`, `||`, `!` (C-style) or `and`, `or`, `not` (Django-style)
- **Membership**: `in`, `not in` (check if value exists in string/list/map)

### **3. Expression Examples**

#### **C-style Operators (Backward Compatible)**
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

#### **Django-style Operators (Recommended)**
```plaintext
{% if user.age >= 18 and user.verified %}
    Adult and verified
{% endif %}

{% if price > 100 or quantity >= 5 %}
    Eligible for discount
{% endif %}

{% if not user.blocked %}
    User is not blocked
{% endif %}
```

#### **Membership Operators**
```plaintext
{% if "admin" in user.roles %}
    User is an admin
{% endif %}

{% if "error" in message %}
    Message contains error
{% endif %}

{% if user.status not in banned_statuses %}
    User is not banned
{% endif %}
```

#### **Null Checking**
```plaintext
{% if user != null and user.active %}
    Active user
{% endif %}

{% if settings == none %}
    No settings configured
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
parseLogicalOr -> parseLogicalAnd ('||' | 'or') parseLogicalAnd
parseLogicalAnd -> parseNot ('&&' | 'and') parseNot
parseNot -> ('!' | 'not') parseNot | parseIn
parseIn -> parseComparison (('in' | 'not in') parseComparison)?
parseComparison -> parseAdditive (CompOp parseAdditive)（CompOp = '==' | '!=' | '<' | '>' | '<=' | '>='）
parseAdditive -> parseMultiplicative ([+-] parseMultiplicative)
parseMultiplicative -> parseUnary ([*/%] parseUnary)
parseUnary -> ('-') parseUnary | parsePrimary
parsePrimary -> parseBasicPrimary ('|' FilterName)*
parseBasicPrimary -> Number | String | Boolean | Null | Variable | '(' parseExpression ')'
```

**Operator Precedence (from lowest to highest):**
1. `or`, `||` - Logical OR (lowest precedence)
2. `and`, `&&` - Logical AND
3. `not`, `!` - Logical NOT
4. `in`, `not in` - Membership test
5. `==`, `!=`, `<`, `>`, `<=`, `>=` - Comparison
6. `+`, `-` - Addition, Subtraction
7. `*`, `/`, `%` - Multiplication, Division, Modulo
8. Unary `-` - Negation
9. `|` - Filter application (highest precedence)

## **Truthiness Rules**

The template engine follows Django-style truthiness for conditional evaluation:

### **Falsy Values**
- `false`, `False` - Boolean false
- `null`, `Null`, `none`, `None` - Null/nil values
- `0` - Integer zero
- `0.0` - Float zero
- `""` - Empty string
- `[]` - Empty list/slice
- `{}` - Empty map

### **Truthy Values**
- `true`, `True` - Boolean true
- Non-zero numbers
- Non-empty strings
- Non-empty lists/slices
- Non-empty maps
- Struct values

### **Truthiness Examples**
```plaintext
{% if items %}
    items list is not empty
{% endif %}

{% if user.name %}
    user has a name
{% endif %}

{% if count %}
    count is not zero
{% endif %}
```

## **Implementation Details**

1. **Expression Evaluation**
   - Evaluated at runtime with precedence rules.
   - Short-circuit evaluation for `&&`/`and` and `||`/`or`.

2. **Error Handling**
   - Syntax errors during parsing.
   - Runtime errors (e.g., type mismatches) during execution.

3. **Variable Scope**
   - Outer scope variables are accessible within control structures.
   - `for` loops create new scopes for iteration variables.

4. **Type Safety**
   - Operators require matching operand types.
   - No implicit type conversion for arithmetic/comparison.
   - Truthiness conversion for logical operators.

5. **Performance**
   - Expressions are parsed once and cached.
   - Minimal memory allocation during evaluation.

6. **Operator Compatibility**
   - Both C-style (`&&`, `||`, `!`) and Django-style (`and`, `or`, `not`) operators are supported.
   - Can be mixed in the same template for backward compatibility.


## **Advanced Usage**

### **1. Complex Expressions**
```plaintext
{% if (price * quantity > 1000) and (user.level|upper == 'VIP') %}
    Apply VIP discount
{% endif %}
```

### **2. Nested Operations**
```plaintext
{% if not (user.age < 18 or user.restricted) and user.verified %}
    Access granted
{% endif %}
```

### **3. Filter Chaining**
```plaintext
{% if product.name|trim|length > 0 %}
    Product has valid name
{% endif %}
```

### **4. Membership Testing**
```plaintext
{% if request.method in ['POST', 'PUT', 'PATCH'] %}
    Modifying request
{% endif %}

{% if user.email not in blacklist %}
    Valid email
{% endif %}
```

### **5. Combining Different Operators**
```plaintext
{% if user != null and user.role in allowed_roles and not user.banned %}
    Access granted
{% endif %}

{% if (score >= 90 or extra_credit) and "math" in subjects %}
    Award distinction
{% endif %}
```

## **Best Practices**

1. **Proper Closing**
   - Always close control structures with `{% endfor %}` or `{% endif %}`.

2. **Indentation**
   - Use consistent indentation for readability in nested structures.

3. **Operator Style**
   - Prefer Django-style operators (`and`, `or`, `not`) for better readability.
   - C-style operators (`&&`, `||`, `!`) are supported for backward compatibility.
   - Be consistent within the same template.

4. **Negation**
   - Use `not` (Django-style) or `!` (C-style) for logical negation.
   - Prefer `not` for better readability: `{% if not user.blocked %}`

5. **Null Checking**
   - Use `!= null` or `== null` for explicit null checks.
   - Use truthiness for general existence checks: `{% if user %}`

6. **Membership Testing**
   - Use `in` for checking if a value exists in a collection.
   - Use `not in` for negative membership tests.

7. **Variable Access**
   - Outer scope variables are accessible within loops and conditions.

8. **Missing Data**
   - If a variable or property is missing, the template renders an empty string.

9. **Operator Precedence**
   - Use parentheses to make complex expressions more readable.
   - Remember: `or` < `and` < `not` < `in` < comparisons
