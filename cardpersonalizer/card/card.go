package card

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moov-io/iso8583/encoding"
)

var CardNameToIndex = map[string]int{
	"AmericanExpress": 1,
	"DinersClub":      2,
	"DinersClubUS":    3,
	"Discover":        4,
	"JCB":             5,
	"Laser":           6,
	"Maestro":         7,
	"Mastercard":      8,
	"Solo":            9,
	"Unionpay":        10,
	"Visa":            11,
	"Mir":             12,
}

// Operation represents a single file modification operation
type Operation struct {
	LineNumber  int      // 1-indexed line number
	Tags        []string // Pattern to find and cut at (e.g., "0x5A")
	Value       string   // New BCD-encoded data to insert
	Encoding    encoding.Encoder
	Description string // Tag for logging (e.g., "PAN", "ExpiryDate")
}

// createIsolatedWorkspace creates a unique workspace for a request
func createIsolatedWorkspace(requestID string) (string, error) {
	workspaceDir := fmt.Sprintf("javacard_%s", requestID)

	// Remove workspace if it already exists (cleanup from previous runs)
	if err := os.RemoveAll(workspaceDir); err != nil {
		return "", fmt.Errorf("failed to cleanup existing workspace: %v", err)
	}

	// Copy the entire directory to the isolated workspace
	if err := copyDir("javacard", workspaceDir); err != nil {
		return "", fmt.Errorf("failed to copy javacard directory: %v", err)
	}

	log.Printf("Created isolated workspace: %s", workspaceDir)
	return workspaceDir, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

// cleanupWorkspace removes the isolated workspace
func cleanupWorkspace(workspaceDir string) error {
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("failed to cleanup workspace %s: %v", workspaceDir, err)
	}
	log.Printf("Cleaned up workspace: %s", workspaceDir)
	return nil
}

// UpdateEMVStaticData updates the EMVStaticData.java file with multiple operations
func UpdateEMVStaticData(operations []Operation, requestID string) error {
	// Create isolated workspace
	workspaceDir, err := createIsolatedWorkspace(requestID)
	if err != nil {
		return err
	}

	// Update the file path to use the isolated workspace
	filePath := filepath.Join(workspaceDir, "src/openemv/EMVStaticData.java")

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Process each operation
	for _, op := range operations {
		// Validate line number (convert to 0-indexed)
		if op.LineNumber < 1 || op.LineNumber > len(lines) {
			return fmt.Errorf("invalid line number %d for operation %s", op.LineNumber, op.Description)
		}

		lineIndex := op.LineNumber - 1 // Convert to 0-indexed
		line := lines[lineIndex]

		// Find the cut pattern in the line
		var pattern string
		for _, tag := range op.Tags {
			pattern = fmt.Sprintf("0x%s, ", tag)
		}

		patternIndex := strings.Index(line, pattern)
		patternIndex += len(pattern)

		encodedValue, err := op.Encoding.Encode([]byte(op.Value))
		if err != nil {
			return fmt.Errorf("failed to encode value: %v", err)
		}
		var encodedValueHex string
		for _, b := range encodedValue {
			encodedValueHex += fmt.Sprintf("(byte)0x%02X, ", b)
		}

		// Cut the line at the pattern and stitch the new data
		newLine := line[:patternIndex] + fmt.Sprintf("0x%02X, %s", len(encodedValue), encodedValueHex)
		lines[lineIndex] = newLine

		log.Printf("Updated %s in line %d: %s: %s %X", op.Description, op.LineNumber, op.Tags, op.Value, encodedValue)
	}

	// Write the file back
	newContent := strings.Join(lines, "\n")
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Printf("Successfully updated EMVStaticData.java with %d operations in workspace %s", len(operations), workspaceDir)
	return nil
}

// FlashCardWithBinary runs the ant reinstall command in the isolated workspace
func FlashCardWithBinary(requestID string) error {
	workspaceDir := fmt.Sprintf("javacard_%s", requestID)
	defer cleanupWorkspace(workspaceDir)

	// Change to the isolated workspace directory where build.xml is located
	if err := os.Chdir(workspaceDir); err != nil {
		return fmt.Errorf("failed to change to workspace directory %s: %v", workspaceDir, err)
	}
	defer os.Chdir("..") // Change back to original directory

	// Execute ant reinstall command
	cmd := exec.Command("ant", "reinstall")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Executing: ant reinstall in workspace %s", workspaceDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ant reinstall failed in workspace %s: %v", workspaceDir, err)
	}

	log.Printf("Successfully executed ant reinstall in workspace %s", workspaceDir)
	return nil
}
