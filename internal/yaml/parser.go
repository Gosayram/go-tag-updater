// Package yaml provides YAML file parsing and manipulation utilities
package yaml

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Gosayram/go-tag-updater/pkg/errors"
)

const (
	// DefaultIndentation defines the default YAML indentation
	DefaultIndentation = 2
	// MaxFileSize defines the maximum file size for processing
	MaxFileSize = 10 * 1024 * 1024 // 10MB
	// BackupExtension defines the file extension for backup files
	BackupExtension = ".backup"

	// MaxTagValueLength defines the maximum length for tag values
	MaxTagValueLength = 256
	// MinTagValueLength defines the minimum length for tag values
	MinTagValueLength = 1
)

// Parser handles YAML file parsing and manipulation
type Parser struct {
	preserveComments bool
	preserveOrder    bool
	indentation      int
}

// TagLocation represents the location of a tag in YAML structure
type TagLocation struct {
	Path   []string    // Path to the tag (e.g., ["spec", "containers", 0, "image"])
	Line   int         // Line number in original file
	Column int         // Column number in original file
	Value  interface{} // Current value
	Node   *yaml.Node  // Reference to YAML node
}

// ParseResult contains the result of YAML parsing
type ParseResult struct {
	Content         *yaml.Node
	TagLocations    []TagLocation
	OriginalContent string
	IsValid         bool
	Errors          []string
}

// UpdateOptions contains options for tag updates
type UpdateOptions struct {
	TagPath         []string // Path to the tag to update
	NewValue        string   // New value to set
	CreateIfMissing bool     // Create the tag if it doesn't exist
	BackupContent   bool     // Keep backup of original content
}

// NewParser creates a new YAML parser with default settings
func NewParser() *Parser {
	return &Parser{
		preserveComments: true,
		preserveOrder:    true,
		indentation:      DefaultIndentation,
	}
}

// NewParserWithOptions creates a new YAML parser with custom options
func NewParserWithOptions(preserveComments, preserveOrder bool, indentation int) *Parser {
	if indentation <= 0 {
		indentation = DefaultIndentation
	}

	return &Parser{
		preserveComments: preserveComments,
		preserveOrder:    preserveOrder,
		indentation:      indentation,
	}
}

// ParseContent parses YAML content and returns structured information
func (p *Parser) ParseContent(content string) (*ParseResult, error) {
	if content == "" {
		return nil, errors.NewValidationError("YAML content cannot be empty")
	}

	if len(content) > MaxFileSize {
		return nil, errors.NewValidationError(fmt.Sprintf(
			"YAML content too large: %d bytes (max %d)", len(content), MaxFileSize))
	}

	result := &ParseResult{
		OriginalContent: content,
		TagLocations:    []TagLocation{},
		Errors:          []string{},
	}

	// Parse the YAML content
	var rootNode yaml.Node
	err := yaml.Unmarshal([]byte(content), &rootNode)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML parsing error: %v", err))
		return result, errors.NewInvalidYAMLError(fmt.Sprintf("failed to parse YAML: %v", err))
	}

	result.Content = &rootNode
	result.IsValid = true

	// Find all tag locations
	p.findTagLocations(&rootNode, []string{}, result)

	return result, nil
}

// UpdateTag updates a specific tag value in the YAML content
func (p *Parser) UpdateTag(parseResult *ParseResult, options *UpdateOptions) (string, error) {
	if parseResult == nil {
		return "", errors.NewValidationError("parse result cannot be nil")
	}

	if options == nil {
		return "", errors.NewValidationError("update options cannot be nil")
	}

	if options.NewValue == "" {
		return "", errors.NewValidationError("new tag value cannot be empty")
	}

	if len(options.NewValue) > MaxTagValueLength {
		return "", errors.NewValidationError(fmt.Sprintf(
			"tag value too long: %d characters (max %d)", len(options.NewValue), MaxTagValueLength))
	}

	// Find the tag to update
	tagLocation := p.findTagByPath(parseResult, options.TagPath)
	if tagLocation == nil {
		if options.CreateIfMissing {
			return p.createAndUpdateTag(parseResult, options)
		}
		return "", errors.NewValidationError(fmt.Sprintf("tag not found at path: %v", options.TagPath))
	}

	// Update the tag value
	if tagLocation.Node != nil {
		tagLocation.Node.Value = options.NewValue
	}

	// Convert back to YAML string
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(p.indentation)

	err := encoder.Encode(parseResult.Content)
	if err != nil {
		return "", errors.NewInvalidYAMLError(fmt.Sprintf("failed to encode updated YAML: %v", err))
	}

	if err := encoder.Close(); err != nil {
		return "", errors.NewInvalidYAMLError(fmt.Sprintf("failed to close YAML encoder: %v", err))
	}
	return buf.String(), nil
}

// UpdateTagSimple provides a simple interface for updating a tag by searching for common patterns
func (p *Parser) UpdateTagSimple(content, newTag string) (string, error) {
	parseResult, err := p.ParseContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse YAML content: %w", err)
	}

	// Look for common tag patterns
	commonTagPaths := [][]string{
		{"tag"},
		{"image", "tag"},
		{"spec", "template", "spec", "containers", "image"},
		{"spec", "containers", "image"},
		{"metadata", "labels", "version"},
		{"version"},
	}

	for _, tagPath := range commonTagPaths {
		options := &UpdateOptions{
			TagPath:         tagPath,
			NewValue:        newTag,
			CreateIfMissing: false,
		}

		if p.findTagByPath(parseResult, tagPath) != nil {
			return p.UpdateTag(parseResult, options)
		}
	}

	return "", errors.NewValidationError("no suitable tag field found in YAML content")
}

