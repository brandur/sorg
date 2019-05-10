package mtoml

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/brandur/modulir"
	"github.com/pkg/errors"
	"github.com/pelletier/go-toml"
)

func ParseFile(c *modulir.Context, source string, data interface{}) error {
	raw, err := ioutil.ReadFile(source)
	if err != nil {
		return errors.Wrap(err, "Error reading file")
	}

	err = toml.Unmarshal(raw, data)
	if err != nil {
		return errors.Wrap(err, "Error unmarshaling TOML")
	}

	c.Log.Debugf("mtoml: Parsed file: %s", source)
	return nil
}

func ParseFileFrontmatter(c *modulir.Context, source string, data interface{}) ([]byte, error) {
	raw, err := ioutil.ReadFile(source)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading file")
	}

	frontmatter, content, err := splitFrontmatter(string(raw))
	if err != nil {
		return nil, errors.Wrap(err, "Error splitting frontmatter")
	}

	err = toml.Unmarshal([]byte(frontmatter), data)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling TOML frontmatter")
	}

	c.Log.Debugf("mtoml: Parsed file frontmatter: %s", source)
	return []byte(content), nil
}

//
// Private
//

var errBadFrontmatter = fmt.Errorf("Unable to split TOML frontmatter")

var splitFrontmatterRE = regexp.MustCompile("(?m)^+{3}")

func splitFrontmatter(content string) (string, string, error) {
	parts := splitFrontmatterRE.Split(content, 3)

	if len(parts) > 1 && parts[0] != "" {
		return "", "", errBadFrontmatter
	} else if len(parts) == 2 {
		return "", strings.TrimSpace(parts[1]), nil
	} else if len(parts) == 3 {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
	}

	return "", strings.TrimSpace(parts[0]), nil
}
