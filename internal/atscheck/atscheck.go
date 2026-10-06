// Package atscheck reads a generated PDF back the way an ATS parser would
// (plain text extraction) and checks what screeners rely on.
package atscheck

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"

	"cv/internal/resume"
)

type Report struct {
	Lang     string
	File     string
	Pages    int
	Title    string
	Lines    []string
	Problems []string
	Notes    []string
}

func (r Report) OK() bool { return len(r.Problems) == 0 }

func (r Report) String() string {
	var b strings.Builder
	status := "OK"
	if !r.OK() {
		status = "FAILED"
	}
	fmt.Fprintf(&b, "ATS check %s [%s] %s: %d page(s), %d text lines\n", r.Lang, status, r.File, r.Pages, len(r.Lines))
	for _, p := range r.Problems {
		fmt.Fprintf(&b, "  ✗ %s\n", p)
	}
	for _, n := range r.Notes {
		fmt.Fprintf(&b, "  · %s\n", n)
	}
	return b.String()
}

// Extract returns the PDF's text, one entry per visual line, plus page count
// and document title. Glyphs are grouped into lines by baseline and sorted
// left to right, the way most resume parsers rebuild text.
func Extract(path string) (lines []string, pages int, title string, err error) {
	f, rd, err := pdf.Open(path)
	if err != nil {
		return nil, 0, "", err
	}
	defer f.Close()
	pages = rd.NumPage()
	title = rd.Trailer().Key("Info").Key("Title").Text()
	for i := 1; i <= pages; i++ {
		glyphs := rd.Page(i).Content().Text
		type line struct {
			y      float64
			glyphs []pdf.Text
		}
		var ls []*line
		for _, g := range glyphs {
			var hit *line
			for _, l := range ls {
				if math.Abs(l.y-g.Y) < g.FontSize*0.35 {
					hit = l
					break
				}
			}
			if hit == nil {
				hit = &line{y: g.Y}
				ls = append(ls, hit)
			}
			hit.glyphs = append(hit.glyphs, g)
		}
		sort.SliceStable(ls, func(a, b int) bool { return ls[a].y > ls[b].y }) // top to bottom
		for _, l := range ls {
			sort.SliceStable(l.glyphs, func(a, b int) bool { return l.glyphs[a].X < l.glyphs[b].X })
			var b strings.Builder
			for j, g := range l.glyphs {
				// A jump much wider than a character means a gap (e.g. right-aligned dates).
				if j > 0 {
					prev := l.glyphs[j-1]
					if g.X-prev.X > prev.FontSize*1.2 && prev.S != " " && g.S != " " {
						b.WriteByte(' ')
					}
				}
				b.WriteString(g.S)
			}
			if s := strings.TrimSpace(spaces.ReplaceAllString(b.String(), " ")); s != "" {
				lines = append(lines, s)
			}
		}
	}
	return lines, pages, title, nil
}

// Links returns the targets of every clickable link (URI annotation) in the PDF.
func Links(path string) (map[string]bool, error) {
	f, rd, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]bool{}
	for i := 1; i <= rd.NumPage(); i++ {
		annots := rd.Page(i).V.Key("Annots")
		for j := 0; j < annots.Len(); j++ {
			if uri := annots.Index(j).Key("A").Key("URI").Text(); uri != "" {
				out[uri] = true
			}
		}
	}
	return out, nil
}

var spaces = regexp.MustCompile(`\s+`)

func norm(s string) string { return strings.ToLower(spaces.ReplaceAllString(s, " ")) }

