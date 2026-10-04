// Package render turns the resume data into the static site, the printable
// resume pages and the JSON Resume export.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"cv/internal/resume"
)

// Options controls one build.
type Options struct {
	Root    string // repo root (templates/, assets/ and the outputs live here)
	Updated string // "YYYY-MM-DD" shown as "last updated"; also the "as of" month for durations
}

// Page is what every template receives.
type Page struct {
	Lang    string
	Target  resume.Target
	R       resume.Resume // already filtered for Target
	Prefix  string        // relative path from this page to the site root
	Updated string

	Sections []string
	Blocks   [][]string // web: consecutive compact sections grouped into one row
	PDFs     []PDFLink
	Alts     []Alt
	Self     string // absolute URL of this page
	Other    *Alt   // the same page in the next language
	OGImage  string
	JSONLD   template.JS
	DocTitle string
}

type PDFLink struct{ Lang, Href, Label, File string }

type Alt struct{ Lang, Href string }

// T localizes text.
func (p Page) T(t resume.LText) string { return p.lang(t) }

func (p Page) lang(t resume.LText) string { return t.In(p.Lang) }

// L returns a UI label; a missing label is a build error, not a blank.
func (p Page) L(key string) (string, error) {
	t, ok := p.R.Site.Labels[key]
	if !ok {
		return "", fmt.Errorf("site.json: missing label %q", key)
	}
	return t.In(p.Lang), nil
}

func (p Page) Dates(j resume.Job) string {
	present, _ := p.L("present")
	return resume.Range(p.Lang, j.Start, j.End, present)
}

func (p Page) Duration(j resume.Job) string {
	var u [4]string
	for i, k := range []string{"yr", "yrs", "mo", "mos"} {
		u[i], _ = p.L(k)
	}
	return resume.Duration(j.Start, j.End, p.Updated[:7], u)
}

func (p Page) Month(ym string) string { return resume.Month(p.Lang, ym) }

// Skills localizes a skill group's items.
func (p Page) Skills(g resume.SkillGroup) []string {
	out := make([]string, len(g.Skills))
	for i, s := range g.Skills {
		out[i] = s.In(p.Lang)
	}
	return out
}

// Current is the first job without an end date, if any.
func (p Page) Current() *resume.Job {
	for _, j := range p.R.Experience {
		if j.End == nil {
			return &j
		}
	}
	return nil
}

func (p Page) Has(section string) bool {
	for _, s := range p.Sections {
		if s == section {
			return true
		}
	}
	return false
}

// sectionCtx passes the page plus one section name into a sub-template.
type sectionCtx struct {
	P Page
	S string
}

var funcs = template.FuncMap{
	"dict":  func(p Page, s string) sectionCtx { return sectionCtx{p, s} },
	"join":  strings.Join,
	"host":  Host,
	"lower": strings.ToLower,
	"icon":  icon,
	"years": func(e resume.Education) string {
		switch {
		case e.StartYear > 0 && e.EndYear > 0:
			return fmt.Sprintf("%d – %d", e.StartYear, e.EndYear)
		case e.EndYear > 0:
			return fmt.Sprint(e.EndYear)
		}
		return ""
	},
}

// Host turns a URL into the short text printed on the resume:
// "https://www.linkedin.com/in/x/" → "linkedin.com/in/x".
func Host(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "mailto:")
	u = strings.TrimPrefix(u, "www.")
	return strings.TrimSuffix(u, "/")
}

// Result lists what Build wrote that later steps (PDF printing) need.
type Result struct {
	ResumePages map[string]string // lang → absolute path of resume/<lang>.html
	PDFFiles    map[string]string // lang → absolute path the PDF must be written to
	OGPages     map[string]string // lang → absolute path of a temporary OG card page
	OGImages    map[string]string // lang → absolute path of the PNG to write
	Pages       []string          // every HTML file written (for link checks)
}

