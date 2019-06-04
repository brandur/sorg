package mfile

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/brandur/modulir"
	"github.com/pkg/errors"
)

func CopyFile(c *modulir.Context, source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return errors.Wrap(err, "Error opening copy source")
	}
	defer in.Close()

	out, err := os.Create(target)
	if err != nil {
		return errors.Wrap(err, "Error creating copy target")
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrap(err, "Error copying data")
	}

	c.Log.Debugf("mfile: Copied '%s' to '%s'", source, target)
	return nil
}

func CopyFileToDir(c *modulir.Context, source, targetDir string) error {
	return CopyFile(c, source, path.Join(targetDir, filepath.Base(source)))
}

func EnsureDir(c *modulir.Context, target string) error {
	err := os.MkdirAll(target, 0755)
	if err != nil {
		return errors.Wrap(err, "Error creating directory")
	}

	c.Log.Debugf("mfile: Ensured dir existence: %s", target)
	return nil
}

func EnsureSymlink(c *modulir.Context, source, target string) error {
	c.Log.Debugf("Checking symbolic link (%v): %v -> %v",
		path.Base(source), source, target)

	var actual string

	_, err := os.Stat(target)

	// Note that if a symlink file does exist, but points to a non-existent
	// location, we still get an "does not exist" error back, so we fall down
	// to the general create path so that the symlink file can be removed.
	//
	// The call to RemoveAll does not affect the other path of the symlink file
	// not being present because it doesn't care whether or not the file it's
	// trying remove is actually there.
	if os.IsNotExist(err) {
		c.Log.Debugf("Destination link does not exist. Creating.")
		goto create
	}
	if err != nil {
		return errors.Wrap(err, "Error checking symlink")
	}

	actual, err = os.Readlink(target)
	if err != nil {
		return errors.Wrap(err, "Error reading symlink")
	}

	if actual == source {
		c.Log.Debugf("Link exists.")
		return nil
	}

	c.Log.Debugf("Destination links to wrong source. Creating.")

create:
	err = os.RemoveAll(target)
	if err != nil {
		return errors.Wrap(err, "Error removing symlink")
	}

	source, err = filepath.Abs(source)
	if err != nil {
		return err
	}

	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}

	err = os.Symlink(source, target)
	if err != nil {
		return errors.Wrap(err, "Error creating symlink")
	}

	return nil
}

func IsBackup(base string) bool {
	return strings.HasSuffix(base, "~")
}

func IsHidden(base string) bool {
	return strings.HasPrefix(base, ".")
}

func IsMeta(base string) bool {
	return strings.HasPrefix(base, "_")
}

// Exists is a shortcut to check if a file exists. It panics if encountering an
// unexpected error.
func Exists(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

// MustAbs is a shortcut variant of filepath.Abs which panics instead of
// returning an error.
func MustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return absPath
}

//
// ReadDir
//

// ReadDir reads files in a directory and returns a list of file paths.
//
// Unlike ioutil.ReadDir, this function skips hidden, "meta" (i.e. prefixed by
// an underscore), and Vim backup (i.e. suffixed with a tilde) files, and
// returns a list of full paths (easier to plumb into other functions), and
// sets up a watch on the listed source.
func ReadDir(c *modulir.Context, source string) ([]string, error) {
	return ReadDirWithOptions(c, source, nil)
}

// ReadDirOptions are options for ReadDirWithOptions.
type ReadDirOptions struct {
	// ShowBackup tells the function to not skip backup files like those
	// produced by Vim. These are suffixed with a tilde '~'.
	ShowBackup bool

	// ShowDirs tell the function not to skip directories.
	ShowDirs bool

	// ShowHidden tells the function to not skip hidden files (prefixed with a
	// dot '.').
	ShowHidden bool

	// ShowMeta tells the function to not skip so-called "meta" files
	// (prefixed with an underscore '_').
	ShowMeta bool
}

// ReadDirWithOptions reads files in a directory and returns a list of file
// paths.
//
// Unlike ReadDir, its behavior can be tweaked.
func ReadDirWithOptions(c *modulir.Context, source string,
	opts *ReadDirOptions) ([]string, error) {

	infos, err := ioutil.ReadDir(source)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading directory")
	}

	var files []string

	for _, info := range infos {
		base := filepath.Base(info.Name())

		if (opts == nil || !opts.ShowBackup) && IsBackup(base) {
			continue
		}

		if (opts == nil || !opts.ShowDirs) && info.IsDir() {
			continue
		}

		if (opts == nil || !opts.ShowHidden) && IsHidden(base) {
			continue
		}

		if (opts == nil || !opts.ShowMeta) && IsMeta(base) {
			continue
		}

		files = append(files, path.Join(source, info.Name()))
	}

	c.Log.Debugf("mfile: Read dir: %s", source)
	return files, nil
}
