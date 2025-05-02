package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var defaultExample = ".env.example"
var defaultEnv = ".env"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "envcheck",
		Short: "Helps you check your env files",
		Run: func(cmd *cobra.Command, args []string) {
			// cmd.Help()
			// Custom help looks cleaner
			showHelp()
		},
	}

	var listCmd = &cobra.Command{
		Short: "List all .env.* files and .env.example files and their diffs",
		Use:   "list [path]",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cmdList,
	}
	listCmd.Flags().StringP("path", "p", ".", "Path to search for env files")

	var createCmd = &cobra.Command{
		Short: "Create .env from .env.example",
		Use:   "create",
		RunE:  cmdCreate,
	}
	createCmd.Flags().StringP("env-file", "e", defaultEnv, "Path to the environment file")
	createCmd.Flags().StringP("example-file", "x", defaultExample, "Path to the example file")

	var updateCmd = &cobra.Command{
		Short: "Update an .env file with missing keys from an .env.example file",
		Use:   "update",
		RunE:  cmdUpdate,
	}
	updateCmd.Flags().StringP("env-file", "e", defaultEnv, "Path to the environment file")
	updateCmd.Flags().StringP("example-file", "x", defaultExample, "Path to the example file")

	var scanCmd = &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan all text files in the specified path to identify env vars",
		Long:  "This command scans all text files in the specified path to identify environment variables. Default path is '.'. Files larger than 5MB are skipped.",
		RunE:  cmdScan,
	}
	scanCmd.Flags().StringP("path", "p", ".", "Path to scan for env vars")

	// rootCmd.AddCommand(listCmd, createCmd, updateCmd, scanCmd)
	rootCmd.AddCommand(listCmd, createCmd, updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cmdList(cmd *cobra.Command, args []string) error {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		path = "."
	}

	envFiles, exampleFiles, err := listEnvFiles(defaultExample, path)
	if err != nil {
		return err
	}

	fmt.Printf("Listing env files in path: %s\n", path)

	if len(exampleFiles) == 0 {
		fmt.Printf("✗ No example files (like %s) found.\n", defaultExample)
	}

	fmt.Printf("Found %d example files and %d env files.\n\n", len(exampleFiles), len(envFiles))

	for _, exampleFile := range exampleFiles {
		exampleVars, err := parseEnvFile(exampleFile)
		if err != nil {
			return err
		}

		// Determine the corresponding env file
		baseName := exampleFile
		if strings.HasSuffix(baseName, ".example") {
			baseName = strings.TrimSuffix(baseName, ".example")
		} else if strings.Contains(baseName, ".example.") {
			baseName = strings.Replace(baseName, ".example.", ".", 1)
		} else if baseName == defaultExample {
			baseName = ".env"
		}

		envFile := baseName

		if _, err := os.Stat(envFile); !os.IsNotExist(err) {
			envVars, err := parseEnvFile(envFile)
			if err != nil {
				return err
			}
			missingKeys := getDifferences(exampleVars, envVars)
			if len(missingKeys) > 0 {
				fmt.Printf("⚠ %s is missing %d keys from %s:\n", envFile, len(missingKeys), exampleFile)
				for _, key := range missingKeys {
					fmt.Printf("  - %s\n", key)
				}
			} else {
				fmt.Printf("✓ %s is in sync with %s\n", envFile, exampleFile)
			}
		} else {
			fmt.Printf("⚠ %s doesn't exist (template available: %s)\n", envFile, exampleFile)
		}

		fmt.Println()
	}

	return nil
}

func cmdCreate(cmd *cobra.Command, args []string) error {
	envFile := cmd.Flag("env-file").Value.String()
	exampleFile := cmd.Flag("example-file").Value.String()

	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		return fmt.Errorf("✗ Error: Example file %s not found.", exampleFile)
	}
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("✗ Error: Env file %s not found.", envFile)
	}

	return createEnvFile(envFile, exampleFile)
}

func cmdUpdate(cmd *cobra.Command, args []string) error {
	envFile := cmd.Flag("env-file").Value.String()
	exampleFile := cmd.Flag("example-file").Value.String()

	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		return fmt.Errorf("✗ Error: Example file %s not found.", exampleFile)
	}
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf("✗ Error: Env file %s not found.", envFile)
	}

	return updateEnvFile(envFile, exampleFile)
}

