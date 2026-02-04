package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

const version = "1.0.0"

// OutputFormat represents the output format type
type OutputFormat int

const (
	FormatJSDoc OutputFormat = iota
	FormatTypeScript
)

// excludePatterns is a custom type to support multiple --exclude flags
type excludePatterns []string

func (e *excludePatterns) String() string {
	return strings.Join(*e, ",")
}

func (e *excludePatterns) Set(value string) error {
	*e = append(*e, value)
	return nil
}

// toCamelCase converts PascalCase to camelCase
// Examples: UserID -> userId, FirstName -> firstName, HTTPServer -> httpServer
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	result := make([]rune, 0, len(runes))

	for i := 0; i < len(runes); i++ {
		curr := runes[i]

		if i == 0 {
			// First character is always lowercase
			result = append(result, unicode.ToLower(curr))
		} else if unicode.IsUpper(curr) {
			// Check if this starts a new word or is part of an acronym
			isLast := i == len(runes)-1
			nextIsLower := !isLast && unicode.IsLower(runes[i+1])
			prevIsUpper := i > 0 && unicode.IsUpper(runes[i-1])

			if nextIsLower && prevIsUpper {
				// This is the start of a new word after an acronym
				// e.g., "ID" in "UserID" -> "Id", or "K" in "APIKey" -> "Key"
				result = append(result, curr)
			} else if nextIsLower && !prevIsUpper {
				// This is the start of a new word
				// e.g., "N" in "FirstName" -> keep uppercase
				result = append(result, curr)
			} else {
				// Part of an acronym or last character of an acronym
				// e.g., "I" in "ID", "P", "I" in "API"
				result = append(result, unicode.ToLower(curr))
			}
		} else {
			// Lowercase or other character, keep as-is
			result = append(result, curr)
		}
	}

	return string(result)
}

var goToJSMappings = map[string]string{
	"string":    "string",
	"int":       "number",
	"int8":      "number",
	"int16":     "number",
	"int32":     "number",
	"int64":     "number",
	"uint":      "number",
	"uint8":     "number",
	"uint16":    "number",
	"uint32":    "number",
	"uint64":    "number",
	"float32":   "number",
	"float64":   "number",
	"bool":      "boolean",
	"byte":      "number",
	"rune":      "number",
	"any":       "*",
	"interface": "*",
}

var goToTSMappings = map[string]string{
	"string":    "string",
	"int":       "number",
	"int8":      "number",
	"int16":     "number",
	"int32":     "number",
	"int64":     "number",
	"uint":      "number",
	"uint8":     "number",
	"uint16":    "number",
	"uint32":    "number",
	"uint64":    "number",
	"float32":   "number",
	"float64":   "number",
	"bool":      "boolean",
	"byte":      "number",
	"rune":      "number",
	"any":       "any",
	"interface": "any",
}

// knownStructs holds all struct names discovered in the first pass
var knownStructs = make(map[string]bool)

// goTypeToJSType converts a Go AST type expression to a JSDoc type string
func goTypeToJSType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple identifier: string, int, CustomType, etc.
		if jsType, ok := goToJSMappings[t.Name]; ok {
			return jsType
		}
		// Check if it's a known struct from our codebase
		if knownStructs[t.Name] {
			return t.Name
		}
		// Handle common stdlib types
		if t.Name == "Time" {
			return "Date"
		}
		return t.Name

	case *ast.StarExpr:
		// Pointer type: *string, *User -> just use the underlying type
		return goTypeToJSType(t.X)

	case *ast.ArrayType:
		// Array or slice: []string, [5]int
		elemType := goTypeToJSType(t.Elt)
		return elemType + "[]"

	case *ast.MapType:
		// Map: map[string]int -> Object.<string, number>
		keyType := goTypeToJSType(t.Key)
		valType := goTypeToJSType(t.Value)
		return fmt.Sprintf("Object.<%s, %s>", keyType, valType)

	case *ast.SelectorExpr:
		// Qualified identifier: time.Time, uuid.UUID
		if ident, ok := t.X.(*ast.Ident); ok {
			fullName := ident.Name + "." + t.Sel.Name
			// Handle common external types
			switch fullName {
			case "time.Time":
				return "Date"
			case "uuid.UUID":
				return "string"
			case "json.RawMessage":
				return "*"
			default:
				return t.Sel.Name
			}
		}
		return "*"

	case *ast.InterfaceType:
		// interface{} -> *
		return "*"

	case *ast.StructType:
		// Anonymous struct -> Object
		return "Object"

	default:
		return "*"
	}
}

