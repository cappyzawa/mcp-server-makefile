package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cappyzawa/mcp-server-makefile/internal/parser"
)

// Server implements the MCP server for Makefile exploration
type Server struct {
	parser *parser.Parser
	cache  map[string]*parser.Makefile
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	return &Server{
		parser: parser.NewParser(),
		cache:  make(map[string]*parser.Makefile),
	}
}

// Initialize implements the MCP initialize handler
func (s *Server) Initialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": "1.0",
		"serverInfo": map[string]interface{}{
			"name":    "mcp-server-makefile",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"available": true,
			},
		},
	}, nil
}

// ListTools implements the MCP tools/list handler
func (s *Server) ListTools(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{
		"tools": []interface{}{
			map[string]interface{}{
				"name":        "list_targets",
				"description": "List all targets in the Makefile",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the Makefile (optional, defaults to ./Makefile)",
						},
					},
				},
			},
			map[string]interface{}{
				"name":        "get_target",
				"description": "Get detailed information about a specific target",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"target": map[string]interface{}{
							"type":        "string",
							"description": "Target name",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the Makefile (optional)",
						},
					},
					"required": []string{"target"},
				},
			},
			map[string]interface{}{
				"name":        "get_dependencies",
				"description": "Get dependency graph for a target",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"target": map[string]interface{}{
							"type":        "string",
							"description": "Target name",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the Makefile (optional)",
						},
						"max_depth": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum dependency depth (optional)",
						},
					},
					"required": []string{"target"},
				},
			},
			map[string]interface{}{
				"name":        "list_variables",
				"description": "List all variables defined in the Makefile",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the Makefile (optional)",
						},
						"include_env": map[string]interface{}{
							"type":        "boolean",
							"description": "Include environment variables (default: false)",
						},
					},
				},
			},
			map[string]interface{}{
				"name":        "expand_variable",
				"description": "Expand a variable to its full value",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"variable": map[string]interface{}{
							"type":        "string",
							"description": "Variable name",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the Makefile (optional)",
						},
					},
					"required": []string{"variable"},
				},
			},
			map[string]interface{}{
				"name":        "find_makefiles",
				"description": "Find all Makefiles in the project",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"root": map[string]interface{}{
							"type":        "string",
							"description": "Root directory to search (optional, defaults to current directory)",
						},
						"pattern": map[string]interface{}{
							"type":        "string",
							"description": "File pattern to match (optional, defaults to common Makefile names)",
						},
					},
				},
			},
		},
	}, nil
}

