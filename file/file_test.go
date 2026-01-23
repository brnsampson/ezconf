package file_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brnsampson/ezconf/file"
	"gotest.tools/v3/assert"
)

func TestFileType(t *testing.T) {
	o := file.SomeFile("/not/a/real/path")
	assert.Equal(t, reflect.TypeOf(o).Name(), o.Type())
	assert.Assert(t, !o.Exists())

	_, err := o.Stat()
	assert.ErrorIs(t, err, os.ErrNotExist)

	valid, err := o.FilePermsValid(0600, 0111)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Assert(t, !valid)

	path := "../testing/rsa/cert.pem"
	o = file.SomeFile(path)
	assert.Assert(t, o.Exists())

	stat, err := o.Stat()
	assert.NilError(t, err)

	// Make sure it is at least readable/writable by root but not executable.
	valid, err = o.FilePermsValid(0600, 0111)
	assert.NilError(t, err)
	assert.Assert(t, valid, "expected file perms 0644 but found %o", stat.Mode())
}

func TestFileGet(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	o := file.SomeFile(path)
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	tmp, ok := o.Get()
	assert.Assert(t, ok)

	assert.Equal(t, path, tmp)
	a, err := o.Abs()
	assert.NilError(t, err)

	tmp, ok = a.Get()
	assert.Assert(t, ok)
	assert.Equal(t, abs, tmp)
}

func TestFileString(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	noneStr := "None[File]"

	o := file.SomeFile(path)
	assert.Equal(t, path, o.String())

	// Test None case displays correctly
	o = file.NoFile()
	assert.Equal(t, noneStr, o.String())
}

func TestFileMarshalText(t *testing.T) {
	path := "../testing/rsa/cert.pem"

	o := file.SomeFile(path)

	s, err := o.MarshalText()
	assert.NilError(t, err)
	assert.Equal(t, path, string(s))
}

func TestFileUnmarshalText(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	nullFile := "null"
	intFile := "42"

	// Text sucessful unmarshaling
	o := file.NoFile()
	err := o.UnmarshalText([]byte(path))
	assert.NilError(t, err)
	assert.Assert(t, o.IsSome())

	ret, ok := o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, path, ret)

	// Test unmarshaling null
	err = o.UnmarshalText([]byte(nullFile))
	assert.NilError(t, err)
	assert.Assert(t, o.IsNone())

	// Test unmarshaling non-path. This will work because it should interpret this as a file named "41", which you could
	// theoretically have.
	err = o.UnmarshalText([]byte(intFile))
	assert.NilError(t, err)

	ret, ok = o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, intFile, ret)
}

func TestFileReadFile(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	badpath := "does/not/exist.txt"
	o := file.SomeFile(path)

	str, ok := o.ReadFile()
	assert.Assert(t, ok)
	assert.Assert(t, str.IsSome())

	// Test with a file that does not exist
	o = file.SomeFile(badpath)

	str, ok = o.ReadFile()
	assert.Assert(t, !ok)
	assert.Assert(t, str.IsNone())
}