// goTypeToTSType converts a Go AST type expression to a TypeScript type string
func goTypeToTSType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple identifier: string, int, CustomType, etc.
		if tsType, ok := goToTSMappings[t.Name]; ok {
			return tsType
		}
		// Check if it's a known struct from our codebase
		if knownStructs[t.Name] {
			return t.Name
		}
		// Handle common stdlib types
		if t.Name == "Time" {
			return "Date"
		}
		return t.Name

	case *ast.StarExpr:
		// Pointer type: *string, *User -> just use the underlying type (TS doesn't have pointers)
		return goTypeToTSType(t.X)

	case *ast.ArrayType:
		// Array or slice: []string, [5]int
		elemType := goTypeToTSType(t.Elt)
		return elemType + "[]"

	case *ast.MapType:
		// Map: map[string]int -> Record<string, number>
		keyType := goTypeToTSType(t.Key)
		valType := goTypeToTSType(t.Value)
		return fmt.Sprintf("Record<%s, %s>", keyType, valType)

	case *ast.SelectorExpr:
		// Qualified identifier: time.Time, uuid.UUID
		if ident, ok := t.X.(*ast.Ident); ok {
			fullName := ident.Name + "." + t.Sel.Name
			// Handle common external types
			switch fullName {
			case "time.Time":
				return "Date"
			case "uuid.UUID":
				return "string"
			case "json.RawMessage":
				return "any"
			default:
				return t.Sel.Name
			}
		}
		return "any"

	case *ast.InterfaceType:
		// interface{} -> any
		return "any"

	case *ast.StructType:
		// Anonymous struct -> Record<string, any>
		return "Record<string, any>"

	default:
		return "any"
	}
}

// parseJSONTag extracts the field name and whether omitempty is set from a struct tag
func parseJSONTag(tag *ast.BasicLit) (name string, omitempty bool) {
	if tag == nil {
		return "", false
	}

	tagValue := tag.Value
	// Remove backticks
	tagValue = strings.Trim(tagValue, "`")

	// Find json:"..."
	jsonIdx := strings.Index(tagValue, `json:"`)
	if jsonIdx == -1 {
		return "", false
	}

	// Extract the value after json:"
	start := jsonIdx + 6
	end := strings.Index(tagValue[start:], `"`)
	if end == -1 {
		return "", false
	}

	jsonValue := tagValue[start : start+end]

	// Split by comma to separate name from options
	parts := strings.Split(jsonValue, ",")
	name = parts[0]

	// Check for omitempty in remaining parts
	omitempty = slices.Contains(parts[1:], "omitempty")

	// Handle json:"-" (skip field)
	if name == "-" {
		return "-", false
	}

	return name, omitempty
}

// extractFieldComment gets the comment associated with a field
func extractFieldComment(field *ast.Field) string {
	var comment string

	// Check inline comment first (more common for struct fields)
	if field.Comment != nil && len(field.Comment.List) > 0 {
		comment = field.Comment.List[0].Text
	} else if field.Doc != nil && len(field.Doc.List) > 0 {
		// Fall back to doc comment above the field
		comment = field.Doc.List[0].Text
	}

	if comment == "" {
		return ""
	}

	// Clean up the comment
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimPrefix(comment, "/*")
	comment = strings.TrimSuffix(comment, "*/")
	comment = strings.TrimSpace(comment)

	return comment
}

