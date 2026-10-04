// Command build generates the portfolio site, the printable resume pages and
// the PDFs from data/*.json, then checks the PDFs the way an ATS would.
//
//	go run ./cmd/build              # everything
//	go run ./cmd/build -pdf=false   # site only (fast, no Chrome needed)
//	go run ./cmd/build -check       # only re-check the committed PDFs
//	go run ./cmd/build -match job.txt [-lang pt]
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cv/internal/atscheck"
	"cv/internal/chrome"
	"cv/internal/render"
	"cv/internal/resume"
)

func main() {
	root := flag.String("root", ".", "repository root")
	withPDF := flag.Bool("pdf", true, "print the PDFs and social images with headless Chrome")
	checkOnly := flag.Bool("check", false, "skip rendering; only run the ATS check on the existing PDFs")
	match := flag.String("match", "", "job description text file: report which of its terms the resume covers")
	lang := flag.String("lang", "", "language for -match (default: first in site.json)")
	chromeBin := flag.String("chrome", "", "path to Chrome/Chromium (default: $CHROME or PATH)")
	date := flag.String("date", "", "'last updated' date YYYY-MM-DD (default: last git commit touching data/)")
	flag.Parse()

	if err := run(*root, *withPDF, *checkOnly, *match, *lang, *chromeBin, *date); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(root string, withPDF, checkOnly bool, match, lang, chromeBin, date string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	r, err := resume.Load(filepath.Join(root, "data"))
	if err != nil {
		return err
	}
	if lang == "" {
		lang = r.Site.Languages[0]
	}
	pdfPath := func(l string) string {
		return filepath.Join(root, "files", r.Site.PDF.FileName.In(l)+".pdf")
	}

	if match != "" {
		return runMatch(r, match, lang, pdfPath(lang))
	}

	if !checkOnly {
		if date == "" {
			date = lastUpdated(root)
		}
		res, err := render.Build(r, render.Options{Root: root, Updated: date})
		if err != nil {
			return err
		}
		fmt.Printf("rendered %d pages (updated %s)\n", len(res.Pages), date)
		if withPDF {
			bin, err := chrome.Find(chromeBin)
			if err != nil {
				return err
			}
			c := chrome.Chrome{Bin: bin}
			for _, l := range r.Site.Languages {
				if err := c.PDF(res.ResumePages[l], res.PDFFiles[l]); err != nil {
					return err
				}
				fmt.Println("pdf  ", rel(root, res.PDFFiles[l]))
				if err := c.Screenshot(res.OGPages[l], res.OGImages[l], 1200, 630); err != nil {
					return err
				}
				fmt.Println("image", rel(root, res.OGImages[l]))
			}
		}
		os.RemoveAll(filepath.Join(root, ".build")) // scratch OG pages
	}

	failed := false
	for _, l := range r.Site.Languages {
		p := pdfPath(l)
		if _, err := os.Stat(p); err != nil {
			fmt.Printf("ATS check %s skipped: %s not built yet\n", l, rel(root, p))
			continue
		}
		rep := atscheck.Check(p, l, r)
		rep.File = rel(root, p)
		fmt.Print(rep)
		failed = failed || !rep.OK()
	}
	if failed {
		return fmt.Errorf("ATS check failed")
	}
	return nil
}

func runMatch(r resume.Resume, jdFile, lang, pdf string) error {
	jd, err := os.ReadFile(jdFile)
	if err != nil {
		return err
	}
	lines, _, _, err := atscheck.Extract(pdf)
	if err != nil {
		return fmt.Errorf("read %s (build it first): %w", pdf, err)
	}
	var vocab []string
	for _, g := range r.Skills {
		for _, s := range g.Skills {
			for _, l := range r.Site.Languages {
				vocab = append(vocab, s.In(l))
			}
		}
	}
	for _, j := range r.Experience {
		vocab = append(vocab, j.Stack...)
	}
	for _, p := range r.Projects {
		vocab = append(vocab, p.Stack...)
	}
	for _, ks := range r.Site.ATS.Keywords {
		vocab = append(vocab, ks...)
	}
	vocab = append(vocab, extraVocabulary...)
	have, missing := atscheck.Match(string(jd), lines, vocab)
	fmt.Printf("Job description terms covered by the %s resume: %d/%d\n", lang, len(have), len(have)+len(missing))
	if len(have) > 0 {
		fmt.Println("  ✓ " + strings.Join(have, ", "))
	}
	if len(missing) > 0 {
		fmt.Println("  ✗ missing: " + strings.Join(missing, ", "))
		fmt.Println("  Add the ones you really have to a bullet (with context), not just to the skills list.")
	}
	return nil
}

// extraVocabulary adds common backend terms that aren't in the data yet, so
// -match can point out when a job asks for them.
var extraVocabulary = []string{
	"Kubernetes", "Kafka", "gRPC", "Protobuf", "MySQL", "Elasticsearch", "OpenTelemetry", "Prometheus",
	"Grafana", "Datadog", "New Relic", "GCP", "Azure", "Linux", "Kotlin", "Rust", "C#", ".NET", "Ruby",
	"PHP", "Spring", "Django", "FastAPI", "DDD", "TDD", "SOLID", "Clean Architecture", "Hexagonal",
	"CQRS", "Event Sourcing", "Saga", "Idempotency", "Observability", "SRE", "Pix", "Open Finance",
	"PLD", "AML", "LGPD", "PCI", "OAuth", "JWT", "Microsserviços", "Mensageria", "Escalabilidade",
}

// lastUpdated is the date of the last commit that touched data/, so the
// output only changes when the content does. Falls back to today.
func lastUpdated(root string) string {
	out, err := exec.Command("git", "-C", root, "log", "-1", "--format=%cs", "--", "data").Output()
	if d := strings.TrimSpace(string(out)); err == nil && len(d) == 10 {
		return d
	}
	return time.Now().Format("2006-01-02")
}

func rel(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil {
		return r
	}
	return p
}
