package parser

// Target represents a Makefile target
type Target struct {
	Name         string
	Dependencies []string
	Commands     []string
	IsPhony      bool
	Description  string // From comment above target
	LineNumber   int
}

// Variable represents a Makefile variable
type Variable struct {
	Name       string
	Value      string
	IsExported bool
	IsOverride bool
	LineNumber int
	Type       VariableType
}

// VariableType represents the type of variable assignment
type VariableType int

const (
	SimpleAssignment VariableType = iota // VAR = value
	RecursiveAssignment                  // VAR := value
	ConditionalAssignment                // VAR ?= value
	AppendAssignment                     // VAR += value
)

// Makefile represents a parsed Makefile
type Makefile struct {
	Path      string
	Targets   map[string]*Target
	Variables map[string]*Variable
	Includes  []string
}

// DependencyGraph represents target dependencies
type DependencyGraph struct {
	Nodes map[string]*DependencyNode
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Name         string
	Dependencies []string
	Dependents   []string
}