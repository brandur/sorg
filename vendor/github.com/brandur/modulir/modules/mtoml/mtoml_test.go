package mtoml

import (
	"os"
	"testing"

	"github.com/brandur/modulir/modules/mtesting"
	assert "github.com/stretchr/testify/require"
)

func TestSplitFrontmatter(t *testing.T) {
	type testStruct struct {
		Foo string `toml:"foo"`
	}

	c := mtesting.NewContext()

	{
		path := mtesting.WriteTempFile(t, []byte(`+++
foo = "bar"
+++

other`))
		defer os.Remove(path)

		var v testStruct
		content, err := ParseFileFrontmatter(c, path, &v)
		assert.NoError(t, err)
		assert.Equal(t, "bar", v.Foo)
		assert.Equal(t, []byte("other"), content)
	}

	{
		path := mtesting.WriteTempFile(t, []byte("other"))
		defer os.Remove(path)

		var v testStruct
		content, err := ParseFileFrontmatter(c, path, &v)
		assert.NoError(t, err)
		assert.Equal(t, "", v.Foo)
		assert.Equal(t, []byte("other"), content)
	}

	{
		path := mtesting.WriteTempFile(t, []byte(`+++
foo = "bar"
+++
`))
		defer os.Remove(path)

		var v testStruct
		content, err := ParseFileFrontmatter(c, path, &v)
		assert.NoError(t, err)
		assert.Equal(t, "bar", v.Foo)
		assert.Equal(t, []byte(nil), content)
	}

	{
		path := mtesting.WriteTempFile(t, []byte(`foo = "bar"
foo = "bar"
+++
`))
		defer os.Remove(path)

		var v testStruct
		content, err := ParseFileFrontmatter(c, path, &v)
		assert.Equal(t, errBadFrontmatter, err)
		assert.Equal(t, []byte(nil), content)
	}
}
