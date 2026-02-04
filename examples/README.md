# Examples

This directory contains example Go structs to demonstrate and test the struct2jsdoc tool with both JSDoc and TypeScript output formats.

## Files

- **user.go** - User and Profile structs with basic types, time.Time, pointers, and slices
- **product.go** - E-commerce product structures with nested structs and arrays
- **api.go** - API response structures with interfaces and the `struct2jsdoc: nogen` marker
- **advanced.go** - Advanced examples with various data types, pointers to time.Time, uint64, maps, etc.

## Features Demonstrated

### Basic Types
- `string`, `int`, `float64`, `bool`
- `uint64`, `uint`, `int8`, etc.

### JSON Tags
- Standard JSON field names
- `omitempty` option for optional fields
- `json:"-"` to exclude fields

### Complex Types
- Pointers (`*Profile`, `*Address`, `*time.Time`)
- Slices (`[]string`, `[]OrderItem`)
- Maps (`map[string]string`, `map[string]interface{}`)
- Nested structs

### External Types
- `time.Time` → converted to `Date` in JSDoc

### Comments
- Inline comments on struct fields
- Documentation for each struct

### Special Directives
- `struct2jsdoc: nogen` - Skip generation for sensitive/internal structs (see `InternalConfig` in api.go)
- `--exclude` flag - Exclude structs matching glob patterns from the command line

## Running the Tool

From the project root directory:

```bash
# Build the tool
go build

# Generate JSDoc types from examples
./struct2jsdoc ./examples ./examples/output.js

# Generate TypeScript interfaces (auto-detected from .ts extension)
./struct2jsdoc ./examples ./examples/output.ts

# TypeScript with custom options
./struct2jsdoc --no-export --ts-type ./examples ./examples/output.ts

# Exclude API-related structs
./struct2jsdoc --exclude "API*" ./examples ./examples/output.js

# Exclude multiple patterns (API structs and anything with Meta in the name)
./struct2jsdoc --exclude "API*" --exclude "*Meta*" ./examples ./examples/output.ts

# View the generated output
cat ./examples/output.js   # JSDoc
cat ./examples/output.ts   # TypeScript
```

## Excluding Structs

There are two ways to exclude structs from generation:

### 1. Using `--exclude` flag (command-line)
Use glob patterns to exclude structs by name:
- `--exclude "Internal*"` - Excludes InternalConfig, InternalState, etc.
- `--exclude "*Test"` - Excludes UserTest, OrderTest, etc.
- `--exclude "*Secret*"` - Excludes SecretConfig, UserSecrets, etc.

Patterns support standard glob syntax:
- `*` matches any sequence of characters
- `?` matches a single character
- Can be repeated multiple times: `--exclude "A*" --exclude "B*"`

### 2. Using `struct2jsdoc: nogen` comment (in code)
Add a comment directive above the struct in your Go code:
```go
// struct2jsdoc: nogen
type InternalConfig struct {
    // This struct will be skipped
}
```

Both methods can be used together. The `--exclude` flag is useful for:
- Temporarily excluding structs without modifying code
- Excluding patterns across many files
- Build-time conditional exclusions

The `nogen` directive is useful for:
- Permanently marking sensitive structs
- Documenting exclusion intent in code

## Output Formats

### JSDoc Output (`output.js`)
Generates JSDoc `@typedef` comments for JavaScript projects:
```javascript
/**
 * @typedef {Object} User
 * @property {number} id - Unique user identifier
 * @property {string} username - User's login name
 * @property {string} [email] - User's email (optional)
 */
```

### TypeScript Output (`output.ts`)
Generates TypeScript interfaces with proper syntax and camelCase conversion:
```typescript
export interface User {
  /** Unique user identifier */
  id: number;
  /** User's login name */
  username: string;
  /** User's email (optional) */
  email?: string;
}
```

**Key Differences:**
- TypeScript uses `interface` keyword (or `type` with `--ts-type`)
- Optional fields use `?` instead of brackets
- Field names use JSON tag when present, otherwise converted to camelCase
- Export keyword added by default (disable with `--no-export`)

**Important:** TypeScript field naming follows this priority:
1. **JSON tag name** (if present) - used as-is: `json:"userId"` → `userId`
2. **Go field name** (if no JSON tag) - converted to camelCase: `UserID` → `userId`

For consistent and predictable naming, always define JSON tags in your Go structs.

## Expected Output

The tool will generate type definitions for all structs except:
- Those marked with `struct2jsdoc: nogen` comment directive
- Those matching `--exclude` patterns

The output will include:
- Type definitions for all non-excluded structs
- Property types converted from Go to JavaScript/JSDoc notation
- Optional properties (those with `omitempty`)
- Field descriptions from comments
