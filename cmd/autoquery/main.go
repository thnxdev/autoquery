package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"
)

var (
	version = "dev"
	cli     struct {
		OutDir  string           `help:"Destination directory for the sql files." optional:""`
		Pkg     string           `arg:"" help:"Package to scan for autoquery comments." default:"."`
		Version kong.VersionFlag `help:"Show version."`
	}
)

var directiveRe = regexp.MustCompile(`^autoquery\s+(.*)$`)

func main() {
	kctx := kong.Parse(&cli, kong.Vars{"version": version}, kong.Description(`autoquery is a tool for generating sqlc queries from Go comments

Example:

	/* autoquery name: GetChargesSyncable :many

	SELECT ...
	*/`))

	var root = cli.OutDir
	if root == "" {
		root = findGitRoot()
		if root == "" {
			kctx.Fatalf("could not find git root")
		}
		root = filepath.Join(root, "database", "queries")
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, cli.Pkg)
	kctx.FatalIfErrorf(err)
	files := map[string]*bytes.Buffer{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, comment := range file.Comments {
				text := strings.TrimSpace(comment.Text())
				lines := strings.Split(text, "\n")
				if len(lines) == 0 {
					continue
				}
				directive := strings.TrimSpace(lines[0])
				groups := directiveRe.FindStringSubmatch(directive)
				if groups == nil {
					continue
				}
				lines = dedent(lines[1:])
				sqlcDirective := groups[1]
				dest := filepath.Join(root, pkg.Name+".sql")
				w, ok := files[dest]
				if !ok {
					w = &bytes.Buffer{}
					files[dest] = w
				}
				fmt.Fprintf(w, "-- %s\n%s\n\n", sqlcDirective, strings.Join(lines, "\n"))
			}
		}
	}
	for dest, sqlc := range files {
		err = os.WriteFile(dest, sqlc.Bytes(), 0600)
		kctx.FatalIfErrorf(err)
	}
}

func dedent(lines []string) []string {
	// Find the minimum indentation.
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	// Dedent.
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return lines
}

// Find path to git root.
func findGitRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if dir == "/" {
			return ""
		}
		dir = filepath.Dir(dir)
	}
}
