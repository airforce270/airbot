// Documentation generator.
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/airforce270/airbot/commands"

	"github.com/hashicorp/go-multierror"
)

const (
	directory = "docs"

	generatedFileMessage = "[//]: # ( !!! DO NOT EDIT MANUALLY !!!  This is a generated file, any changes will be overwritten! )\n\n"
)

var (
	//go:embed "commands.md.gtpl"
	commandsTmplData string
	commandsTmpl     = template.Must(template.New(commandsFilePath).Funcs(funcs).Parse(commandsTmplData))
	commandsFilePath = path.Join(directory, "commands.md")

	files = map[string]any{
		commandsFilePath: commands.CommandGroups,
	}

	funcs = map[string]any{
		"formatAlternateNames": func(strs []string) string {
			var joined []string
			for _, str := range strs {
				joined = append(joined, fmt.Sprintf("`$%s`", str))
			}
			return strings.Join(joined, ", ")
		},
	}
)

func gen(fileName string, data any) error {
	var buf bytes.Buffer
	if err := commandsTmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(fileName, []byte(generatedFileMessage+buf.String()), 0666)
}

func main() {
	var errs *multierror.Error

	for file, data := range files {
		err := gen(file, data)
		if err != nil {
			log.Printf("Failed to generate %s: %v", file, err)
		}
		errs = multierror.Append(errs, err)
	}

	if errs.ErrorOrNil() != nil {
		fmt.Printf("Errors occurred while generating docs: %v", errs.ErrorOrNil())
		os.Exit(1)
	}
}
