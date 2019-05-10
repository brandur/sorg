package modulir

import (
	"fmt"
	"net/http"
	"path"

	"github.com/brandur/modulir/context"
	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

func serveTargetDirHTTP(c *context.Context) error {
	c.Log.Infof("Serving '%s' to: http://localhost:%v/", path.Clean(c.TargetDir), c.Port)

	handler := http.FileServer(http.Dir(c.TargetDir))

	err := http.ListenAndServe(fmt.Sprintf(":%v", c.Port), handler)
	if err != nil {
		return errors.Wrap(err, "Error starting HTTP server")
	}
	return nil
}