// generateJSDocType converts a Go struct's AST fields into a JSDoc typedef string.
func generateJSDocType(structName string, fields []*ast.Field) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("/**\n * @typedef {Object} %s\n", structName))

	for _, field := range fields {
		// Skip embedded or anonymous fields
		if len(field.Names) == 0 {
			continue
		}
		fieldName := field.Names[0].Name

		// Convert the Go type to JSDoc type
		jsType := goTypeToJSType(field.Type)

		// Extract JSON tag info
		jsonName, omitempty := parseJSONTag(field.Tag)

		// Skip fields with json:"-"
		if jsonName == "-" {
			continue
		}

		// Use JSON name if present, otherwise use Go field name
		propName := jsonName
		if propName == "" {
			propName = fieldName
		}

		// Make optional if omitempty
		if omitempty {
			propName = "[" + propName + "]"
		}

		// Extract field comment for description
		comment := extractFieldComment(field)

		// Build the property line
		if comment != "" {
			buf.WriteString(fmt.Sprintf(" * @property {%s} %s - %s\n", jsType, propName, comment))
		} else {
			buf.WriteString(fmt.Sprintf(" * @property {%s} %s\n", jsType, propName))
		}
	}

	buf.WriteString(" */\n")
	return buf.String()
}

// generateTypeScriptInterface converts a Go struct's AST fields into a TypeScript interface.
func generateTypeScriptInterface(structName string, fields []*ast.Field, useTypeKeyword bool, addExport bool) string {
	var buf bytes.Buffer

	// Add export keyword if requested
	if addExport {
		buf.WriteString("export ")
	}

	// Use 'type' or 'interface' keyword
	if useTypeKeyword {
		buf.WriteString(fmt.Sprintf("type %s = {\n", structName))
	} else {
		buf.WriteString(fmt.Sprintf("interface %s {\n", structName))
	}

	for _, field := range fields {
		// Skip embedded or anonymous fields
		if len(field.Names) == 0 {
			continue
		}
		fieldName := field.Names[0].Name

		// Convert the Go type to TypeScript type
		tsType := goTypeToTSType(field.Type)

		// Extract JSON tag info
		jsonName, omitempty := parseJSONTag(field.Tag)

		// Skip fields with json:"-"
		if jsonName == "-" {
			continue
		}

		// Use JSON name if present, otherwise convert Go field name to camelCase
		propName := jsonName
		if propName == "" {
			propName = toCamelCase(fieldName)
		}

		// Extract field comment for description
		comment := extractFieldComment(field)

		// Add comment if present
		if comment != "" {
			buf.WriteString(fmt.Sprintf("  /** %s */\n", comment))
		}

		// Build the property line with optional marker
		optionalMarker := ""
		if omitempty {
			optionalMarker = "?"
		}

		buf.WriteString(fmt.Sprintf("  %s%s: %s;\n", propName, optionalMarker, tsType))
	}

	if useTypeKeyword {
		buf.WriteString("}\n")
	} else {
		buf.WriteString("}\n")
	}

	return buf.String()
}

// matchesExcludePattern checks if a struct name matches any exclude pattern
func matchesExcludePattern(structName string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, structName)
		if err != nil {
			// Invalid pattern, skip it
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// collectStructNames does a first pass to gather all struct names
func collectStructNames(pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						knownStructs[typeSpec.Name.Name] = true
					}
				}
			}
		}
	}
}

