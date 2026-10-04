package render

import "html/template"

// Inline SVG icons (Lucide-style strokes, 24×24). They are decorative:
// every icon sits next to visible text, so they are hidden from screen readers.
var icons = map[string]string{
	"download": `<path d="M12 15V3"/><path d="m7 10 5 5 5-5"/><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>`,
	"mail":     `<rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-10 6L2 7"/>`,
	"phone":    `<path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1 1 .4 1.9.7 2.8a2 2 0 0 1-.5 2.1L8 9.9a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.8.7a2 2 0 0 1 1.8 2z"/>`,
	"pin":      `<path d="M20 10c0 5-8 12-8 12s-8-7-8-12a8 8 0 0 1 16 0z"/><circle cx="12" cy="10" r="3"/>`,
	"linkedin": `<path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-4 0v7h-4v-7a6 6 0 0 1 6-6z"/><rect x="2" y="9" width="4" height="12"/><circle cx="4" cy="4" r="2"/>`,
	"github":   `<path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.1-1.3-.3-2.5-1-3.5.3-1.2.3-2.4 0-3.5 0 0-1 0-3 1.5a10.3 10.3 0 0 0-5.5 0C8 2 7 2 7 2c-.3 1.1-.3 2.3 0 3.5-.7 1-1.1 2.2-1 3.5 0 3.5 3 5.5 6 5.5-.4.5-.7 1-.8 1.6-.2.6-.3 1.3-.2 1.9v4"/><path d="M9 18c-4.5 2-5-2-7-2"/>`,
	"credly":   `<circle cx="12" cy="8" r="6"/><path d="M15.5 12.9 17 22l-5-3-5 3 1.5-9.1"/>`,
	"website":  `<circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>`,
	"external": `<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>`,
	"play":     `<circle cx="12" cy="12" r="10"/><path d="m10 8 6 4-6 4z"/>`,
	"code":     `<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>`,
	"sun":      `<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2m-7.1-17.1 1.4 1.4m11.4 11.4 1.4 1.4M2 12h2m16 0h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>`,
	"moon":     `<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9z"/>`,
	"globe":    `<circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>`,
	"briefcase": `<rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>`,
	"award":    `<circle cx="12" cy="8" r="6"/><path d="M15.5 12.9 17 22l-5-3-5 3 1.5-9.1"/>`,
}

func icon(name string) template.HTML {
	body, ok := icons[name]
	if !ok {
		return ""
	}
	return template.HTML(`<svg class="icon" aria-hidden="true" focusable="false" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`)
}
