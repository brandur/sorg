package sassets

import (
	"io/ioutil"
	"testing"

	"github.com/brandur/sorg/modules/stesting"
	assert "github.com/stretchr/testify/require"
)

func TestCompileJavascripts(t *testing.T) {
	dir, err := ioutil.TempDir("", "javascripts")

	file0 := dir + "/.hidden"
	file1 := dir + "/file1.js"
	file2 := dir + "/file2.js"
	file3 := dir + "/file3.js"
	out := dir + "/app.js"

	// This file is hidden and doesn't show up in output.
	err = ioutil.WriteFile(file0, []byte(`hidden`), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file1, []byte(`function() { return "file1" }`), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file2, []byte(`function() { return "file2" }`), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file3, []byte(`function() { return "file3" }`), 0755)
	assert.NoError(t, err)

	err = CompileJavascripts(stesting.NewContext(), dir, out)
	assert.NoError(t, err)

	actual, err := ioutil.ReadFile(out)
	assert.NoError(t, err)

	expected := `/* file1.js */

(function() {

function() { return "file1" }

}).call(this);

/* file2.js */

(function() {

function() { return "file2" }

}).call(this);

/* file3.js */

(function() {

function() { return "file3" }

}).call(this);

`
	assert.Equal(t, expected, string(actual))
}

func TestCompileStylesheets(t *testing.T) {
	dir, err := ioutil.TempDir("", "stylesheets")

	file0 := dir + "/.hidden"
	file1 := dir + "/file1.sass"
	file2 := dir + "/file2.sass"
	file3 := dir + "/file3.css"
	out := dir + "/app.css"

	// This file is hidden and doesn't show up in output.
	err = ioutil.WriteFile(file0, []byte("hidden"), 0755)
	assert.NoError(t, err)

	// The syntax of the first and second files is GCSS and the third is in
	// CSS.
	err = ioutil.WriteFile(file1, []byte("p\n  margin: 10px"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file2, []byte("p\n  padding: 10px"), 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(file3, []byte("p {\n  border: 10px;\n}"), 0755)
	assert.NoError(t, err)

	err = CompileStylesheets(stesting.NewContext(), dir, out)
	assert.NoError(t, err)

	actual, err := ioutil.ReadFile(out)
	assert.NoError(t, err)

	// Note that the first two files have no spacing in the output because they
	// go through the GCSS compiler.
	expected := `/* file1.sass */

p{margin:10px;}

/* file2.sass */

p{padding:10px;}

/* file3.css */

p {
  border: 10px;
}

`
	assert.Equal(t, expected, string(actual))
}
