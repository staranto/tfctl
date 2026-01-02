// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	md2man "github.com/cpuguy83/go-md2man/v2/md2man"
)

// Minimal doc generator:
// - Reads docs/commands/*.md as canonical command docs
// - Generates:
//   - docs/man/share/man1/tfctl-<cmd>.1 via md2man (convert full markdown)
//   - docs/tldr/tfctl-<cmd>.md using the Quick examples block and short description

func main() {
	var (
		repoRoot           string
		commandsDir        string
		manOutDir          string
		tldrOutDir         string
		writeOnlyIfChanged bool
	)

	flag.StringVar(&repoRoot, "root", ".", "repo root (default current dir)")
	flag.BoolVar(&writeOnlyIfChanged, "only-if-changed", true, "only write files if content changed")
	flag.Parse()

	commandsDir = filepath.Join(repoRoot, "docs", "commands")
	manOutDir = filepath.Join(repoRoot, "docs", "man", "share", "man1")
	tldrOutDir = filepath.Join(repoRoot, "docs", "tldr")

	if err := os.MkdirAll(manOutDir, 0o755); err != nil {
		fatalf("creating man output dir: %v", err)
	}
	if err := os.MkdirAll(tldrOutDir, 0o755); err != nil {
		fatalf("creating tldr output dir: %v", err)
	}

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		fatalf("reading commands dir %s: %v", commandsDir, err)
	}

	var processed int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		cmd := strings.TrimSuffix(e.Name(), ".md")
		inPath := filepath.Join(commandsDir, e.Name())
		raw, err := os.ReadFile(inPath)
		if err != nil {
			fatalf("reading %s: %v", inPath, err)
		}

		// Generate man page from full markdown
		manBytes := md2man.Render(raw)
		manPath := filepath.Join(manOutDir, fmt.Sprintf("tfctl-%s.1", cmd))
		if err := writeFileIfChanged(manPath, manBytes, writeOnlyIfChanged); err != nil {
			fatalf("writing man page for %s: %v", cmd, err)
		}

		// Generate TLDR page from short description + quick examples
		title, shortDesc := extractTitleAndShortDesc(string(raw))
		examples := extractQuickExamples(string(raw))
		tldr := buildTLDR(cmd, title, shortDesc, examples)
		tldrPath := filepath.Join(tldrOutDir, fmt.Sprintf("tfctl-%s.md", cmd))
		if err := writeFileIfChanged(tldrPath, []byte(tldr), writeOnlyIfChanged); err != nil {
			fatalf("writing TLDR for %s: %v", cmd, err)
		}

		processed++
	}

	if processed == 0 {
		fatalf("no command markdown found under %s", commandsDir)
	}
}

func fatalf(f string, a ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", a...)
	os.Exit(1)
}

func writeFileIfChanged(path string, new []byte, onlyIfChanged bool) error {
	if !onlyIfChanged {
		return os.WriteFile(path, new, 0o644)
	}
	old, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return os.WriteFile(path, new, 0o644)
		}
		return err
	}
	if bytes.Equal(bytes.TrimSpace(old), bytes.TrimSpace(new)) {
		return nil
	}
	return os.WriteFile(path, new, 0o644)
}

var (
	h1Re = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	// sectionRe = regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z\s]+)\n+`)
)

func extractTitleAndShortDesc(md string) (title, short string) {
	// Title from first H1
	if m := h1Re.FindStringSubmatch(md); m != nil {
		title = strings.TrimSpace(m[1])
	}
	// Find "Short description" section and take the next non-empty paragraph
	idx := strings.Index(strings.ToLower(md), "short description")
	if idx >= 0 {
		rest := md[idx:]
		// Skip the header line
		if nl := strings.Index(rest, "\n"); nl >= 0 {
			rest = rest[nl+1:]
		}
		// Take the next non-empty line(s) until blank
		lines := strings.Split(rest, "\n")
		var b strings.Builder
		for _, ln := range lines {
			if strings.TrimSpace(ln) == "" {
				if b.Len() > 0 { // stop after first paragraph
					break
				}
				continue
			}
			// stop if we hit another section header
			if strings.TrimSpace(ln) == "Flags and related docs" || strings.HasPrefix(ln, "#") || strings.HasSuffix(ln, ":") {
				break
			}
			b.WriteString(strings.TrimSpace(ln))
			b.WriteString(" ")
		}
		short = strings.TrimSpace(b.String())
	}
	if short == "" {
		// Fallback to a generic sentence using title
		if title != "" {
			short = fmt.Sprintf("%s.", title)
		}
	}
	return
}

type example struct {
	Desc string
	Cmd  string
}

func extractQuickExamples(md string) []example {
	// Find the "Quick examples" section; capture the first fenced code block after it
	lower := strings.ToLower(md)
	idx := strings.Index(lower, "quick examples")
	if idx < 0 {
		return nil
	}
	rest := md[idx:]
	// Find first code fence after the header
	fence := "```"
	fenceStart := strings.Index(rest, fence)
	if fenceStart < 0 {
		return nil
	}
	rest = rest[fenceStart+len(fence):]
	fenceEnd := strings.Index(rest, fence)
	if fenceEnd < 0 {
		return nil
	}
	code := rest[:fenceEnd]
	lines := strings.Split(code, "\n")
	var exs []example
	var cur example
	for _, ln := range lines {
		s := strings.TrimRight(ln, "\r")
		if strings.TrimSpace(s) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(s), "# ") || strings.HasPrefix(strings.TrimSpace(s), "#\t") || strings.HasPrefix(strings.TrimSpace(s), "#") {
			// Start a new description; if cur has both, push and reset
			if cur.Desc != "" && cur.Cmd != "" {
				exs = append(exs, cur)
				cur = example{}
			}
			cur.Desc = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "# "), "#"))
			continue
		}
		// Treat as command line
		if cur.Cmd == "" {
			cur.Cmd = strings.TrimSpace(s)
			if cur.Desc == "" {
				// Provide a generic description if missing
				cur.Desc = "Example"
			}
			exs = append(exs, cur)
			cur = example{}
		}
	}
	// If leftover cur is complete, append
	if cur.Desc != "" && cur.Cmd != "" {
		exs = append(exs, cur)
	}
	return exs
}

func buildTLDR(cmd, title, short string, exs []example) string {
	var b strings.Builder
	// Header
	b.WriteString("# tfctl-" + cmd + "\n\n")
	if short != "" {
		b.WriteString("> " + short + "\n")
	} else if title != "" {
		b.WriteString("> " + title + "\n")
	} else {
		b.WriteString("> tfctl " + cmd + "\n")
	}
	b.WriteString("> More information: https://github.com/staranto/tfctlgo.\n\n")

	if len(exs) == 0 {
		// Fallback examples
		b.WriteString("- Show help for the command:\n\n")
		b.WriteString("`tfctl " + cmd + " --help`\n")
		b.WriteString("\n")
		return b.String()
	}

	for i, ex := range exs {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- " + strings.TrimSpace(ex.Desc) + ":\n\n")
		// Ensure backticks and placeholder style
		b.WriteString("`" + sanitizeCommand(ex.Cmd) + "`\n")
	}
	return b.String()
}

func sanitizeCommand(s string) string {
	// Replace angle-bracket placeholders with {{...}} if present
	// For now, just compress runs of whitespace
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
