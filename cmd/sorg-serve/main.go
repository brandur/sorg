package main

import (
	"net/http"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/brandur/sorg"
	"github.com/joeshaw/envdecode"
)

// Conf contains configuration information for the command.
type Conf struct {
	// Port is the port on which the command will serve the site over HTTP.
	Port int `env:"PORT,default=5001"`

	// TargetDir is the target location where the site was built to.
	TargetDir string `env:"TARGET_DIR,default=./public"`
}

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

func main() {
	sorg.InitLog(false)

	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	err = sorg.CreateOutputDirs(conf.TargetDir)
	if err != nil {
		log.Fatal(err)
	}

	err = serve(conf.TargetDir, conf.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func serve(targetDir string, port int) error {
	log.Infof("Serving '%v' on port %v", path.Clean(targetDir), port)
	log.Infof("Open browser to: http://localhost:%v/", port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
