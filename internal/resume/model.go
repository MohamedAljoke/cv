// Package resume holds the resume data model, loads it from data/*.json and
// validates it against the JSON Schemas in data/schema.
package resume

import (
	"bytes"
	"encoding/json"
)

// Target is an output the data is rendered to.
type Target string

const (
	Web Target = "web"
	PDF Target = "pdf"
)

// LText is localized text. In JSON it is either a plain string (the same in
// every language) or an object with one string per language.
type LText struct {
	any string
	by  map[string]string
}

func (t *LText) UnmarshalJSON(b []byte) error {
	if bytes.HasPrefix(bytes.TrimSpace(b), []byte(`"`)) {
		return json.Unmarshal(b, &t.any)
	}
	return json.Unmarshal(b, &t.by)
}

func (t LText) MarshalJSON() ([]byte, error) {
	if t.by != nil {
		return json.Marshal(t.by)
	}
	return json.Marshal(t.any)
}

// In returns the text for lang.
func (t LText) In(lang string) string {
	if t.by != nil {
		return t.by[lang]
	}
	return t.any
}

func (t LText) IsZero() bool { return t.any == "" && len(t.by) == 0 }

// Vis holds the visibility flags shared by every item.
// hidden drops the item everywhere; pdf/web false drop it from that output only.
type Vis struct {
	Hidden bool  `json:"hidden"`
	PDFOn  *bool `json:"pdf"`
	WebOn  *bool `json:"web"`
}

// Shows reports whether the item belongs in the given output.
func (v Vis) Shows(t Target) bool {
	if v.Hidden {
		return false
	}
	switch t {
	case PDF:
		return v.PDFOn == nil || *v.PDFOn
	case Web:
		return v.WebOn == nil || *v.WebOn
	}
	return true
}

// Bullet is a line of text with its own visibility. In JSON it is either
// plain localized text or {"text": ..., "pdf": false, ...}.
type Bullet struct {
	Text LText  `json:"text"`
	Note string `json:"note"`
	Vis
}

func (b *Bullet) UnmarshalJSON(data []byte) error {
	var probe map[string]json.RawMessage
	if json.Unmarshal(data, &probe) == nil {
		if _, ok := probe["text"]; ok {
			type plain Bullet // no UnmarshalJSON method: avoids recursion
			return json.Unmarshal(data, (*plain)(b))
		}
	}
	return json.Unmarshal(data, &b.Text)
}

type Group struct {
	Title   LText    `json:"title"`
	Bullets []Bullet `json:"bullets"`
	Vis
}

type Job struct {
	ID         string   `json:"id"`
	Company    string   `json:"company"`
	CompanyURL string   `json:"companyUrl"`
	Blurb      LText    `json:"blurb"`
	Highlight  LText    `json:"highlight"`
	Title      LText    `json:"title"`
	Location   LText    `json:"location"`
	Start      string   `json:"start"`
	End        *string  `json:"end"`
	Groups     []Group  `json:"groups"`
	Bullets    []Bullet `json:"bullets"`
	Stack      []string `json:"stack"`
	Note       string   `json:"note"`
	Vis
}

type Project struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Summary    LText    `json:"summary"`
	Highlights []Bullet `json:"highlights"`
	Stack      []string `json:"stack"`
	Repo       string   `json:"repo"`
	Demo       string   `json:"demo"`
	DemoLabel  LText    `json:"demoLabel"`
	Note       string   `json:"note"`
	Vis
}

type SkillGroup struct {
	ID     string  `json:"id"`
	Label  LText   `json:"label"`
	Skills []LText `json:"skills"`
	Vis
}

type Link struct {
	Kind  string `json:"kind"`
	Label LText  `json:"label"`
	URL   string `json:"url"`
	Vis
}

type Education struct {
	ID        string `json:"id"`
	School    string `json:"school"`
	Degree    LText  `json:"degree"`
	Location  LText  `json:"location"`
	StartYear int    `json:"startYear"`
	EndYear   int    `json:"endYear"`
	Detail    LText  `json:"detail"`
	Vis
}

type Certification struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
	URL    string `json:"url"`
	Badge  string `json:"badge"`
	Note   string `json:"note"`
	Vis
}

type Language struct {
	Name  LText `json:"name"`
	Level LText `json:"level"`
}

type Phone struct {
	Value string `json:"value"`
	Web   bool   `json:"web"`
}

type Profile struct {
	Name           string          `json:"name"`
	Headline       LText           `json:"headline"`
	Location       LText           `json:"location"`
	Email          string          `json:"email"`
	Phone          *Phone          `json:"phone"`
	Links          []Link          `json:"links"`
	Summary        LText           `json:"summary"`
	Focus          []LText         `json:"focus"`
	CoreStack      []string        `json:"coreStack"`
	Education      []Education     `json:"education"`
	Certifications []Certification `json:"certifications"`
	Languages      []Language      `json:"languages"`
}

type Site struct {
	BaseURL   string   `json:"baseUrl"`
	Languages []string `json:"languages"`
	Sections  struct {
		Web []string `json:"web"`
		PDF []string `json:"pdf"`
	} `json:"sections"`
	PDF struct {
		FileName LText `json:"fileName"`
		Title    LText `json:"title"`
		MaxPages int   `json:"maxPages"`
	} `json:"pdf"`
	ATS struct {
		Keywords map[string][]string `json:"keywords"`
	} `json:"ats"`
	Labels map[string]LText `json:"labels"`
}

// Resume is everything under data/.
type Resume struct {
	Site       Site
	Profile    Profile
	Experience []Job
	Projects   []Project
	Skills     []SkillGroup
}

// Filter returns a copy with only what belongs in target: hidden items,
// items switched off for that output, and groups left empty are removed.
func (r Resume) Filter(t Target) Resume {
	out := r
	out.Experience = nil
	for _, j := range r.Experience {
		if !j.Shows(t) {
			continue
		}
		j.Bullets = filterBullets(j.Bullets, t)
		var groups []Group
		for _, g := range j.Groups {
			g.Bullets = filterBullets(g.Bullets, t)
			if g.Shows(t) && len(g.Bullets) > 0 {
				groups = append(groups, g)
			}
		}
		j.Groups = groups
		out.Experience = append(out.Experience, j)
	}
	out.Projects = nil
	for _, p := range r.Projects {
		if p.Shows(t) {
			p.Highlights = filterBullets(p.Highlights, t)
			out.Projects = append(out.Projects, p)
		}
	}
	out.Skills = keep(r.Skills, t)
	out.Profile.Links = keep(r.Profile.Links, t)
	out.Profile.Education = keep(r.Profile.Education, t)
	out.Profile.Certifications = keep(r.Profile.Certifications, t)
	if t == Web && r.Profile.Phone != nil && !r.Profile.Phone.Web {
		out.Profile.Phone = nil
	}
	return out
}

type shower interface{ Shows(Target) bool }

func keep[T shower](items []T, t Target) []T {
	var out []T
	for _, it := range items {
		if it.Shows(t) {
			out = append(out, it)
		}
	}
	return out
}

func filterBullets(bs []Bullet, t Target) []Bullet { return keep(bs, t) }
