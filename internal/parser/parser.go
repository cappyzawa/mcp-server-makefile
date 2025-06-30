package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Regular expressions for parsing
	targetRegex   = regexp.MustCompile(`^([^#\s][^:=]*):(.*)$`)
	variableRegex = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*([?:+]?=)\s*(.*)$`)
	includeRegex  = regexp.MustCompile(`^-?include\s+(.+)$`)
	exportRegex   = regexp.MustCompile(`^export\s+([A-Za-z_][A-Za-z0-9_]*)`)
	phonyRegex    = regexp.MustCompile(`^\.PHONY:\s*(.*)$`)
)

// Parser is the main Makefile parser
type Parser struct {
	makefile *Makefile
	phony    map[string]bool
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		makefile: &Makefile{
			Targets:   make(map[string]*Target),
			Variables: make(map[string]*Variable),
			Includes:  []string{},
		},
		phony: make(map[string]bool),
	}
}

// ParseFile parses a Makefile from a file path
func (p *Parser) ParseFile(path string) (*Makefile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	p.makefile.Path = path
	return p.parse(file)
}

// parse parses a Makefile from a reader
func (p *Parser) parse(r io.Reader) (*Makefile, error) {
	scanner := bufio.NewScanner(r)
	lineNumber := 0
	var currentTarget *Target
	var lastComment string
	var continuedLine string

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Handle line continuations
		if strings.HasSuffix(line, "\\") {
			continuedLine += strings.TrimSuffix(line, "\\") + " "
			continue
		}
		if continuedLine != "" {
			line = continuedLine + line
			continuedLine = ""
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			currentTarget = nil
			continue
		}

		// Handle comments
		if strings.HasPrefix(line, "#") {
			lastComment = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}

		// Check for .PHONY targets
		if matches := phonyRegex.FindStringSubmatch(line); matches != nil {
			phonyTargets := strings.Fields(matches[1])
			for _, t := range phonyTargets {
				p.phony[t] = true
			}
			continue
		}

		// Check for include directives
		if matches := includeRegex.FindStringSubmatch(line); matches != nil {
			includes := strings.Fields(matches[1])
			p.makefile.Includes = append(p.makefile.Includes, includes...)
			continue
		}

		// Check for export directives
		if matches := exportRegex.FindStringSubmatch(line); matches != nil {
			if v, ok := p.makefile.Variables[matches[1]]; ok {
				v.IsExported = true
			}
			continue
		}

		// Check for variable assignments
		if matches := variableRegex.FindStringSubmatch(line); matches != nil {
			varType := SimpleAssignment
			switch matches[2] {
			case ":=":
				varType = RecursiveAssignment
			case "?=":
				varType = ConditionalAssignment
			case "+=":
				varType = AppendAssignment
			}

			var value string
			if varType == AppendAssignment {
				if existing, ok := p.makefile.Variables[matches[1]]; ok {
					value = existing.Value + " " + matches[3]
				} else {
					value = matches[3]
				}
			} else {
				value = matches[3]
			}

			p.makefile.Variables[matches[1]] = &Variable{
				Name:       matches[1],
				Value:      value,
				Type:       varType,
				LineNumber: lineNumber,
			}
			currentTarget = nil
			continue
		}

		// Check for targets
		if matches := targetRegex.FindStringSubmatch(line); matches != nil {
			targetNames := strings.Fields(matches[1])
			deps := strings.Fields(matches[2])

			for _, targetName := range targetNames {
				target := &Target{
					Name:         targetName,
					Dependencies: deps,
					Commands:     []string{},
					IsPhony:      p.phony[targetName],
					Description:  lastComment,
					LineNumber:   lineNumber,
				}
				p.makefile.Targets[targetName] = target
				currentTarget = target
			}
			lastComment = ""
			continue
		}

		// If we're in a target and the line starts with a tab, it's a command
		if currentTarget != nil && strings.HasPrefix(scanner.Text(), "\t") {
			command := strings.TrimPrefix(scanner.Text(), "\t")
			currentTarget.Commands = append(currentTarget.Commands, command)
			continue
		}

		// Reset current target if we hit a non-command line
		currentTarget = nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return p.makefile, nil
}

// ExpandVariable expands a variable with all its references resolved
func (p *Parser) ExpandVariable(name string) (string, error) {
	visited := make(map[string]bool)
	return p.expandVariableRecursive(name, visited)
}

func (p *Parser) expandVariableRecursive(name string, visited map[string]bool) (string, error) {
	if visited[name] {
		return "", fmt.Errorf("circular reference detected for variable: %s", name)
	}

	variable, ok := p.makefile.Variables[name]
	if !ok {
		// Check environment variables
		if envVal := os.Getenv(name); envVal != "" {
			return envVal, nil
		}
		return "", fmt.Errorf("variable not found: %s", name)
	}

	visited[name] = true
	defer delete(visited, name)

	// Expand variable references in the value
	value := variable.Value
	varRefRegex := regexp.MustCompile(`\$\(([^)]+)\)|\$\{([^}]+)\}`)
	
	expandedValue := varRefRegex.ReplaceAllStringFunc(value, func(match string) string {
		// Extract variable name from $(VAR) or ${VAR}
		varName := match[2 : len(match)-1]
		if expanded, err := p.expandVariableRecursive(varName, visited); err == nil {
			return expanded
		}
		return match // Keep original if expansion fails
	})

	return expandedValue, nil
}

// BuildDependencyGraph builds a dependency graph for all targets
func (p *Parser) BuildDependencyGraph() *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// Create nodes for all targets
	for name, target := range p.makefile.Targets {
		node := &DependencyNode{
			Name:         name,
			Dependencies: target.Dependencies,
			Dependents:   []string{},
		}
		graph.Nodes[name] = node
	}

	// Build reverse dependencies (dependents)
	for name, node := range graph.Nodes {
		for _, dep := range node.Dependencies {
			if depNode, ok := graph.Nodes[dep]; ok {
				depNode.Dependents = append(depNode.Dependents, name)
			}
		}
	}

	return graph
}

// GetTargetDependencies recursively gets all dependencies for a target
func (p *Parser) GetTargetDependencies(targetName string, maxDepth int) ([]string, error) {
	_, ok := p.makefile.Targets[targetName]
	if !ok {
		return nil, fmt.Errorf("target not found: %s", targetName)
	}

	visited := make(map[string]bool)
	deps := []string{}
	
	var collectDeps func(name string, depth int)
	collectDeps = func(name string, depth int) {
		if depth > maxDepth || visited[name] {
			return
		}
		visited[name] = true

		if t, ok := p.makefile.Targets[name]; ok {
			for _, dep := range t.Dependencies {
				if !visited[dep] {
					deps = append(deps, dep)
					collectDeps(dep, depth+1)
				}
			}
		}
	}

	collectDeps(targetName, 0)
	return deps, nil
}

// FindMakefiles finds all Makefiles in a directory tree
func FindMakefiles(root string, pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "Makefile|makefile|GNUmakefile|*.mk"
	}

	patterns := strings.Split(pattern, "|")
	makefiles := []string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		for _, p := range patterns {
			if matched, _ := filepath.Match(p, base); matched || base == p {
				makefiles = append(makefiles, path)
				break
			}
		}

		return nil
	})

	return makefiles, err
}