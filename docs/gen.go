// Documentation generator.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/airforce270/airbot/commands"
	"github.com/airforce270/airbot/commands/basecommand"
)

const (
	generatedFileWarning       = "[//]: # ( !!! DO NOT EDIT MANUALLY !!!  This is a generated file, any changes will be overwritten! )"
	disableMarkdownLintMessage = "<!-- markdownlint-disable line-length -->"
	generatedFileMessage       = generatedFileWarning + "\n\n" + disableMarkdownLintMessage + "\n\n"
	commandsFilePath           = "commands.md"
)

var (
	//go:embed commands.md.gtpl
	commandsTmplData string

	files = []fileToGenerate{
		{
			Template: template.Must(template.New(commandsFilePath).Funcs(funcs).Parse(commandsTmplData)),
			Path:     commandsFilePath,
			Data:     commands.CommandGroups,
		},
	}

	funcs = map[string]any{
		"formatAliases": func(strs []string) string {
			var joined []string
			for _, str := range strs {
				joined = append(joined, fmt.Sprintf("`$%s`", str))
			}
			return strings.Join(joined, ", ")
		},
		"formatUsage": func(command basecommand.Command) string {
			return command.Usage("$")
		},
	}
)

type fileToGenerate struct {
	Template *template.Template
	Path     string
	Data     any
}

func gen(tmpl *template.Template, fileName string, data any) error {
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("tmpl.Execute with data %+v failed: %w", data, err)
	}
	return os.WriteFile(fileName, []byte(generatedFileMessage+buf.String()), 0666)
}

func main() {
	var errs []error

	for _, file := range files {
		err := gen(file.Template, file.Path, file.Data)
		if err != nil {
			log.Printf("Failed to generate %s: %v", file.Path, err)
		}
		errs = append(errs, err)
	}

	if err := errors.Join(errs...); err != nil {
		fmt.Printf("Errors occurred while generating docs: %v", err)
		os.Exit(1)
	}
}
