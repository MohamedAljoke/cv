package render

import (
	"encoding/json"
	"fmt"
	"os"

	"cv/internal/resume"
)

// writeJSONResume exports the English web view in the jsonresume.org format,
// so tools such as Reactive Resume can import it.
func writeJSONResume(r resume.Resume, o Options, path string) error {
	lang := r.Site.Languages[0]
	w := r.Filter(resume.Web)
	p := Page{Lang: lang, R: w}

	type profile struct {
		Network  string `json:"network"`
		URL      string `json:"url"`
		Username string `json:"username,omitempty"`
	}
	var profiles []profile
	website := ""
	for _, l := range w.Profile.Links {
		if l.Kind == "website" {
			website = l.URL
			continue
		}
		profiles = append(profiles, profile{Network: p.T(l.Label), URL: l.URL})
	}

	var work []map[string]any
	for _, j := range w.Experience {
		var hl []string
		for _, g := range j.Groups {
			for _, b := range g.Bullets {
				hl = append(hl, p.T(b.Text))
			}
		}
		for _, b := range j.Bullets {
			hl = append(hl, p.T(b.Text))
		}
		item := map[string]any{
			"name": j.Company, "position": p.T(j.Title), "startDate": j.Start,
			"summary": p.T(j.Blurb), "highlights": hl, "url": j.CompanyURL,
		}
		if j.End != nil {
			item["endDate"] = *j.End
		}
		work = append(work, item)
	}

	var projects []map[string]any
	for _, pr := range w.Projects {
		var hl []string
		for _, b := range pr.Highlights {
			hl = append(hl, p.T(b.Text))
		}
		projects = append(projects, map[string]any{
			"name": pr.Name, "description": p.T(pr.Summary), "highlights": hl,
			"keywords": pr.Stack, "url": pr.Repo,
		})
	}

	var skills []map[string]any
	for _, g := range w.Skills {
		skills = append(skills, map[string]any{"name": p.T(g.Label), "keywords": p.Skills(g)})
	}

	var education []map[string]any
	for _, e := range w.Profile.Education {
		item := map[string]any{"institution": e.School, "studyType": p.T(e.Degree)}
		if e.StartYear > 0 {
			item["startDate"] = fmt.Sprint(e.StartYear)
		}
		if e.EndYear > 0 {
			item["endDate"] = fmt.Sprint(e.EndYear)
		}
		education = append(education, item)
	}

	var certs []map[string]any
	for _, c := range w.Profile.Certifications {
		item := map[string]any{"name": c.Name, "issuer": c.Issuer, "url": c.URL}
		if c.Date != "" {
			item["date"] = c.Date
		}
		certs = append(certs, item)
	}

	var langs []map[string]any
	for _, l := range w.Profile.Languages {
		langs = append(langs, map[string]any{"language": p.T(l.Name), "fluency": p.T(l.Level)})
	}

	doc := map[string]any{
		"$schema": "https://raw.githubusercontent.com/jsonresume/resume-schema/v1.0.0/schema.json",
		"basics": map[string]any{
			"name": w.Profile.Name, "label": p.T(w.Profile.Headline), "email": w.Profile.Email,
			"url": website, "summary": p.T(w.Profile.Summary), "profiles": profiles,
		},
		"work": work, "projects": projects, "skills": skills, "education": education,
		"certificates": certs, "languages": langs,
		"meta": map[string]string{"lastModified": o.Updated, "canonical": r.Site.BaseURL + "resume.json"},
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
