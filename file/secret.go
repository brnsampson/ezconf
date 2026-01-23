package file

import (
	"io/fs"
	"os"

	"github.com/brnsampson/optional"
)

const (
	SecretFilePermsInclude fs.FileMode = 0600
	SecretFilePermsExclude fs.FileMode = 0177
)

type SecretFile struct {
	File
}

// MakeSecret takes a pointer to an existing File and returns a SecretFile
// then clears the original so that it cannot be used and accidentally expose
// the sercret.
func MakeSecret(f *File) SecretFile {
	var sf SecretFile
	path, ok := f.Get()
	if ok {
		sf = SecretFile{SomeFile(path)}
		f.Clear()
	}
	return sf
}

func SomeSecretFile(path string) SecretFile {
	return SecretFile{SomeFile(path)}
}

func NoSecretFile() SecretFile {
	return SecretFile{NoFile()}
}

// Override the Type() method from the inner value. Part of the flag.Value interface.
func (o SecretFile) Type() string {
	return "SecretFile"
}

// Override the String() method from the inner value just so we return the correct None[Type] string.
func (o SecretFile) String() string {
	if o.IsNone() {
		return "None[SecretFile]"
	} else {
		tmp, ok := o.Get()
		if !ok {
			return "Error[SecretFile]"
		}
		return tmp
	}
}

func (o SecretFile) Abs() (opt SecretFile, err error) {
	tmp, err := o.File.Abs()
	return SecretFile{tmp}, err
}

func (o SecretFile) FilePermsValid() (bool, error) {
	return o.File.FilePermsValid(SecretFilePermsInclude, SecretFilePermsExclude)
}

func (o SecretFile) OpenFile(flag int) (*os.File, error) {
	return o.File.OpenFile(flag, SecretFilePermsInclude)
}

func (o SecretFile) ReadFile() (secret optional.Secret, ok bool) {
	str, ok := o.File.ReadFile()
	secret = optional.MakeSecret(&str)
	return secret, ok
}