// Check runs every rule for one language's PDF.
func Check(path, lang string, r resume.Resume) Report {
	rep := Report{Lang: lang, File: path}
	lines, pages, title, err := Extract(path)
	if err != nil {
		rep.Problems = append(rep.Problems, "cannot read PDF: "+err.Error())
		return rep
	}
	rep.Lines, rep.Pages, rep.Title = lines, pages, title
	text := norm(strings.Join(lines, " "))
	p := r.Filter(resume.PDF)

	if max := r.Site.PDF.MaxPages; pages > max {
		rep.Problems = append(rep.Problems, fmt.Sprintf("%d pages, limit is %d (set \"pdf\": false on lower-value bullets)", pages, max))
	}
	if want := r.Site.PDF.Title.In(lang); title != want {
		rep.Problems = append(rep.Problems, fmt.Sprintf("PDF title is %q, want %q", title, want))
	}

	// Section headings: each on its own line, in the configured order.
	last, lastName := -1, ""
	for _, sec := range r.Site.Sections.PDF {
		if !hasContent(p, sec) {
			continue
		}
		h := norm(r.Site.Labels[sec].In(lang))
		at := -1
		for i, l := range lines {
			if norm(l) == h {
				at = i
				break
			}
		}
		switch {
		case at < 0:
			rep.Problems = append(rep.Problems, fmt.Sprintf("heading %q is not on a line of its own", h))
		case at < last:
			rep.Problems = append(rep.Problems, fmt.Sprintf("heading %q comes before %q", h, lastName))
		default:
			last, lastName = at, h
		}
	}

	// Name and email must be plain text; profile links are clickable labels,
	// so check the label text and the link target inside the PDF.
	must := []string{p.Profile.Name, p.Profile.Email}
	links, err := Links(path)
	if err != nil {
		rep.Problems = append(rep.Problems, "cannot read links: "+err.Error())
	}
	for _, l := range p.Profile.Links {
		must = append(must, l.Label.In(lang))
		if !links[l.URL] {
			rep.Problems = append(rep.Problems, fmt.Sprintf("no clickable link to %s", l.URL))
		}
	}
	for _, pr := range p.Projects {
		if l := pr.Link(); l != "" && !links[l] {
			rep.Problems = append(rep.Problems, fmt.Sprintf("no clickable link to %s", l))
		}
	}
	for _, c := range p.Profile.Certifications {
		if c.URL != "" && !links[c.URL] {
			rep.Problems = append(rep.Problems, fmt.Sprintf("no clickable link to %s", c.URL))
		}
	}
	for _, j := range p.Experience {
		must = append(must, j.Company, j.Title.In(lang))
	}
	for _, m := range must {
		if !strings.Contains(text, norm(m)) {
			rep.Problems = append(rep.Problems, fmt.Sprintf("%q is not in the extracted text", m))
		}
	}

	// Keywords the target roles ask for.
	var missing []string
	for _, k := range r.Site.ATS.Keywords[lang] {
		if !containsWord(text, norm(k)) {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		rep.Problems = append(rep.Problems, "missing keywords: "+strings.Join(missing, ", "))
	}

	// The Google Docs export bug: text broken into one word per line.
	single := 0
	for _, l := range lines {
		if !strings.Contains(strings.TrimSpace(l), " ") {
			single++
		}
	}
	if len(lines) > 0 {
		ratio := float64(single) / float64(len(lines))
		note := fmt.Sprintf("one-word lines: %d of %d (%.0f%%)", single, len(lines), ratio*100)
		if ratio > 0.2 {
			rep.Problems = append(rep.Problems, note+": text is fragmented, parsers will mangle it")
		} else {
			rep.Notes = append(rep.Notes, note)
		}
	}
	rep.Notes = append(rep.Notes, fmt.Sprintf("title %q, %d clickable links", title, len(links)))
	return rep
}

func hasContent(r resume.Resume, sec string) bool {
	switch sec {
	case "summary":
		return !r.Profile.Summary.IsZero()
	case "experience":
		return len(r.Experience) > 0
	case "projects":
		return len(r.Projects) > 0
	case "skills":
		return len(r.Skills) > 0
	case "education":
		return len(r.Profile.Education) > 0
	case "certifications":
		return len(r.Profile.Certifications) > 0
	case "languages":
		return len(r.Profile.Languages) > 0
	}
	return false
}

// containsWord matches k as a whole word/phrase, so "Go" doesn't match "Google".
func containsWord(text, k string) bool {
	re := regexp.MustCompile(`(^|[^\pL\pN])` + regexp.QuoteMeta(k) + `($|[^\pL\pN])`)
	return re.MatchString(text)
}

// Match compares a job description with the resume text: which known terms
// the job mentions, and which of those the resume is missing.
func Match(jobDescription string, resumeLines []string, vocabulary []string) (have, missing []string) {
	jd := norm(jobDescription)
	text := norm(strings.Join(resumeLines, " "))
	seen := map[string]bool{}
	for _, v := range vocabulary {
		k := norm(v)
		if k == "" || seen[k] || !containsWord(jd, k) {
			continue
		}
		seen[k] = true
		if containsWord(text, k) {
			have = append(have, v)
		} else {
			missing = append(missing, v)
		}
	}
	sort.Strings(have)
	sort.Strings(missing)
	return have, missing
}
