package modulir

import (
	"fmt"
	"net/http"
	"path"

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

func startServingTargetDirHTTP(c *Context) *http.Server {
	c.Log.Infof("Serving '%s' to: http://localhost:%v/", path.Clean(c.TargetDir), c.Port)

	handler := http.FileServer(http.Dir(c.TargetDir))

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", c.Port),
		Handler: handler,
	}

	go func() {
		err := server.ListenAndServe()

		// ListenAndServe always returns a non-nil error
		if err != http.ErrServerClosed {
			exitWithError(errors.Wrap(err, "Error starting HTTP server"))
		}
	}()

	return server
}

