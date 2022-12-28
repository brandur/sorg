package main

import (
	"bufio"
	"context"
	"html/template"
	"io"
	"os"
	"regexp"
	"sync"

	"golang.org/x/xerrors"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mtemplatemd"
	"github.com/brandur/sorg/modules/scommon"
)

//
// TODO: Extract types/functions below this line to something better, probably
// in Modulir.
//

// DependencyRegistry maps Go template sources to other Go template sources that
// have been included in them as dependencies. It's used to know when to trigger
// a rebuild on a file change.
type DependencyRegistry struct {
	// Maps sources to their dependencies.
	sources   map[string][]string
	sourcesMu sync.RWMutex
}

func NewDependencyRegistry() *DependencyRegistry {
	return &DependencyRegistry{
		sources: make(map[string][]string),
	}
}

func (r *DependencyRegistry) getDependencies(source string) []string {
	r.sourcesMu.RLock()
	defer r.sourcesMu.RUnlock()

	return r.sources[source]
}

func (r *DependencyRegistry) parseGoTemplate(baseTmpl *template.Template,
	path string,
) (*template.Template, []string, error) {
	templateData, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, xerrors.Errorf("error reading template file %q: %w", path, err)
	}

	dependencies := []string{path}

	for _, subTemplatePath := range findGoSubTemplates(string(templateData)) {
		newBaseTmpl, subDependencies, err := r.parseGoTemplate(baseTmpl, subTemplatePath)
		if err != nil {
			return nil, nil, err
		}

		dependencies = append(dependencies, subDependencies...)
		baseTmpl = newBaseTmpl
	}

	newBaseTmpl, err := baseTmpl.New(path).Funcs(scommon.HTMLTemplateFuncMap).Parse(string(templateData))
	if err != nil {
		return nil, nil, xerrors.Errorf("error reading parsing template %q: %w", path, err)
	}

	return newBaseTmpl, dependencies, nil
}

func (r *DependencyRegistry) renderGoTemplate(ctx context.Context, c *modulir.Context,
	source, target string, locals map[string]interface{},
) error {
	file, err := os.Create(target)
	if err != nil {
		return xerrors.Errorf("error creating target file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	return r.renderGoTemplateWriter(ctx, c, source, writer, locals)
}

func (r *DependencyRegistry) renderGoTemplateWriter(ctx context.Context, c *modulir.Context,
	source string, writer io.Writer, locals map[string]interface{},
) error {
	ctx, includeMarkdownContainer := mtemplatemd.Context(ctx)

	locals["Ctx"] = ctx

	tmpl, dependencies, err := r.parseGoTemplate(template.New("base_empty"), source)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(writer, locals); err != nil {
		return xerrors.Errorf("error executing template: %w", err)
	}

	r.setDependencies(ctx, c, source, append(dependencies, includeMarkdownContainer.Dependencies...))

	return nil
}

func (r *DependencyRegistry) setDependencies(_ context.Context, c *modulir.Context,
	source string, dependencies []string,
) {
	r.sourcesMu.Lock()
	r.sources[source] = dependencies
	r.sourcesMu.Unlock()

	// Make sure all dependencies are watched.
	c.ChangedAny(dependencies...)
}

var goFileTemplateRE = regexp.MustCompile(`\{\{\-? ?template "([^"]+\.tmpl.html)"`)

func findGoSubTemplates(templateData string) []string {
	subTemplateMatches := goFileTemplateRE.FindAllStringSubmatch(templateData, -1)

	subTemplateNames := make([]string, len(subTemplateMatches))
	for i, match := range subTemplateMatches {
		subTemplateNames[i] = match[1]
	}

	return subTemplateNames
}
