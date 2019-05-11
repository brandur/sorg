package mmarkdown

import (
	"io/ioutil"

	"github.com/brandur/modulir"
	"github.com/pkg/errors"
	"gopkg.in/russross/blackfriday.v2"
)

func Render(c *modulir.Context, data []byte) []byte {
	return blackfriday.Run(data)
}

func RenderFile(c *modulir.Context, source, target string) error {
	inData, err := ioutil.ReadFile(source)
	if err != nil {
		return errors.Wrap(err, "Error reading file")
	}

	outData := Render(c, inData)

	err = ioutil.WriteFile(target, outData, 0644)
	if err != nil {
		return errors.Wrap(err, "Error writing file")
	}

	c.Log.Debugf("mmarkdown: Rendered '%s' to '%s'", source, target)
	return nil
}
