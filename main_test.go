package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// Test toCamelCase conversion
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserID", "userId"},
		{"FirstName", "firstName"},
		{"LastName", "lastName"},
		{"IsActive", "isActive"},
		{"APIKey", "apiKey"},
		{"HTTPServer", "httpServer"},
		{"URL", "url"},
		{"ID", "id"},
		{"A", "a"},
		{"", ""},
		{"lowercase", "lowercase"},
		{"UPPERCASE", "uppercase"},
		{"HTMLParser", "htmlParser"},
		{"XMLDocument", "xmlDocument"},
		{"JSONData", "jsonData"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test parseJSONTag
func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		name          string
		tagValue      string
		expectedName  string
		expectedOmit  bool
	}{
		{
			name:          "simple tag",
			tagValue:      "`json:\"username\"`",
			expectedName:  "username",
			expectedOmit:  false,
		},
		{
			name:          "tag with omitempty",
			tagValue:      "`json:\"email,omitempty\"`",
			expectedName:  "email",
			expectedOmit:  true,
		},
		{
			name:          "tag with dash (skip)",
			tagValue:      "`json:\"-\"`",
			expectedName:  "-",
			expectedOmit:  false,
		},
		{
			name:          "no json tag",
			tagValue:      "`db:\"user_id\"`",
			expectedName:  "",
			expectedOmit:  false,
		},
		{
			name:          "empty tag",
			tagValue:      "",
			expectedName:  "",
			expectedOmit:  false,
		},
		{
			name:          "tag with multiple options",
			tagValue:      "`json:\"data,omitempty,string\"`",
			expectedName:  "data",
			expectedOmit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tag *ast.BasicLit
			if tt.tagValue != "" {
				tag = &ast.BasicLit{Value: tt.tagValue}
			}

			name, omit := parseJSONTag(tag)
			if name != tt.expectedName {
				t.Errorf("parseJSONTag() name = %q, want %q", name, tt.expectedName)
			}
			if omit != tt.expectedOmit {
				t.Errorf("parseJSONTag() omitempty = %v, want %v", omit, tt.expectedOmit)
			}
		})
	}
}

// Test matchesExcludePattern
func TestMatchesExcludePattern(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		patterns   []string
		expected   bool
	}{
		{
			name:       "exact match",
			structName: "User",
			patterns:   []string{"User"},
			expected:   true,
		},
		{
			name:       "wildcard prefix",
			structName: "InternalConfig",
			patterns:   []string{"Internal*"},
			expected:   true,
		},
		{
			name:       "wildcard suffix",
			structName: "UserTest",
			patterns:   []string{"*Test"},
			expected:   true,
		},
		{
			name:       "wildcard both sides",
			structName: "UserSecret",
			patterns:   []string{"*Secret*"},
			expected:   true,
		},
		{
			name:       "no match",
			structName: "User",
			patterns:   []string{"Admin*", "*Test"},
			expected:   false,
		},
		{
			name:       "multiple patterns, one matches",
			structName: "APIResponse",
			patterns:   []string{"User", "API*", "Internal*"},
			expected:   true,
		},
		{
			name:       "empty patterns",
			structName: "User",
			patterns:   []string{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesExcludePattern(tt.structName, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesExcludePattern(%q, %v) = %v, want %v",
					tt.structName, tt.patterns, result, tt.expected)
			}
		})
	}
}

