# struct2jsdoc

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A CLI tool that converts Go structs to [JSDoc](https://jsdoc.app/) type definitions or TypeScript interfaces.

## Installation

```shell
go install uradical.io/go/struct2jsdoc@latest
```

## Usage

```shell
struct2jsdoc [options] <go_structs_dir> <output_file>
```

The output format is automatically detected from the file extension (`.js` for JSDoc, `.ts` for TypeScript), or can be explicitly specified with the `-t` flag.

### Options

- `-h`, `--help` - Show help message
- `-V`, `--version` - Show version information
- `-t`, `--type <format>` - Output type: `jsdoc`, `js`, `typescript`, `ts` (auto-detected if not specified)
- `--exclude <pattern>` - Exclude structs matching glob pattern (can be repeated)
- `--no-export` - Don't add `export` keyword (TypeScript only)
- `--ts-type` - Use `type` instead of `interface` (TypeScript only)

### Examples

```shell
# Generate JSDoc (auto-detected from .js extension)
struct2jsdoc ./models ./output/types.js

# Generate TypeScript (auto-detected from .ts extension)
struct2jsdoc ./models ./output/types.ts

# Explicit type specification
struct2jsdoc -t typescript ./models ./output/types.d.ts

# TypeScript with custom options
struct2jsdoc --no-export --ts-type ./models ./types.ts

# Exclude structs matching patterns (glob-style)
struct2jsdoc --exclude "Internal*" --exclude "*Config" ./models ./output.ts

# Combine exclude with TypeScript
struct2jsdoc --exclude "*Test" --exclude "*Secret*" -t ts ./models ./output.ts

# Try it with the included examples
struct2jsdoc ./examples ./examples/output.js    # JSDoc
struct2jsdoc ./examples ./examples/output.ts    # TypeScript

# Show help
struct2jsdoc -h

# Show version
struct2jsdoc -V
```

## Features

### Dual Output Formats
- **JSDoc** - Generate JSDoc `@typedef` comments for JavaScript projects
- **TypeScript** - Generate TypeScript interfaces or types with proper syntax

### Smart Type Conversion
- Converts Go types to JavaScript/TypeScript equivalents
- Maps, slices, and complex types handled automatically
- Preserves field comments as documentation

### Flexible Configuration
- Auto-detect output format from file extension
- Exclude patterns to skip sensitive structs
- TypeScript-specific options (`--no-export`, `--ts-type`)
- CamelCase conversion for TypeScript field names

**Note on TypeScript Field Naming:** Field names are taken from JSON tags when present. Fields without JSON tags are automatically converted to camelCase. For best results and consistent naming, always use JSON tags in your Go structs.

### Example Output

**Go Input:**
```go
type User struct {
    ID       int       `json:"id"`
    Username string    `json:"username"`
    Email    string    `json:"email,omitempty"`  // User's email
}
```

**JSDoc Output:**
```javascript
/**
 * @typedef {Object} User
 * @property {number} id
 * @property {string} username
 * @property {string} [email] - User's email
 */
```

**TypeScript Output:**
```typescript
export interface User {
  id: number;
  username: string;
  /** User's email */
  email?: string;
}
```

## Examples Directory

The `examples/` directory contains sample Go structs demonstrating various features:
- Basic types and JSON tags
- Nested structs and arrays
- Maps and interfaces
- Optional fields with `omitempty`
- Field comments and documentation
- The `struct2jsdoc: nogen` directive to skip struct generation

See [examples/README.md](examples/README.md) for more details.
