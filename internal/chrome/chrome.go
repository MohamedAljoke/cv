// Package chrome drives a local headless Chrome/Chromium through its command
// line to print PDFs and take screenshots. No CDP library is needed.
package chrome

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Find returns the Chrome binary: the explicit path if given, then $CHROME,
// then the usual names on PATH.
func Find(explicit string) (string, error) {
	for _, c := range []string{explicit, os.Getenv("CHROME")} {
		if c != "" {
			return c, nil
		}
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	if p := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"; fileExists(p) {
		return p, nil
	}
	return "", errors.New("chrome not found: install Chrome/Chromium or pass -chrome /path/to/chrome")
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// Chrome runs one headless command at a time with a throwaway profile, so it
// never touches (or waits on) a Chrome you have open.
type Chrome struct{ Bin string }

func (c Chrome) run(args ...string) error {
	profile, err := os.MkdirTemp("", "cv-chrome-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(profile)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	base := []string{
		"--headless", "--disable-gpu", "--no-sandbox", "--no-first-run",
		"--disable-extensions", "--hide-scrollbars", "--force-color-profile=srgb",
		"--user-data-dir=" + profile,
		"--run-all-compositor-stages-before-draw", "--virtual-time-budget=4000",
	}
	cmd := exec.CommandContext(ctx, c.Bin, append(base, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", filepath.Base(c.Bin), err, out)
	}
	return nil
}

// PDF prints an HTML file to a tagged PDF (structure tags help parsers keep
// reading order). Page size and margins come from the page's @page CSS.
func (c Chrome) PDF(htmlFile, pdfFile string) error {
	if err := os.MkdirAll(filepath.Dir(pdfFile), 0o755); err != nil {
		return err
	}
	tmp := pdfFile + ".tmp"
	if err := c.run("--no-pdf-header-footer", "--export-tagged-pdf", "--generate-pdf-document-outline",
		"--print-to-pdf="+tmp, fileURL(htmlFile)); err != nil {
		return err
	}
	if !fileExists(tmp) {
		return fmt.Errorf("chrome did not write %s", pdfFile)
	}
	return os.Rename(tmp, pdfFile)
}

// Screenshot renders an HTML file at width×height to a PNG.
//
// Headless Chrome's viewport can be shorter than --window-size (newer
// versions subtract browser UI), so it captures a taller window and crops.
func (c Chrome) Screenshot(htmlFile, pngFile string, width, height int) error {
	if err := os.MkdirAll(filepath.Dir(pngFile), 0o755); err != nil {
		return err
	}
	raw := pngFile + ".raw.png"
	defer os.Remove(raw)
	if err := c.run(fmt.Sprintf("--window-size=%d,%d", width, height+240), "--screenshot="+raw, fileURL(htmlFile)); err != nil {
		return err
	}
	return crop(raw, pngFile, width, height)
}

func crop(src, dst string, width, height int) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	img, err := png.Decode(in)
	if err != nil {
		return err
	}
	b := img.Bounds()
	if b.Dx() < width || b.Dy() < height {
		return fmt.Errorf("screenshot is %dx%d, smaller than %dx%d", b.Dx(), b.Dy(), width, height)
	}
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(f, out); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func fileURL(path string) string {
	abs, _ := filepath.Abs(path)
	return "file://" + filepath.ToSlash(abs)
}
