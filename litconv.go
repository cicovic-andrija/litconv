package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	TemplateName         = "my-library"
	OutputFile           = TemplateName + ".md"
	TemplateMarkdownText = `---
tags: [literature]
title: {{ .Title }}
type: reference
comment: 'This document is auto-generated from the master blob using https://github.com/cicovic-andrija/litconv.'
created: '2023-11-01T12:34:00.000Z'
modified: '{{ .ModifiedTimeUTC }}'
---

# {{ .Title }}

This document lists all the books in my personal library.

## Contents

1. [Novels](#novels)
2. [Short Stories](#short-stories)
3. [Textbooks](#textbooks)
4. [Other](#other)

## 1. Novels
{{ range .Novels }}
- **_{{ .Title }}_**, {{ .Author }} {{ .SymbolsWithMarkup }}{{ end }}

## 2. Short Stories
{{ range .ShortStories }}
- **_{{ .Title }}_**, {{ .Author }} {{ .SymbolsWithMarkup }}{{ end }}

## 3. Textbooks
{{ range .Textbooks }}
- **_{{ .Title }}_**, {{ .Author }} {{ .SymbolsWithMarkup }}{{ end }}

## 4. Other
{{ range .Other }}
- **_{{ .Title }}_**, {{ .Author }} {{ .SymbolsWithMarkup }}{{ end }}
`
)

type TemplateData struct {
	Title           string
	ModifiedTimeUTC string
	Novels          []BookData
	ShortStories    []BookData
	Textbooks       []BookData
	Other           []BookData
}

type BookData struct {
	Title             string
	Author            string
	SymbolsWithMarkup string
}

func bookFromRecord(rec []string) BookData {
	symbolsWithMarkup := fmt.Sprintf("`%s`", rec[3])
	if rec[4] != "" {
		symbolsWithMarkup = fmt.Sprintf("`%s%s`", rec[3], rec[4])
	}

	return BookData{
		Title:             rec[1],
		Author:            rec[2],
		SymbolsWithMarkup: symbolsWithMarkup,
	}
}

// Convert dive data in CSV to Markdown.
func main() {
	var inputFile string
	flag.Parse()
	if inputFile = flag.Arg(0); inputFile == "" {
		panic("input file not provided")
	}

	fh, err := os.Open(inputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open %s: %v", inputFile, err))
	}
	defer fh.Close()

	csvreader := csv.NewReader(fh)
	records, err := csvreader.ReadAll()
	if err != nil {
		panic(fmt.Errorf("failed to read CSV data from %s: %v", inputFile, err))
	}

	novels := make([]BookData, 0)
	shortStories := make([]BookData, 0)
	textbooks := make([]BookData, 0)
	other := make([]BookData, 0)

	// Note: This is a custom conversion program, it assumes that input file is
	// valid in every sense of that word.
	for _, rec := range records {
		if rec[1] == "" || strings.HasPrefix(rec[4], "-") {
			continue
		}
		switch rec[5] {
		case "Novel":
			novels = append(novels, bookFromRecord(rec))
		case "Short Stories":
			shortStories = append(shortStories, bookFromRecord(rec))
		case "Textbook":
			textbooks = append(textbooks, bookFromRecord(rec))
		case "Other":
			other = append(other, bookFromRecord(rec))
		default:
			continue
		}
	}

	t := template.Must(template.New(TemplateName).Parse(TemplateMarkdownText))
	outFd, err := os.Create(OutputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open output file %s for writing: %v", OutputFile, err))
	}
	data := &TemplateData{
		Title:           "My Library",
		ModifiedTimeUTC: time.Now().UTC().Format(time.RFC3339),
		Novels:          novels,
		ShortStories:    shortStories,
		Textbooks:       textbooks,
		Other:           other,
	}
	t.Execute(outFd, data)
}
