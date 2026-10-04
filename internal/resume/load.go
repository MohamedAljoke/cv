package resume

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Load reads and validates every file in dataDir. Any schema or consistency
// problem is returned as one error listing each offending JSON path.
func Load(dataDir string) (Resume, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return Resume{}, err
	}
	var r Resume
	var problems []string
	files := []struct {
		name string
		dst  any
	}{
		{"site", &r.Site},
		{"profile", &r.Profile},
		{"experience", &itemsOf[Job]{&r.Experience}},
		{"projects", &itemsOf[Project]{&r.Projects}},
		{"skills", &itemsOf[SkillGroup]{&r.Skills}},
	}
	for _, f := range files {
		file := filepath.Join(abs, f.name+".json")
		raw, err := os.ReadFile(file)
		if err != nil {
			return Resume{}, err
		}
		schema := filepath.Join(abs, "schema", f.name+".schema.json")
		problems = append(problems, validate(schema, raw, f.name+".json")...)
		if len(problems) > 0 {
			continue // decoding invalid data would only add noise
		}
		if err := json.Unmarshal(raw, f.dst); err != nil {
			problems = append(problems, fmt.Sprintf("%s.json: %v", f.name, err))
		}
	}
	if len(problems) == 0 {
		problems = append(problems, r.consistency()...)
	}
	if len(problems) > 0 {
		return Resume{}, errors.New("invalid data:\n  - " + strings.Join(problems, "\n  - "))
	}
	return r, nil
}

// itemsOf decodes the {"items": [...]} wrapper used by list files (a wrapper
// object is needed so the file can carry its own "$schema").
type itemsOf[T any] struct{ dst *[]T }

func (w *itemsOf[T]) UnmarshalJSON(b []byte) error {
	var v struct {
		Items []T `json:"items"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*w.dst = v.Items
	return nil
}

func validate(schemaFile string, raw []byte, name string) []string {
	c := jsonschema.NewCompiler()
	sch, err := c.Compile(schemaFile)
	if err != nil {
		return []string{fmt.Sprintf("%s: schema: %v", name, err)}
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return []string{fmt.Sprintf("%s: not valid JSON: %v", name, err)}
	}
	err = sch.Validate(inst)
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return nil
	}
	// Collect leaf errors per JSON location.
	p := message.NewPrinter(language.English)
	byLoc := map[string][]string{}
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			loc := "/" + strings.Join(e.InstanceLocation, "/")
			byLoc[loc] = append(byLoc[loc], e.ErrorKind.LocalizedString(p))
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)

	var out []string
	for loc, msgs := range byLoc {
		// Localized text is "string OR {en, pt}", so the validator reports every
		// alternative. When a language is missing, keep only what matters.
		if anyContains(msgs, "missing property 'en'", "missing property 'pt'") {
			var keep []string
			for _, m := range msgs {
				lang := strings.HasPrefix(m, "missing property 'en'") || strings.HasPrefix(m, "missing property 'pt'")
				extra := strings.HasPrefix(m, "additional properties") && !strings.Contains(m, "'en'") && !strings.Contains(m, "'pt'")
				if lang || extra {
					keep = append(keep, m)
				}
			}
			msgs = append(keep, `localized text needs both "en" and "pt" (or one plain string for all languages)`)
		}
		seen := map[string]bool{}
		for _, m := range msgs {
			if !seen[m] {
				seen[m] = true
				out = append(out, fmt.Sprintf("%s at %s: %s", name, loc, m))
			}
		}
	}
	sort.Strings(out)
	return out
}

func anyContains(msgs []string, subs ...string) bool {
	for _, m := range msgs {
		for _, s := range subs {
			if strings.Contains(m, s) {
				return true
			}
		}
	}
	return false
}

// consistency checks rules a JSON Schema can't express.
func (r Resume) consistency() []string {
	var out []string
	ids := map[string]string{}
	addID := func(file, id string) {
		if prev, ok := ids[id]; ok {
			out = append(out, fmt.Sprintf("%s: id %q is already used in %s", file, id, prev))
		}
		ids[id] = file
	}
	for _, j := range r.Experience {
		addID("experience.json", j.ID)
		if j.End != nil && *j.End < j.Start {
			out = append(out, fmt.Sprintf("experience.json: %s ends (%s) before it starts (%s)", j.ID, *j.End, j.Start))
		}
	}
	for _, p := range r.Projects {
		addID("projects.json", p.ID)
	}
	for _, s := range r.Skills {
		addID("skills.json", s.ID)
	}
	for _, sec := range append(r.Site.Sections.Web, r.Site.Sections.PDF...) {
		if _, ok := r.Site.Labels[sec]; !ok {
			out = append(out, fmt.Sprintf("site.json: section %q has no entry in labels", sec))
		}
	}
	for _, lang := range r.Site.Languages {
		if r.Site.PDF.FileName.In(lang) == "" || r.Site.PDF.Title.In(lang) == "" {
			out = append(out, fmt.Sprintf("site.json: pdf.fileName/title missing for %q", lang))
		}
	}
	return out
}