// Test goTypeToJSType with AST nodes
func TestGoTypeToJSType(t *testing.T) {
	tests := []struct {
		name     string
		goCode   string
		expected string
	}{
		{
			name:     "string type",
			goCode:   "package test\ntype T struct { F string }",
			expected: "string",
		},
		{
			name:     "int type",
			goCode:   "package test\ntype T struct { F int }",
			expected: "number",
		},
		{
			name:     "bool type",
			goCode:   "package test\ntype T struct { F bool }",
			expected: "boolean",
		},
		{
			name:     "slice type",
			goCode:   "package test\ntype T struct { F []string }",
			expected: "string[]",
		},
		{
			name:     "pointer type",
			goCode:   "package test\ntype T struct { F *string }",
			expected: "string",
		},
		{
			name:     "map type",
			goCode:   "package test\ntype T struct { F map[string]int }",
			expected: "Object.<string, number>",
		},
		{
			name:     "interface type",
			goCode:   "package test\ntype T struct { F interface{} }",
			expected: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.goCode, 0)
			if err != nil {
				t.Fatalf("Failed to parse Go code: %v", err)
			}

			// Extract the field type
			var fieldType ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if field, ok := n.(*ast.Field); ok && len(field.Names) > 0 {
					fieldType = field.Type
					return false
				}
				return true
			})

			if fieldType == nil {
				t.Fatal("Failed to extract field type from Go code")
			}

			result := goTypeToJSType(fieldType)
			if result != tt.expected {
				t.Errorf("goTypeToJSType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test goTypeToTSType with AST nodes
func TestGoTypeToTSType(t *testing.T) {
	tests := []struct {
		name     string
		goCode   string
		expected string
	}{
		{
			name:     "string type",
			goCode:   "package test\ntype T struct { F string }",
			expected: "string",
		},
		{
			name:     "int type",
			goCode:   "package test\ntype T struct { F int }",
			expected: "number",
		},
		{
			name:     "bool type",
			goCode:   "package test\ntype T struct { F bool }",
			expected: "boolean",
		},
		{
			name:     "slice type",
			goCode:   "package test\ntype T struct { F []string }",
			expected: "string[]",
		},
		{
			name:     "pointer type",
			goCode:   "package test\ntype T struct { F *string }",
			expected: "string",
		},
		{
			name:     "map type",
			goCode:   "package test\ntype T struct { F map[string]int }",
			expected: "Record<string, number>",
		},
		{
			name:     "interface type",
			goCode:   "package test\ntype T struct { F interface{} }",
			expected: "any",
		},
		{
			name:     "any type",
			goCode:   "package test\ntype T struct { F any }",
			expected: "any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.goCode, 0)
			if err != nil {
				t.Fatalf("Failed to parse Go code: %v", err)
			}

			// Extract the field type
			var fieldType ast.Expr
			ast.Inspect(file, func(n ast.Node) bool {
				if field, ok := n.(*ast.Field); ok && len(field.Names) > 0 {
					fieldType = field.Type
					return false
				}
				return true
			})

			if fieldType == nil {
				t.Fatal("Failed to extract field type from Go code")
			}

			result := goTypeToTSType(fieldType)
			if result != tt.expected {
				t.Errorf("goTypeToTSType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test generateJSDocType
func TestGenerateJSDocType(t *testing.T) {
	goCode := `
package test
type User struct {
	ID       int    ` + "`json:\"id\"`" + `
	Username string ` + "`json:\"username\"`" + `           // User login name
	Email    string ` + "`json:\"email,omitempty\"`" + `   // User email
	Active   bool   ` + "`json:\"active\"`" + `
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCode, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	// Extract struct fields
	var fields []*ast.Field
	ast.Inspect(file, func(n ast.Node) bool {
		if structType, ok := n.(*ast.StructType); ok {
			fields = structType.Fields.List
			return false
		}
		return true
	})

	if len(fields) == 0 {
		t.Fatal("Failed to extract struct fields")
	}

	result := generateJSDocType("User", fields)

	// Check that result contains expected elements
	expectedElements := []string{
		"@typedef {Object} User",
		"@property {number} id",
		"@property {string} username",
		"@property {string} [email]",
		"@property {boolean} active",
		"User login name",
		"User email",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("generateJSDocType() result missing %q\nGot:\n%s", expected, result)
		}
	}
}

// Test generateTypeScriptInterface
func TestGenerateTypeScriptInterface(t *testing.T) {
	goCode := `
package test
type User struct {
	ID       int    ` + "`json:\"id\"`" + `
	Username string ` + "`json:\"username\"`" + `           // User login name
	Email    string ` + "`json:\"email,omitempty\"`" + `   // User email
	Active   bool   ` + "`json:\"active\"`" + `
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCode, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	// Extract struct fields
	var fields []*ast.Field
	ast.Inspect(file, func(n ast.Node) bool {
		if structType, ok := n.(*ast.StructType); ok {
			fields = structType.Fields.List
			return false
		}
		return true
	})

	if len(fields) == 0 {
		t.Fatal("Failed to extract struct fields")
	}

	t.Run("interface with export", func(t *testing.T) {
		result := generateTypeScriptInterface("User", fields, false, true)

		expectedElements := []string{
			"export interface User {",
			"id: number;",
			"username: string;",
			"email?: string;",
			"active: boolean;",
			"/** User login name */",
			"/** User email */",
		}

		for _, expected := range expectedElements {
			if !strings.Contains(result, expected) {
				t.Errorf("generateTypeScriptInterface() result missing %q\nGot:\n%s", expected, result)
			}
		}
	})

	t.Run("type without export", func(t *testing.T) {
		result := generateTypeScriptInterface("User", fields, true, false)

		if strings.Contains(result, "export") {
			t.Error("generateTypeScriptInterface() should not contain 'export' when addExport=false")
		}

		if !strings.Contains(result, "type User = {") {
			t.Error("generateTypeScriptInterface() should use 'type' keyword when useTypeKeyword=true")
		}
	})
}

// Test extractFieldComment
func TestExtractFieldComment(t *testing.T) {
	goCode := `
package test
type User struct {
	// Doc comment
	Field1 string
	Field2 string // Inline comment
	Field3 string /* Block comment */
	Field4 string
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCode, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse Go code: %v", err)
	}

	// Extract struct fields
	var fields []*ast.Field
	ast.Inspect(file, func(n ast.Node) bool {
		if structType, ok := n.(*ast.StructType); ok {
			fields = structType.Fields.List
			return false
		}
		return true
	})

	tests := []struct {
		fieldIndex int
		expected   string
	}{
		{0, "Doc comment"},
		{1, "Inline comment"},
		{2, "Block comment"},
		{3, ""},
	}

	for _, tt := range tests {
		t.Run(fields[tt.fieldIndex].Names[0].Name, func(t *testing.T) {
			result := extractFieldComment(fields[tt.fieldIndex])
			if result != tt.expected {
				t.Errorf("extractFieldComment() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Benchmark toCamelCase
func BenchmarkToCamelCase(b *testing.B) {
	inputs := []string{"UserID", "FirstName", "APIKey", "HTTPServer"}
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_ = toCamelCase(input)
		}
	}
}

// Benchmark matchesExcludePattern
func BenchmarkMatchesExcludePattern(b *testing.B) {
	patterns := []string{"Internal*", "*Test", "API*", "*Secret*"}
	names := []string{"User", "InternalConfig", "APIResponse", "UserTest"}
	for i := 0; i < b.N; i++ {
		for _, name := range names {
			_ = matchesExcludePattern(name, patterns)
		}
	}
}