// Build renders every page for every language.
func Build(r resume.Resume, o Options) (Result, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseGlob(filepath.Join(o.Root, "templates", "*.gohtml"))
	if err != nil {
		return Result{}, err
	}
	res := Result{
		ResumePages: map[string]string{}, PDFFiles: map[string]string{},
		OGPages: map[string]string{}, OGImages: map[string]string{},
	}
	langs := r.Site.Languages
	base := r.Site.BaseURL
	sitePath := func(lang string) string { // path relative to the root
		if lang == langs[0] {
			return ""
		}
		return lang + "/"
	}
	var pdfs []PDFLink
	for _, lang := range langs {
		file := r.Site.PDF.FileName.In(lang) + ".pdf"
		label, _ := Page{Lang: lang, R: r}.L("pdf_" + lang)
		pdfs = append(pdfs, PDFLink{Lang: lang, Href: "files/" + file, File: file, Label: label})
		res.PDFFiles[lang] = filepath.Join(o.Root, "files", file)
	}
	var alts []Alt
	for _, lang := range langs {
		alts = append(alts, Alt{Lang: lang, Href: base + sitePath(lang)})
	}

	// Scratch pages (the OG cards) live in .build/ (gitignored) so they can use
	// relative asset paths: html/template refuses file:// URLs.
	ogDir := filepath.Join(o.Root, ".build")

	for i, lang := range langs {
		web := r.Filter(resume.Web)
		pdf := r.Filter(resume.PDF)
		other := alts[(i+1)%len(alts)]

		// Website page.
		prefix := strings.Repeat("../", strings.Count(sitePath(lang), "/"))
		p := Page{
			Lang: lang, Target: resume.Web, R: web, Prefix: prefix, Updated: o.Updated,
			Sections: r.Site.Sections.Web, Blocks: blocks(r.Site.Sections.Web),
			PDFs: withPrefix(pdfs, prefix), Alts: alts, Self: base + sitePath(lang),
			OGImage: base + "assets/og-" + lang + ".png",
		}
		if len(alts) > 1 {
			p.Other = &other
			p.Other.Href = prefix + sitePath(other.Lang)
			if p.Other.Href == "" {
				p.Other.Href = "./"
			}
		}
		p.DocTitle = r.Profile.Name + " – " + p.T(r.Profile.Headline)
		p.JSONLD, err = personLD(p)
		if err != nil {
			return res, err
		}
		out := filepath.Join(o.Root, sitePath(lang), "index.html")
		if err := write(tmpl, "site", p, out); err != nil {
			return res, err
		}
		res.Pages = append(res.Pages, out)

		// Printable resume page (the PDF source, also published as HTML).
		rp := Page{
			Lang: lang, Target: resume.PDF, R: pdf, Prefix: "../", Updated: o.Updated,
			Sections: r.Site.Sections.PDF, PDFs: withPrefix(pdfs, "../"),
			Self: base + "resume/" + lang + ".html", DocTitle: r.Site.PDF.Title.In(lang),
		}
		out = filepath.Join(o.Root, "resume", lang+".html")
		if err := write(tmpl, "resume", rp, out); err != nil {
			return res, err
		}
		res.ResumePages[lang] = out
		res.Pages = append(res.Pages, out)

		// Social preview card, rendered to a temp page and screenshotted later.
		op := p
		op.Prefix = "../"
		out = filepath.Join(ogDir, "og-"+lang+".html")
		if err := write(tmpl, "og", op, out); err != nil {
			return res, err
		}
		res.OGPages[lang] = out
		res.OGImages[lang] = filepath.Join(o.Root, "assets", "og-"+lang+".png")
	}

	if err := writeJSONResume(r, o, filepath.Join(o.Root, "resume.json")); err != nil {
		return res, err
	}
	return res, nil
}

func withPrefix(in []PDFLink, prefix string) []PDFLink {
	out := make([]PDFLink, len(in))
	for i, l := range in {
		l.Href = prefix + l.Href
		out[i] = l
	}
	return out
}

// blocks groups consecutive compact sections (education, certifications,
// languages) so the site can lay them out side by side.
func blocks(sections []string) [][]string {
	compact := map[string]bool{"education": true, "certifications": true, "languages": true}
	var out [][]string
	for _, s := range sections {
		n := len(out)
		if compact[s] && n > 0 && compact[out[n-1][0]] {
			out[n-1] = append(out[n-1], s)
			continue
		}
		out = append(out, []string{s})
	}
	return out
}

func write(t *template.Template, name string, data any, path string) error {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return fmt.Errorf("render %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// personLD is schema.org structured data, read by search engines and AI tools.
func personLD(p Page) (template.JS, error) {
	var sameAs []string
	for _, l := range p.R.Profile.Links {
		if l.Kind != "website" {
			sameAs = append(sameAs, l.URL)
		}
	}
	var knows []string
	for _, g := range p.R.Skills {
		knows = append(knows, p.Skills(g)...)
	}
	ld := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "Person",
		"name":        p.R.Profile.Name,
		"jobTitle":    p.T(p.R.Profile.Headline),
		"description": p.T(p.R.Profile.Summary),
		"email":       "mailto:" + p.R.Profile.Email,
		"url":         p.Self,
		"sameAs":      sameAs,
		"knowsAbout":  knows,
		"knowsLanguage": func() []string {
			var out []string
			for _, l := range p.R.Profile.Languages {
				out = append(out, p.T(l.Name))
			}
			return out
		}(),
	}
	if j := p.Current(); j != nil {
		org := map[string]string{"@type": "Organization", "name": j.Company}
		if j.CompanyURL != "" {
			org["url"] = j.CompanyURL
		}
		ld["worksFor"] = org
	}
	var alumni []map[string]string
	for _, e := range p.R.Profile.Education {
		alumni = append(alumni, map[string]string{"@type": "CollegeOrUniversity", "name": e.School})
	}
	if alumni != nil {
		ld["alumniOf"] = alumni
	}
	var creds []map[string]string
	for _, c := range p.R.Profile.Certifications {
		creds = append(creds, map[string]string{"@type": "EducationalOccupationalCredential", "name": c.Name, "url": c.URL})
	}
	if creds != nil {
		ld["hasCredential"] = creds
	}
	b, err := json.MarshalIndent(ld, "", "  ")
	if err != nil {
		return "", err
	}
	// Keep "</script>" from ending the tag early.
	return template.JS(strings.ReplaceAll(string(b), "</", `<\/`)), nil
}
