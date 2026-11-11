package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestUpdateEnvFileAddsMissingKey(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	exampleFile := filepath.Join(dir, ".env.example")

	if err := os.WriteFile(envFile, []byte("KEEP=existing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exampleFile, []byte("KEEP=template\nNEW=value\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := updateEnvFile(envFile, exampleFile); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(got)

	for _, want := range []string{"KEEP=existing", "NEW=value"} {
		if !strings.Contains(content, want) {
			t.Errorf("updated .env does not contain %q:\n%s", want, content)
		}
	}

	dateHeader := regexp.MustCompile(`(?m)^# Added by envcheck from ` + regexp.QuoteMeta(exampleFile) + ` on \d{4}-\d{2}-\d{2}$`)
	if !dateHeader.MatchString(content) {
		t.Errorf("updated .env is missing the dated header:\n%s", content)
	}
}