func main() {
	// Define flags
	showVersion := flag.Bool("V", false, "show version information")
	flag.BoolVar(showVersion, "version", false, "show version information")

	var excludes excludePatterns
	flag.Var(&excludes, "exclude", "exclude structs matching pattern (can be repeated, supports glob patterns like 'Internal*')")

	outputType := flag.String("t", "", "output type: jsdoc, js, typescript, ts (auto-detected from file extension if not specified)")
	flag.StringVar(outputType, "type", "", "output type: jsdoc, js, typescript, ts (auto-detected from file extension if not specified)")

	noExport := flag.Bool("no-export", false, "don't add 'export' keyword to TypeScript definitions")
	useTypeKeyword := flag.Bool("ts-type", false, "use 'type' keyword instead of 'interface' for TypeScript")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "struct2jsdoc - Convert Go structs to JSDoc or TypeScript type definitions\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options] <go_structs_dir> <output_file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  go_structs_dir   Directory containing Go struct definitions\n")
		fmt.Fprintf(os.Stderr, "  output_file      Output file (.js for JSDoc, .ts for TypeScript)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -h, --help       Show this help message\n")
		fmt.Fprintf(os.Stderr, "  -V, --version    Show version information\n")
		fmt.Fprintf(os.Stderr, "  -t, --type       Output type: jsdoc|js|typescript|ts (auto-detected if not specified)\n")
		fmt.Fprintf(os.Stderr, "  --exclude        Exclude structs matching glob pattern (can be repeated)\n")
		fmt.Fprintf(os.Stderr, "                   Example: --exclude 'Internal*' --exclude '*Config'\n")
		fmt.Fprintf(os.Stderr, "  --no-export      Don't add 'export' keyword (TypeScript only)\n")
		fmt.Fprintf(os.Stderr, "  --ts-type        Use 'type' instead of 'interface' (TypeScript only)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s ./models ./output/types.js              # JSDoc (auto-detected)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s ./models ./output/types.ts              # TypeScript (auto-detected)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t typescript ./models ./output/types.d.ts\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --exclude 'Internal*' ./models ./types.ts\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --no-export --ts-type ./models ./types.ts\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("struct2jsdoc version %s\n", version)
		os.Exit(0)
	}

	// Check for required arguments
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	goDir := flag.Arg(0)
	outputFile := flag.Arg(1)

	// Determine output format
	var format OutputFormat
	if *outputType != "" {
		// Explicit type specified
		switch strings.ToLower(*outputType) {
		case "jsdoc", "js":
			format = FormatJSDoc
		case "typescript", "ts":
			format = FormatTypeScript
		default:
			fmt.Fprintf(os.Stderr, "Invalid output type: %s. Valid options: jsdoc, js, typescript, ts\n", *outputType)
			os.Exit(1)
		}
	} else {
		// Auto-detect from file extension
		ext := filepath.Ext(outputFile)
		switch ext {
		case ".ts":
			format = FormatTypeScript
		case ".js":
			format = FormatJSDoc
		default:
			// Default to JSDoc for unknown extensions
			format = FormatJSDoc
		}
	}

	// Load packages using go/packages (replaces deprecated parser.ParseDir)
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles,
		Dir:  goDir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load packages from %q: %v\n", goDir, err)
		os.Exit(1)
	}

	// Check for package loading errors
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "Package error: %v\n", e)
			}
			os.Exit(1)
		}
	}

	// First pass: collect all struct names for cross-referencing
	collectStructNames(pkgs)

	var outputBuf bytes.Buffer

	// Write header comment
	outputBuf.WriteString("// Generated by struct2jsdoc - https://github.com/uradical/struct2jsdoc\n\n")

	// Second pass: generate output (JSDoc or TypeScript)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				// Check the doc comments for the marker
				skipStruct := false
				if genDecl.Doc != nil {
					for _, comment := range genDecl.Doc.List {
						if strings.Contains(comment.Text, "struct2jsdoc: nogen") {
							skipStruct = true
							break
						}
					}
				}

				if skipStruct {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					structName := typeSpec.Name.Name

					// Check if struct matches any exclude pattern
					if matchesExcludePattern(structName, excludes) {
						continue
					}

					// Generate output based on format
					var output string
					if format == FormatTypeScript {
						output = generateTypeScriptInterface(structName, structType.Fields.List, *useTypeKeyword, !*noExport)
					} else {
						output = generateJSDocType(structName, structType.Fields.List)
					}
					outputBuf.WriteString(output + "\n")
				}
			}
		}
	}

	err = os.WriteFile(outputFile, outputBuf.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file %q: %v\n", outputFile, err)
		os.Exit(1)
	}

	// Success message
	formatName := "JSDoc"
	if format == FormatTypeScript {
		formatName = "TypeScript"
	}
	fmt.Printf("%s types successfully written to %s\n", formatName, outputFile)
}
