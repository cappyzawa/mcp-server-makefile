package parser

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	parser := NewParser()
	testFile := filepath.Join("testdata", "simple.mk")

	mf, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Test targets
	expectedTargets := []string{"all", "build", "main.o", "utils.o", "test", "clean"}
	if len(mf.Targets) != len(expectedTargets) {
		t.Errorf("Expected %d targets, got %d", len(expectedTargets), len(mf.Targets))
	}

	// Test specific target
	if target, ok := mf.Targets["build"]; ok {
		if len(target.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies for 'build', got %d", len(target.Dependencies))
		}
		if target.Description != "Build the application" {
			t.Errorf("Expected description 'Build the application', got '%s'", target.Description)
		}
	} else {
		t.Error("Target 'build' not found")
	}

	// Test variables
	if len(mf.Variables) < 2 {
		t.Errorf("Expected at least 2 variables, got %d", len(mf.Variables))
	}

	if cc, ok := mf.Variables["CC"]; ok {
		if cc.Value != "gcc" {
			t.Errorf("Expected CC=gcc, got CC=%s", cc.Value)
		}
	} else {
		t.Error("Variable 'CC' not found")
	}

	// Test PHONY targets
	if all, ok := mf.Targets["all"]; ok {
		if !all.IsPhony {
			t.Error("Target 'all' should be PHONY")
		}
	}
}

func TestExpandVariable(t *testing.T) {
	parser := NewParser()
	
	// Create a simple makefile with variable references
	mf := &Makefile{
		Variables: map[string]*Variable{
			"BASE": {Name: "BASE", Value: "/usr/local"},
			"DIR":  {Name: "DIR", Value: "$(BASE)/bin"},
			"PATH": {Name: "PATH", Value: "$(DIR):$(HOME)/bin"},
		},
	}
	parser.makefile = mf

	// Test simple expansion
	expanded, err := parser.ExpandVariable("DIR")
	if err != nil {
		t.Fatalf("Failed to expand variable: %v", err)
	}
	if expanded != "/usr/local/bin" {
		t.Errorf("Expected '/usr/local/bin', got '%s'", expanded)
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	parser := NewParser()
	testFile := filepath.Join("testdata", "simple.mk")

	mf, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	graph := parser.BuildDependencyGraph()

	// Test that all targets are in the graph
	for name := range mf.Targets {
		if _, ok := graph.Nodes[name]; !ok {
			t.Errorf("Target '%s' not found in dependency graph", name)
		}
	}

	// Test specific dependencies
	if node, ok := graph.Nodes["build"]; ok {
		if len(node.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies for 'build', got %d", len(node.Dependencies))
		}
	}

	// Test reverse dependencies
	if node, ok := graph.Nodes["main.o"]; ok {
		found := false
		for _, dep := range node.Dependents {
			if dep == "build" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'build' to depend on 'main.o'")
		}
	}
}

func TestGetTargetDependencies(t *testing.T) {
	parser := NewParser()
	testFile := filepath.Join("testdata", "simple.mk")

	_, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	deps, err := parser.GetTargetDependencies("all", 5)
	if err != nil {
		t.Fatalf("Failed to get dependencies: %v", err)
	}

	// "all" depends on "build" and "test"
	// "build" depends on "main.o" and "utils.o"
	// So we should get at least those 4 dependencies
	if len(deps) < 4 {
		t.Errorf("Expected at least 4 dependencies, got %d", len(deps))
	}
}