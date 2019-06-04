package mtoml

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/brandur/modulir"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

func ParseFile(c *modulir.Context, source string, v interface{}) error {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return errors.Wrap(err, "Error reading file")
	}

	err = toml.Unmarshal(data, v)
	if err != nil {
		return errors.Wrap(err, "Error unmarshaling TOML")
	}

	c.Log.Debugf("mtoml: Parsed file: %s", source)
	return nil
}

func ParseFileFrontmatter(c *modulir.Context, source string, v interface{}) ([]byte, error) {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading file")
	}

	frontmatter, content, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(frontmatter, v)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling TOML frontmatter")
	}

	c.Log.Debugf("mtoml: Parsed file frontmatter: %s", source)
	return content, nil
}

//
// Private
//

var errBadFrontmatter = fmt.Errorf("Unable to split TOML frontmatter")

func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	parts := bytes.Split(data, []byte("+++\n"))

	if len(parts) > 1 && !bytes.Equal(parts[0], []byte("")) {
		return nil, nil, errBadFrontmatter
	} else if len(parts) == 2 {
		return nil, bytes.TrimSpace(parts[1]), nil
	} else if len(parts) == 3 {
		return bytes.TrimSpace(parts[1]), bytes.TrimSpace(parts[2]), nil
	}

	return nil, bytes.TrimSpace(parts[0]), nil
}
