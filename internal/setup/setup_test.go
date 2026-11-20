package setup

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/report"
)

func tempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", "https://github.com/owner/repo.git"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func opts(dir string) Options {
	return Options{
		Root:        dir,
		AuthorName:  "emil rahmat",
		AuthorEmail: "189325051+Emil-02@users.noreply.github.com",
	}
}

func mustExist(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRunCreatesFiles(t *testing.T) {
	dir := tempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Repo\n\nIsi asli.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(opts(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 4 {
		t.Fatalf("created = %v", result.Created)
	}

	config := mustExist(t, filepath.Join(dir, ".jaga.yml"))
	if !strings.Contains(config, "fail_on: high") {
		t.Fatal("config tidak valid")
	}

	workflow := mustExist(t, filepath.Join(dir, ".github", "workflows", "jaga.yml"))
	for _, want := range []string{"emil rahmat", "189325051+Emil-02@users.noreply.github.com", "secrets.JAGA_PAT", "jaga-pr.yml@main"} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow tidak mengandung %q", want)
		}
	}
	if strings.Contains(workflow, "github-actions[bot]") {
		t.Fatal("workflow tidak boleh memakai bot sebagai author")
	}

	readme := mustExist(t, filepath.Join(dir, "README.md"))
	if !strings.Contains(readme, "api/badge?owner=owner&repo=repo") || !strings.Contains(readme, "Isi asli.") {
		t.Fatalf("README = %q", readme)
	}

	data := mustExist(t, filepath.Join(dir, "jaga-report.json"))
	var payload report.Payload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Score < 0 || payload.Score > 100 || payload.GeneratedAt == "" || len(payload.Results) != 3 {
		t.Fatalf("laporan = %+v", payload)
	}
}

func TestRunIdempotent(t *testing.T) {
	dir := tempRepo(t)
	if _, err := Run(opts(dir)); err != nil {
		t.Fatal(err)
	}
	second, err := Run(opts(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Created) != 0 {
		t.Fatalf("run kedua menulis ulang: %v", second.Created)
	}
	readme := mustExist(t, filepath.Join(dir, "README.md"))
	if strings.Count(readme, "api/badge") != 1 {
		t.Fatal("badge duplikat")
	}
}

func TestRunForceOverwrites(t *testing.T) {
	dir := tempRepo(t)
	if _, err := Run(opts(dir)); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".jaga.yml")
	if err := os.WriteFile(path, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o := opts(dir)
	o.Force = true
	if _, err := Run(o); err != nil {
		t.Fatal(err)
	}
	if got := mustExist(t, path); !strings.Contains(got, "fail_on: high") {
		t.Fatalf("force tidak menimpa: %q", got)
	}
}

func TestValidateAuthor(t *testing.T) {
	if err := validateAuthor("emil rahmat", "189325051+Emil-02@users.noreply.github.com"); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"", "a@b.c"}, {"bot\nhack", "a@b.c"}, {"emil", "${{ secrets.X }}"}, {"emil", "a\"b"}} {
		if err := validateAuthor(pair[0], pair[1]); err == nil {
			t.Fatalf("seharusnya error: %q %q", pair[0], pair[1])
		}
	}
}

func TestRepoSlugSSH(t *testing.T) {
	dir := tempRepo(t)
	cmd := exec.Command("git", "-C", dir, "remote", "set-url", "origin", "git@github.com:owner/repo.git")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	owner, repo, err := repoSlug(dir)
	if err != nil || owner != "owner" || repo != "repo" {
		t.Fatalf("slug = %q/%q, err = %v", owner, repo, err)
	}
}

func TestRunWithoutRemoteSkipsBadge(t *testing.T) {
	dir := t.TempDir()
	result, err := Run(opts(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Created {
		if strings.Contains(item, "badge") {
			t.Fatal("badge tidak boleh dibuat tanpa remote")
		}
	}
}
