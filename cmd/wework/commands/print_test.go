package commands

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPrintCommandStructure(t *testing.T) {
	cmd := NewPrintCommand(dummyAuth)

	if cmd.Use != "print" {
		t.Errorf("expected command Use to be 'print', got %q", cmd.Use)
	}

	expectedAliases := []string{"printhub", "print-hub", "printqueue", "print-queue"}
	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found in %v", alias, cmd.Aliases)
		}
	}

	subCommands := map[string]bool{
		"queue": false,
		"add":   false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := subCommands[sub.Name()]; ok {
			subCommands[sub.Name()] = true
		}
	}

	for name, found := range subCommands {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestPrintAddValidation(t *testing.T) {
	tempDir := t.TempDir()
	validFilePath := filepath.Join(tempDir, "test.pdf")
	if err := os.WriteFile(validFilePath, []byte("%PDF-1.4 test content"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	emptyFilePath := filepath.Join(tempDir, "empty.pdf")
	if err := os.WriteFile(emptyFilePath, []byte{}, 0600); err != nil {
		t.Fatalf("failed to create empty temp file: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		errorContains string
	}{
		{
			name:          "missing file path",
			args:          []string{},
			errorContains: "file path is required",
		},
		{
			name:          "nonexistent file",
			args:          []string{filepath.Join(tempDir, "nonexistent.pdf")},
			errorContains: "failed to access file",
		},
		{
			name:          "directory instead of file",
			args:          []string{tempDir},
			errorContains: "is a directory",
		},
		{
			name:          "empty file",
			args:          []string{emptyFilePath},
			errorContains: "is empty",
		},
		{
			name:          "invalid color mode",
			args:          []string{validFilePath, "--color", "sepia"},
			errorContains: "invalid color mode",
		},
		{
			name:          "invalid orientation",
			args:          []string{validFilePath, "--orientation", "upside-down"},
			errorContains: "invalid orientation",
		},
		{
			name:          "invalid sides",
			args:          []string{validFilePath, "--sides", "triplex"},
			errorContains: "invalid sides",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPrintAddCommand(dummyAuth)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("expected error containing %q, got: %v", tt.errorContains, err)
			}
		})
	}
}
