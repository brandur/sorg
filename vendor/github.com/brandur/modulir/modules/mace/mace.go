package mace

import (
	"bufio"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/brandur/modulir"
	"github.com/pkg/errors"
	"github.com/yosssi/ace"
)

func Load(c *modulir.Context, basePath, innerPath string, opts *ace.Options) (*template.Template, error) {
	if opts == nil {
		opts = &ace.Options{}
	}

	// Ace made a really strange decision to not take extensions when passing
	// around the names of templates, which makes working with known files
	// unnecessarily difficult. Here we correct that by allowing an extension
	// to be passed to the function and stripping it off for Ace's benefit.
	extlessBasePath := basePath
	if strings.HasSuffix(basePath, ".ace") {
		extlessBasePath = strings.TrimSuffix(basePath, ".ace")
	}
	extlessInnerPath := innerPath
	if strings.HasSuffix(innerPath, ".ace") {
		extlessInnerPath = strings.TrimSuffix(innerPath, ".ace")
	}

	// See the comment above for some context, but since Ace caches templates
	// we always invoke Load and depend on it to take its own fast path if we
	// didn't set DynamicReload.
	template, err := ace.Load(extlessBasePath, extlessInnerPath, opts)
	if err != nil {
		return nil, err
	}

	c.Log.Debugf("mace: Loaded template layout '%s' view '%s'",
		basePath, innerPath)

	return template, nil
}

// Render is a shortcut for loading an Ace template and rendering it to a
// target file.
func Render(c *modulir.Context, basePath, innerPath string, writer io.Writer,
	opts *ace.Options, locals map[string]interface{}) error {

	template, err := Load(c, basePath, innerPath, opts)
	if err != nil {
		return errors.Wrap(err, "Error loading template")
	}

	err = template.Execute(writer, locals)
	if err != nil {
		return errors.Wrap(err, "Error rendering template")
	}

	c.Log.Debugf("mace: Rendered view '%s'", innerPath)
	return nil
}

// RenderFile is a shortcut for loading an Ace template and rendering it to a
// target file.
func RenderFile(c *modulir.Context, basePath, innerPath, target string,
	opts *ace.Options, locals map[string]interface{}) error {

	template, err := Load(c, basePath, innerPath, opts)
	if err != nil {
		return errors.Wrap(err, "Error loading template")
	}

	file, err := os.Create(target)
	if err != nil {
		return errors.Wrap(err, "Error creating target file")
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	err = template.Execute(writer, locals)
	if err != nil {
		return errors.Wrap(err, "Error rendering template")
	}

	c.Log.Debugf("mace: Rendered view '%s' to '%s'", innerPath, target)
	return nil
}
