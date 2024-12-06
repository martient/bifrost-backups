package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath checks if a given path is within allowed directories
func ValidatePath(path string, allowedPaths []string) error {
	cleanPath := filepath.Clean(path)
	
	// Convert to absolute paths for comparison
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	for _, allowedPath := range allowedPaths {
		cleanAllowedPath := filepath.Clean(allowedPath)
		absAllowedPath, err := filepath.Abs(cleanAllowedPath)
		if err != nil {
			continue
		}
		
		// Check if the path is within allowed directory
		if strings.HasPrefix(absPath, absAllowedPath) {
			return nil
		}
	}
	
	return fmt.Errorf("path is not within allowed directories")
}

// ValidateCommand checks if a command and its arguments are allowed
func ValidateCommand(cmd string, args []string, allowedCmds map[string][]string) error {
    // Check if command is in allowed list
    allowedArgs, ok := allowedCmds[cmd]
    if !ok {
        return fmt.Errorf("command not allowed: %s", cmd)
    }

    // Validate each argument against allowed patterns
    for _, arg := range args {
        valid := false
        for _, pattern := range allowedArgs {
            if strings.HasPrefix(arg, pattern) {
                valid = true
                break
            }
        }
        if !valid {
            return fmt.Errorf("invalid argument: %s", arg)
        }
    }

    return nil
}
