package resume

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRealData(t *testing.T) {
	r, err := Load("../../data")
	if err != nil {
		t.Fatal(err)
	}
	web, pdf := r.Filter(Web), r.Filter(PDF)
	if web.Profile.Phone != nil {
		t.Error("phone must stay off the public website")
	}
	if pdf.Profile.Phone == nil {
		t.Error("phone must be in the PDF")
	}
	for _, j := range web.Experience {
		if j.ID == "rco-engenharia" {
			t.Error("hidden job is shown on the web")
		}
	}
}

// copyData copies data/ to a temp dir so a test can break one file.
func copyData(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir("../../data", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("../../data", p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func replaceIn(t *testing.T, file, old, new string) {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), old) {
		t.Fatalf("%q not found in %s", old, file)
	}
	if err := os.WriteFile(file, []byte(strings.Replace(string(b), old, new, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMissingTranslationNamesThePath(t *testing.T) {
	dir := copyData(t)
	replaceIn(t, filepath.Join(dir, "projects.json"),
		`"pt": "Monorepo com deploy via Serverless Framework v4"`, `"es": "x"`)
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"projects.json at /items/0/highlights/0", "missing property 'pt'"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q, got:\n%v", want, err)
		}
	}
}

func TestHiddenAndPDFOnlyFlags(t *testing.T) {
	dir := copyData(t)
	// Hide a whole project, and keep one bullet off the PDF only.
	replaceIn(t, filepath.Join(dir, "projects.json"), `"id": "mail-sender",`, `"id": "mail-sender", "hidden": true,`)
	r, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []Target{Web, PDF} {
		for _, p := range r.Filter(target).Projects {
			if p.ID == "mail-sender" {
				t.Errorf("hidden project shown on %s", target)
			}
		}
	}
	count := func(t Target, id string) int {
		for _, j := range r.Filter(t).Experience {
			if j.ID == id {
				return len(j.Bullets)
			}
		}
		return -1
	}
	if w, p := count(Web, "idip"), count(PDF, "idip"); w <= p {
		t.Errorf("idip: web should have more bullets than the PDF (pdf:false), got web=%d pdf=%d", w, p)
	}
}

func TestDuplicateIDAndBadDates(t *testing.T) {
	dir := copyData(t)
	replaceIn(t, filepath.Join(dir, "projects.json"), `"id": "mail-sender"`, `"id": "travel-journal"`)
	replaceIn(t, filepath.Join(dir, "experience.json"), `"end": "2022-10"`, `"end": "2021-01"`)
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), `id "travel-journal" is already used`) || !strings.Contains(err.Error(), "ends (2021-01) before it starts") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDuration(t *testing.T) {
	units := [4]string{"yr", "yrs", "mo", "mos"}
	end := "2025-11"
	if got := Duration("2022-10", &end, "2026-10", units); got != "3 yrs 2 mos" { // LinkedIn: 3 anos 2 meses
		t.Errorf("got %q", got)
	}
	if got := Duration("2025-11", nil, "2026-10", units); got != "1 yr" {
		t.Errorf("got %q", got)
	}
}
