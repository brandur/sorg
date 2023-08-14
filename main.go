package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeshaw/envdecode"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mimage"
	"github.com/brandur/sorg/modules/scommon"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Main
//
//
//
//////////////////////////////////////////////////////////////////////////////

func main() {
	rootCmd := &cobra.Command{
		Use:   "sorg",
		Short: "Sorg is a static site generator",
		Long: strings.TrimSpace(`
Sorg is a static site generator for Brandur's personal
homepage and some of its adjacent functions. See the product
in action at https://brandur.org.`),
	}

	buildCommand := &cobra.Command{
		Use:   "build",
		Short: "Run a single build loop",
		Long: strings.TrimSpace(`
Starts the build loop that watches for local changes and runs
when they're detected. A webserver is started on PORT (default
5002).`),
		Run: func(cmd *cobra.Command, args []string) {
			modulir.Build(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(buildCommand)

	loopCommand := &cobra.Command{
		Use:   "loop",
		Short: "Start build and serve loop",
		Long: strings.TrimSpace(`
Runs the build loop one time and places the result in TARGET_DIR
(default ./public/).`),
		Run: func(cmd *cobra.Command, args []string) {
			modulir.BuildLoop(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(loopCommand)

	var live bool
	var staging bool
	sendCommand := &cobra.Command{
		Use:   "send [source newsletter .md file]",
		Short: "Email a Nanoglyph or Passages newsletter",
		Long: strings.TrimSpace(`
Emails the Nanoglyph or Passages newsletter at the location given
as argument. Note that MAILGUN_API_KEY must be set in the
environment for this to work as it executes against the Mailgun
API.`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := &modulir.Context{Log: getLog()}
			sendNewsletter(c, args[0], live, staging)
		},
	}
	sendCommand.Flags().BoolVar(&live, "live", false,
		"Send to list (as opposed to dry run)")
	sendCommand.Flags().BoolVar(&staging, "staging", false,
		"Send to staging list (as opposed to dry run)")
	rootCmd.AddCommand(sendCommand)

	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}

	mimage.MagickBin = conf.MagickBin
	mimage.MozJPEGBin = conf.MozJPEGBin
	mimage.PNGQuantBin = conf.PNGQuantBin
	mimage.TempDir = scommon.TempDir

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v", err)
		os.Exit(1)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Conf contains configuration information for the command. It's extracted from
// environment variables.
type Conf struct {
	// AbsoluteURL is the absolute URL where the compiled site will be hosted.
	// It's used for things like Atom feeds and sending email.
	AbsoluteURL string `env:"ABSOLUTE_URL,default=https://brandur.org"`

	// BlackSwanDatabaseURL is a connection string for a database to connect to
	// in order to extract books, tweets, runs, etc.
	BlackSwanDatabaseURL string `env:"BLACK_SWAN_DATABASE_URL"`

	// Concurrency is the number of build Goroutines that will be used to
	// perform build work items.
	Concurrency int `env:"CONCURRENCY,default=30"`

	// Drafts is whether drafts of articles and fragments should be compiled
	// along with their published versions.
	//
	// Activating drafts also prompts the creation of a robots.txt to make sure
	// that drafts aren't inadvertently accessed by web crawlers.
	Drafts bool `env:"DRAFTS,default=false"`

	// EnableGoatCounter enables Goat Counter analytics, which is a less
	// invasive version of Google Analytics that collects less information,
	// does less user tracking, and is not an ad company. Currently I'm
	// experimenting with it as a possibility of a full alternative.
	EnableGoatCounter bool `env:"ENABLE_GOAT_COUNTER,default=false"`

	// GoogleAnalyticsID is the account identifier for Google Analytics to use.
	GoogleAnalyticsID string `env:"GOOGLE_ANALYTICS_ID"`

	// LocalFonts starts using locally downloaded versions of Google Fonts.
	// This is not ideal for real deployment because you won't be able to
	// leverage Google's CDN and the caching that goes with it, and may not get
	// the font format for requesting browsers, but good for airplane rides
	// where you otherwise wouldn't have the fonts.
	LocalFonts bool `env:"LOCAL_FONTS,default=false"`

	// MailgunAPIKey is a key for Mailgun used to send email. It's required
	// when using the `passages` command.
	MailgunAPIKey string `env:"MAILGUN_API_KEY"`

	// MagickBin is the location of the `magick` binary that ships with the
	// ImageMagick project (an image manipulation utility).
	MagickBin string `env:"MAGICK_BIN"`

	// MozJPEGBin is the location of the `cjpeg` binary that ships with the
	// mozjpeg project (a JPG optimizer). If configured, Sorg will put photos
	// through an optimization pass after resizing them.
	MozJPEGBin string `env:"MOZJPEG_BIN"`

	// NumAtomEntries is the number of entries to put in Atom feeds.
	NumAtomEntries int `env:"NUM_ATOM_ENTRIES,default=20"`

	// PNGQuantBin is the location of the `pnqquant` binary (a PNG optimizer). If
	// configured, PNGs are passed through an optimization pass after resizing
	// them.
	PNGQuantBin string `env:"PNGQUANT_BIN"`

	// Port is the port on which to serve HTTP when looping in development.
	Port int `env:"PORT,default=5002"`

	// SorgEnv is the environment to run the app with. Use "development" to
	// activate development features.
	SorgEnv string `env:"SORG_ENV,default=production"`

	// TargetDir is the target location where the site will be built to.
	TargetDir string `env:"TARGET_DIR,default=./public"`

	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	sorgEnvDevelopment = "development"
)

func getLog() *logrus.Logger {
	log := logrus.New()

	if conf.Verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}

// getModulirConfig interprets Conf to produce a configuration suitable to pass
// to a Modulir build loop.
func getModulirConfig() *modulir.Config {
	return &modulir.Config{
		Concurrency: conf.Concurrency,
		Log:         getLog(),
		LogColor:    term.IsTerminal(int(os.Stdout.Fd())),
		Port:        conf.Port,
		SourceDir:   ".",
		TargetDir:   conf.TargetDir,
		Websocket:   conf.SorgEnv == sorgEnvDevelopment,
	}
}