// ValidateYAML validates YAML syntax and structure
func (p *Parser) ValidateYAML(content string) error {
	if content == "" {
		return errors.NewValidationError("YAML content cannot be empty")
	}

	var temp interface{}
	err := yaml.Unmarshal([]byte(content), &temp)
	if err != nil {
		return errors.NewInvalidYAMLError(fmt.Sprintf("invalid YAML syntax: %v", err))
	}

	return nil
}

// FormatYAML formats YAML content with consistent indentation
func (p *Parser) FormatYAML(content string) (string, error) {
	parseResult, err := p.ParseContent(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse YAML for formatting: %w", err)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(p.indentation)

	err = encoder.Encode(parseResult.Content)
	if err != nil {
		return "", errors.NewInvalidYAMLError(fmt.Sprintf("failed to format YAML: %v", err))
	}

	if err := encoder.Close(); err != nil {
		return "", errors.NewInvalidYAMLError(fmt.Sprintf("failed to close YAML encoder: %v", err))
	}
	return buf.String(), nil
}

// GetTagValue retrieves the value of a tag at the specified path
func (p *Parser) GetTagValue(parseResult *ParseResult, tagPath []string) (string, error) {
	tagLocation := p.findTagByPath(parseResult, tagPath)
	if tagLocation == nil {
		return "", errors.NewValidationError(fmt.Sprintf("tag not found at path: %v", tagPath))
	}

	if tagLocation.Node != nil && tagLocation.Node.Value != "" {
		return tagLocation.Node.Value, nil
	}

	return fmt.Sprintf("%v", tagLocation.Value), nil
}

// ListAllTags returns all tag locations found in the YAML content
func (p *Parser) ListAllTags(parseResult *ParseResult) []TagLocation {
	return parseResult.TagLocations
}

// findTagLocations recursively finds all tag locations in the YAML structure
func (p *Parser) findTagLocations(node *yaml.Node, path []string, result *ParseResult) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			p.findTagLocations(child, path, result)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value != "" {
				nodePath := make([]string, len(path)+1)
				copy(nodePath, path)
				nodePath[len(path)] = keyNode.Value

				// Check if this is a potential tag field
				if p.isTagField(keyNode.Value, valueNode) {
					location := TagLocation{
						Path:   nodePath,
						Line:   valueNode.Line,
						Column: valueNode.Column,
						Value:  valueNode.Value,
						Node:   valueNode,
					}
					result.TagLocations = append(result.TagLocations, location)
				}

				p.findTagLocations(valueNode, nodePath, result)
			}
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			indexPath := make([]string, len(path)+1)
			copy(indexPath, path)
			indexPath[len(path)] = fmt.Sprintf("[%d]", i)
			p.findTagLocations(child, indexPath, result)
		}
	}
}

// findTagByPath finds a tag at the specified path
func (p *Parser) findTagByPath(parseResult *ParseResult, targetPath []string) *TagLocation {
	for _, location := range parseResult.TagLocations {
		if p.pathsEqual(location.Path, targetPath) {
			return &location
		}
	}
	return nil
}

// createAndUpdateTag creates a new tag at the specified path and updates it
func (p *Parser) createAndUpdateTag(_ *ParseResult, _ *UpdateOptions) (string, error) {
	// This is a simplified implementation
	// In a full implementation, you would need to navigate and create the path structure
	return "", errors.NewValidationError("creating new tags is not implemented yet")
}

// isTagField determines if a field is likely a tag field based on common patterns
func (p *Parser) isTagField(key string, valueNode *yaml.Node) bool {
	if valueNode.Kind != yaml.ScalarNode {
		return false
	}

	tagPatterns := []string{
		"tag",
		"version",
		"image",
		"release",
	}

	keyLower := strings.ToLower(key)
	for _, pattern := range tagPatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}

	// Check if the value looks like a tag (contains version-like patterns)
	value := strings.ToLower(valueNode.Value)
	if strings.Contains(value, "v") || strings.Contains(value, ".v") || strings.Contains(value, "-") {
		return strings.Contains(value, "tag") || strings.Contains(value, "version") ||
			strings.Contains(key, "image")
	}

	return false
}

// pathsEqual compares two paths for equality
func (p *Parser) pathsEqual(path1, path2 []string) bool {
	if len(path1) != len(path2) {
		return false
	}

	for i, segment := range path1 {
		if segment != path2[i] {
			return false
		}
	}

	return true
}

// WriteToWriter writes formatted YAML content to an io.Writer
func (p *Parser) WriteToWriter(parseResult *ParseResult, writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(p.indentation)

	err := encoder.Encode(parseResult.Content)
	if err != nil {
		return errors.NewInvalidYAMLError(fmt.Sprintf("failed to write YAML: %v", err))
	}

	if err := encoder.Close(); err != nil {
		return errors.NewInvalidYAMLError(fmt.Sprintf("failed to close YAML encoder: %v", err))
	}

	return nil
}