func showHelp() {
	var helpStr string
	helpStr = `
Usage:
  envcheck [command] [options]

Commands:
  envcheck list [dir] <env_file> <example_file>     - List env files and difference
  envcheck create <env_file> <example_file>         - Create env file from example
  envcheck update <env_file> <example_file>         - Update env file with missing keys

Flags:
  --env_file defaults to ".env"
  --example_file defaults to ".env.example"

Examples:
  envcheck list
  envcheck list ./deploy/
  envcheck scan
  envcheck scan ./build/
  envcheck create .env   
  envcheck create prod/.env prod/.env.example
  envcheck create .env.staging .env.example
  envcheck update .env
  envcheck update .env.development dev/.env.example
`
	fmt.Println(helpStr)
}

func listEnvFiles(defaultExample string, path string) ([]string, []string, error) {
	var envFiles, exampleFiles []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(info.Name(), ".env") {
			if strings.HasSuffix(info.Name(), ".example") || strings.Contains(info.Name(), ".example.") || info.Name() == defaultExample {
				exampleFiles = append(exampleFiles, path)
			} else {
				envFiles = append(envFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return envFiles, exampleFiles, nil
}

func parseEnvFile(filename string) (map[string]string, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return map[string]string{}, nil
	}

	envVars := map[string]string{}
	file, err := os.Open(filename)
	if err != nil {
		return envVars, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		re := regexp.MustCompile(`^([A-Za-z0-9_]+)=(.*?)$`)
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			key := match[1]
			value := match[2]
			envVars[key] = value
		}
	}

	return envVars, scanner.Err()
}

func getDifferences(exampleVars, envVars map[string]string) []string {
	var missingKeys []string
	for key := range exampleVars {
		if _, exists := envVars[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}
	return missingKeys
}

func updateEnvFile(envFile, exampleFile string) error {
	envVars, err := parseEnvFile(envFile)
	if err != nil {
		return err
	}
	exampleVars, err := parseEnvFile(exampleFile)
	if err != nil {
		return err
	}

	missingKeys := getDifferences(exampleVars, envVars)
	if len(missingKeys) == 0 {
		fmt.Printf("✓ %s is in sync with %s\n", envFile, exampleFile)
		return nil
	}

	// Create directory if it doesn't exist
	dirPath := filepath.Dir(envFile)
	if dirPath != "" {
		os.MkdirAll(dirPath, 0755)
	}

	// Create or update env file
	exists := true
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		exists = false
	}
	file, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if exists {
		_, err = file.WriteString(fmt.Sprintf("\n# Added by envcheck from %s\n", exampleFile))
		if err != nil {
			return err
		}
	}

	for _, key := range missingKeys {
		_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, exampleVars[key]))
		if err != nil {
			return err
		}
	}

	fmt.Printf("✓ Added %d missing keys to %s\n", len(missingKeys), envFile)
	for _, key := range missingKeys {
		fmt.Printf("  + %s\n", key)
	}

	return nil
}

func createEnvFile(envFile, exampleFile string) error {
	if _, err := os.Stat(envFile); !os.IsNotExist(err) {
		return fmt.Errorf("✗ Error: %s already exists. Use 'update' instead.", envFile)
	}

	exampleVars, err := parseEnvFile(exampleFile)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dirPath := filepath.Dir(envFile)
	if dirPath != "" {
		os.MkdirAll(dirPath, 0755)
	}

	file, err := os.Create(envFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range exampleVars {
		_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			return err
		}
	}

	fmt.Printf("✓ Created %s with %d keys from %s\n", envFile, len(exampleVars), exampleFile)
	return nil
}

func cmdScan(cmd *cobra.Command, args []string) error {
	path := cmd.Flag("path").Value.String()
	maxFileSize := int64(5 * 1024 * 1024) // 5MB

	// TODO: Doesn't ignore methods correctly
	return scanEnvVars(path, maxFileSize)
}

func scanEnvVars(path string, maxFileSize int64) error {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Size() > maxFileSize {
			return nil
		}

		if !isTextFile(filePath) {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		envVars := findEnvVars(string(content))
		if len(envVars) > 0 {
			fmt.Printf("Found %d env vars in %s:\n", len(envVars), filePath)
			for _, varName := range envVars {
				fmt.Printf("  - %s\n", varName)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func isTextFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	contentType := http.DetectContentType(buffer)
	return strings.HasPrefix(contentType, "text/")
}

func findEnvVars(content string) []string {
	re := regexp.MustCompile(`\b[A-Z_][A-Z0-9_]*\b`)
	matches := re.FindAllString(content, -1)

	uniqueVars := make(map[string]bool)
	for _, match := range matches {
		// _ is usually for unused vars
		// ignore period prefix as they tend to be methods
		if match != "_" && !strings.HasPrefix(match, ".") {
			uniqueVars[match] = true
		}
	}

	var result []string
	for varName := range uniqueVars {
		result = append(result, varName)
	}

	return result
}