// CallTool implements the MCP tools/call handler
func (s *Server) CallTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
	switch name {
	case "list_targets":
		return s.listTargets(args)
	case "get_target":
		return s.getTarget(args)
	case "get_dependencies":
		return s.getDependencies(args)
	case "list_variables":
		return s.listVariables(args)
	case "expand_variable":
		return s.expandVariable(args)
	case "find_makefiles":
		return s.findMakefiles(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) getMakefile(path string) (*parser.Makefile, error) {
	if path == "" {
		path = "Makefile"
	}

	// Check cache
	if mf, ok := s.cache[path]; ok {
		return mf, nil
	}

	// Parse the file
	mf, err := s.parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache[path] = mf
	return mf, nil
}

func (s *Server) listTargets(args json.RawMessage) (interface{}, error) {
	var params struct {
		Path string `json:"path,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	mf, err := s.getMakefile(params.Path)
	if err != nil {
		return nil, err
	}

	targets := []map[string]interface{}{}
	for name, target := range mf.Targets {
		targets = append(targets, map[string]interface{}{
			"name":         name,
			"description":  target.Description,
			"dependencies": target.Dependencies,
			"isPhony":      target.IsPhony,
			"lineNumber":   target.LineNumber,
		})
	}

	return map[string]interface{}{
		"targets": targets,
	}, nil
}

func (s *Server) getTarget(args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string `json:"target"`
		Path   string `json:"path,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	mf, err := s.getMakefile(params.Path)
	if err != nil {
		return nil, err
	}

	target, ok := mf.Targets[params.Target]
	if !ok {
		return nil, fmt.Errorf("target not found: %s", params.Target)
	}

	return map[string]interface{}{
		"name":         target.Name,
		"description":  target.Description,
		"dependencies": target.Dependencies,
		"commands":     target.Commands,
		"isPhony":      target.IsPhony,
		"lineNumber":   target.LineNumber,
	}, nil
}

func (s *Server) getDependencies(args json.RawMessage) (interface{}, error) {
	var params struct {
		Target   string `json:"target"`
		Path     string `json:"path,omitempty"`
		MaxDepth int    `json:"max_depth,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.MaxDepth == 0 {
		params.MaxDepth = 10 // Default max depth
	}

	mf, err := s.getMakefile(params.Path)
	if err != nil {
		return nil, err
	}

	// Update parser's makefile reference
	s.parser = parser.NewParser()
	s.parser.ParseFile(mf.Path)

	deps, err := s.parser.GetTargetDependencies(params.Target, params.MaxDepth)
	if err != nil {
		return nil, err
	}

	// Build dependency tree
	graph := s.parser.BuildDependencyGraph()
	node, ok := graph.Nodes[params.Target]
	if !ok {
		return nil, fmt.Errorf("target not found in dependency graph: %s", params.Target)
	}

	return map[string]interface{}{
		"target":       params.Target,
		"dependencies": deps,
		"graph": map[string]interface{}{
			"directDependencies": node.Dependencies,
			"dependents":         node.Dependents,
		},
	}, nil
}

func (s *Server) listVariables(args json.RawMessage) (interface{}, error) {
	var params struct {
		Path       string `json:"path,omitempty"`
		IncludeEnv bool   `json:"include_env,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	mf, err := s.getMakefile(params.Path)
	if err != nil {
		return nil, err
	}

	variables := []map[string]interface{}{}
	for name, variable := range mf.Variables {
		varTypeStr := ""
		switch variable.Type {
		case parser.SimpleAssignment:
			varTypeStr = "simple"
		case parser.RecursiveAssignment:
			varTypeStr = "recursive"
		case parser.ConditionalAssignment:
			varTypeStr = "conditional"
		case parser.AppendAssignment:
			varTypeStr = "append"
		}

		variables = append(variables, map[string]interface{}{
			"name":       name,
			"value":      variable.Value,
			"type":       varTypeStr,
			"isExported": variable.IsExported,
			"lineNumber": variable.LineNumber,
		})
	}

	// Include environment variables if requested
	if params.IncludeEnv {
		for _, env := range os.Environ() {
			if idx := len(env); idx > 0 {
				if pos := len(env); pos > 0 {
					for i, c := range env {
						if c == '=' {
							name := env[:i]
							value := env[i+1:]
							variables = append(variables, map[string]interface{}{
								"name":       name,
								"value":      value,
								"type":       "environment",
								"isExported": true,
								"lineNumber": -1,
							})
							break
						}
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"variables": variables,
	}, nil
}

func (s *Server) expandVariable(args json.RawMessage) (interface{}, error) {
	var params struct {
		Variable string `json:"variable"`
		Path     string `json:"path,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	mf, err := s.getMakefile(params.Path)
	if err != nil {
		return nil, err
	}

	// Update parser's makefile reference
	s.parser = parser.NewParser()
	s.parser.ParseFile(mf.Path)

	expanded, err := s.parser.ExpandVariable(params.Variable)
	if err != nil {
		return nil, err
	}

	variable := mf.Variables[params.Variable]
	original := ""
	if variable != nil {
		original = variable.Value
	}

	return map[string]interface{}{
		"variable": params.Variable,
		"original": original,
		"expanded": expanded,
	}, nil
}

func (s *Server) findMakefiles(args json.RawMessage) (interface{}, error) {
	var params struct {
		Root    string `json:"root,omitempty"`
		Pattern string `json:"pattern,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.Root == "" {
		params.Root = "."
	}

	makefiles, err := parser.FindMakefiles(params.Root, params.Pattern)
	if err != nil {
		return nil, err
	}

	// Convert to relative paths
	cwd, _ := os.Getwd()
	results := []map[string]interface{}{}
	for _, path := range makefiles {
		relPath, _ := filepath.Rel(cwd, path)
		info, _ := os.Stat(path)
		results = append(results, map[string]interface{}{
			"path":     path,
			"relative": relPath,
			"size":     info.Size(),
			"modified": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return map[string]interface{}{
		"makefiles": results,
		"count":     len(results),
	}, nil
}