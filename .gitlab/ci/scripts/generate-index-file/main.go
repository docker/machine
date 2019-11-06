package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type entry struct {
	FullPath     string
	RelativePath string
	Checksum     string
	SizeMb       float64
}

type release struct {
	Name      string
	Project   string
	SourceURL string
	Revision  string
	Ref       string
	CreatedAt string
	Files     []entry
}

var (
	indexFile     string
	checksumsFile string
)

func main() {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: %s scan_dir version ref revision\n", os.Args[0])
		os.Exit(1)
	}

	scanDir := strings.Trim(os.Args[1], "/")
	version := os.Args[2]
	ref := os.Args[3]
	revision := os.Args[4]

	indexFile = filepath.Join(scanDir, "index.html")
	checksumsFile = filepath.Join(scanDir, "release.sha256")

	files, err := filepath.Glob(fmt.Sprintf("%s/*", scanDir))
	if err != nil {
		log.Fatalf("Failed to scan dir %q for files: %v", scanDir, err)
	}

	entries := make([]entry, 0)
	for _, file := range files {
		entries = append(entries, getFileEntry(scanDir, file))
	}

	checksumFileEntry := createChecksumsFile(scanDir, checksumsFile, entries)
	entries = append(entries, checksumFileEntry)

	releaseInfo := release{
		Name:      version,
		Project:   "Docker Machine (GitLab's fork)",
		SourceURL: fmt.Sprintf("%s/tree/%s", os.Getenv("CI_PROJECT_URL"), ref),
		Revision:  revision,
		Ref:       ref,
		CreatedAt: time.Now().Format(time.RFC3339),
		Files:     entries,
	}

	tpl, err := template.New("release").Parse(indexTemplate)
	if err != nil {
		log.Fatalf("Failed to parse the template: %v", err)
	}

	buf := new(bytes.Buffer)

	err = tpl.Execute(buf, releaseInfo)
	if err != nil {
		log.Fatalf("Failed to parse the template: %v", err)
	}

	err = ioutil.WriteFile(indexFile, buf.Bytes(), 0600)
	if err != nil {
		log.Fatalf("Failed to write to the index file %q: %v", indexFile, err)
	}
}

func getFileEntry(scanDir string, file string) entry {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Failed to open the file %q: %v", file, err)
	}
	defer f.Close()

	hasher := sha256.New()

	_, err = io.Copy(hasher, f)
	if err != nil {
		log.Fatalf("Failed to copy file's %q content to the SHA256 hash calculator: %v", file, err)
	}

	fileInfo, err := os.Stat(file)
	if err != nil {
		log.Fatalf("Failed to stat the file %q: %v", file, err)
	}

	return entry{
		FullPath:     file,
		RelativePath: strings.Replace(file, fmt.Sprintf("%s/", scanDir), "", -1),
		Checksum:     fmt.Sprintf("%x", hasher.Sum(nil)),
		SizeMb:       float64(fileInfo.Size()) / 1048576,
	}
}

func createChecksumsFile(scanDir string, file string, entries []entry) entry {
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Failed to open the file %q: %v", file, err)
	}
	defer f.Close()

	for _, fileEntry := range entries {
		if fileEntry.FullPath == indexFile || fileEntry.FullPath == checksumsFile {
			continue
		}

		_, err := fmt.Fprintf(f, "%s\t%s\n", fileEntry.Checksum, fileEntry.RelativePath)
		if err != nil {
			log.Fatalf("Failed to write to  file %q: %v", file, err)
		}
	}

	return getFileEntry(scanDir, file)
}

var indexTemplate = `
{{ $title := (print .Project " :: " .Name) }}

<html>
    <head>
        <meta charset="utf-8/">
        <title>{{ $title }}</title>
        <style type="text/css">
            body {font-family: monospace; font-size: 14px; margin: 40px; padding: 0;}
            h1 {border-bottom: 1px solid #aaa; padding: 10px;}
            p {font-size: 0.85em; margin: 5px 10px;}
            p span {display: inline-block; font-weight: bold; width: 100px;}
            p a {color: #000; font-weight: bold; text-decoration: none;}
            p a:hover {text-decoration: underline;}
            ul {background: #eee; border: 1px solid #aaa; border-radius: 3px; box-shadow: 0 0 5px #aaa inset; list-style-type: none; margin: 10px 0; padding: 10px;}
            li {margin: 5px 0; padding: 5px; font-size: 12px;}
            li:hover {background: #dedede;}
            .file_name {display: inline-block; font-weight: bold; width: calc(100% - 610px);}
            .file_name a {color: #000; display: inline-block; text-decoration: none; width: calc(100% - 10px);}
            .file_checksum {display: inline-block; text-align: right; width: 500px;}
            .file_size {display: inline-block; text-align: right; width: 90px;}
        </style>
    </head>
    <body>
        <h1>{{ $title }}</h1>

        <p><span>Sources:</span> <a href="{{ .SourceURL }}" target="_blank">{{ .SourceURL }}</a></p>
        <p><span>Revision:</span> {{ .Revision }}</p>
        <p><span>Ref:</span> {{ .Ref }}</p>
        <p><span>Created at:</span> {{ .CreatedAt }}</p>

        <ul>
        {{ range $_, $file := .Files }}
            <li>
                <span class="file_name"><a href="./{{ $file.RelativePath }}">{{ $file.RelativePath }}</a></span>
                <span class="file_checksum">{{ $file.Checksum }}</span>
                <span class="file_size">{{ printf "%2.2f" $file.SizeMb }} MiB</span>
            </li>
        {{ end }}
        </ul>
    </body>
</html>
`
