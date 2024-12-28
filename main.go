package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var goToJSMappings = map[string]string{
	"string":  "string",
	"int":     "number",
	"int32":   "number",
	"int64":   "number",
	"float32": "number",
	"float64": "number",
	"bool":    "boolean",
}

// generateJSDocType converts a Go struct’s AST fields into a JSDoc typedef string.
func generateJSDocType(structName string, fields []*ast.Field) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("/**\n * @typedef {Object} %s\n", structName))

	for _, field := range fields {
		// Skip embedded or anonymous fields
		if len(field.Names) == 0 {
			continue
		}
		fieldName := field.Names[0].Name

		// Convert the Go type (ast.Expr) into a string
		goTypeStr := fmt.Sprintf("%s", field.Type)
		jsType, found := goToJSMappings[goTypeStr]
		if !found {
			// Fall back to '*' which in JSDoc means "any type"
			jsType = "*"
		}

		// Extract the JSON tag (if present), e.g. `json:"foo,omitempty"`
		jsonTag := ""
		if field.Tag != nil {
			tagValue := field.Tag.Value
			parts := strings.Split(tagValue, "json:\"")
			if len(parts) > 1 {
				jsonPart := parts[1]
				// up to the next quote
				jsonTag = strings.Split(jsonPart, "\"")[0]
			}
		}

		// If no JSON tag was found, use the Go field name
		if jsonTag == "" {
			jsonTag = fieldName
		}

		buf.WriteString(fmt.Sprintf(" * @property {%s} %s\n", jsType, jsonTag))
	}

	buf.WriteString(" */\n")
	return buf.String()
}

func main() {
	// Expect exactly two arguments:
	//   1) Directory with .go files (non-recursive)
	//   2) Output .js file
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <go_structs_dir> <output_js_file>\n", os.Args[0])
		os.Exit(1)
	}
	goDir := os.Args[1]
	outputFile := os.Args[2]

	// Create a new token.FileSet for parsing
	fset := token.NewFileSet()
	// Parse all .go files in the directory (non-recursive)
	pkgs, err := parser.ParseDir(fset, goDir, nil, 0)
	if err != nil {
		log.Fatalf("Could not parse directory %q: %v", goDir, err)
	}

	var jsDocBuf bytes.Buffer

	// For simplicity, omit the Git auto-detection from this snippet:
	jsDocBuf.WriteString("// Generated JSDoc definitions\n\n")

	// Traverse each package and file
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				// Check the doc comments for the marker
				skipStruct := false
				if genDecl.Doc != nil {
					for _, comment := range genDecl.Doc.List {
						// If the comment contains our marker, skip
						if strings.Contains(comment.Text, "struct2jsdoc: nogen") {
							skipStruct = true
							break
						}
					}
				}

				// If skipStruct is true, we don't process any of the structs in this GenDecl
				if skipStruct {
					continue
				}

				// Each spec is a type specification
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					// Check if this TypeSpec is a struct
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					// You could also check typeSpec.Doc if doc comments are placed specifically on typeSpec
					// e.g., for multiple type specs in a single GenDecl.
					// But typically, doc comments apply to the entire GenDecl block.

					// Generate JSDoc for this struct
					jsDoc := generateJSDocType(typeSpec.Name.Name, structType.Fields.List)
					jsDocBuf.WriteString(jsDoc + "\n")
				}
			}
		}
	}

	// Write all JSDoc to the specified output file
	err = ioutil.WriteFile(outputFile, jsDocBuf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Error writing to file %q: %v", outputFile, err)
	}

	fmt.Printf("JSDoc types successfully written to %s\n", outputFile)
}
