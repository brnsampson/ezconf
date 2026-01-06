package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/brnsampson/optional/file"
)

func LoadingAndReadingFiles(path *string) error {
	// However we got it, we either have or do not have a path. For our example, let's assume we loaded this from a
	// flag so we end up with a *string which could be nil

	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Just read the contents of a file. Acts like os.ReadFile(path), but returns an optional Str
	// containing the contents as a string if the file had any contents, // or a None if it was empty.
	// If there was some kind of error (e.g. the file does not exist or is not readable), then the
	// second ok value will be false. This is set up this was so that you can just call
	// contents := f.ReadFile() and use the value without really thinking about it.
	contents, ok := f.ReadFile()
	if !ok {
		fmt.Println("Failed to read from file")
	} else {
		fmt.Println("Got file contents: ", contents)
	}

	// Open a file for reading. File.Open() works just like os.Open(path),
	// so the file is opend in ReadOnly mode.
	var opened *os.File
	opened, err := f.Open()
	if err != nil {
		fmt.Println("Failed to open file for reading: ", err)
		return err
	}
	defer opened.Close()

	// Now use the file handle exactly as you would if you called os.Open()
	return nil
}

func SecretFiles(path *string) error {
	// There is a SecretFile type for convenience since this is a common thing to do
	// in an application. SecretFile simple overrides a few methods of File so that
	// we get a Secret option out of loading the contents instead of a Str.
	f := file.NoSecretFile()
	if path != nil {
		f = file.SomeSecretFile(*path)
	}

	// You can also upgrade a File to a SecretFile
	var normf file.File
	if path != nil {
		normf = file.SomeFile(*path)
	}
	secretf := file.MakeSecret(&normf)

	// You can still see the filepath and everything for a secret file, but we
	// do assume some things about secret files such as the premissions allowed.
	valid, err := secretf.FilePermsValid()
	if err != nil {
		fmt.Println("Failed when validating file permissions for secretf!")
		return err
	}

	if !valid {
		fmt.Println("File permissions for a SecretFile were not 0600!")
	}

	// normf will be cleared as part of upgrading a File to SecretFile
	if normf.IsNone() {
		fmt.Println("Sucessfully cleared normal file after upgrading it to a secret file.")
	}

	// Calling ReadFile() on a SecretFile produces a Secret
	secret, ok := f.ReadFile()
	if !ok {
		fmt.Println("Failed to read from secret file")
	}

	// This is a secret, so we will only see a redacted value when we try to write it
	// to the console. The same will happen if we try to log it.
	fmt.Println("Got secret file contents: ", secret)

	// Similarly if we try to use stdlib logging libraries
	slog.Info("Second try printing secret file contents", "secret", secret)
	log.Printf("Third try printing secret file contents: %s", secret)

	return nil
}

func WritingAndDeletingFiles(path *string) error {
	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Delete a file. Works like os.Remove, but also returns an error if the path is still None
	err := f.Remove()
	if err != nil {
		fmt.Println("Failed to remove file: ", err)
	}

	// Write the contents of a file. Acts like os.WriteFile(path)
	data := []byte("Hello, World!")
	err = f.WriteFile(data, 0644)
	if err != nil {
		fmt.Println("Failed to write file: ", err)
	}

	// Open a file for read/write. File.Create() works like like os.Create(path), which means
	// calling this will either create a file or truncate an existing file. If you want to
	// append to a file, you must use File.OpenFile(os.O_RDWR|os.O_CREATE, 0644) in the same way
	// that would need to when calling os.OpenFile. See https://pkg.go.dev/os#OpenFile for details.
	var opened *os.File
	opened, err = f.Create()
	if err != nil {
		fmt.Println("Failed to open/create file: ", err)
		return err
	}
	defer opened.Close()

	// Now use the file handle exactly as you would if you called os.Create(path)
	opened.Write(data)

	return nil
}

func AdditionalFileTools(path *string) error {
	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Read back the path
	p, ok := f.Get()
	if ok {
		fmt.Println("Got path: ", p)
	} else {
		fmt.Println("No path given!")
		os.Exit(1)
	}

	// Check if the given path is the same as some other path, matching all equivalent absolute and relative paths.
	// In this case, check if the given path is equivalent to our working directory.
	if f.Match(".") {
		fmt.Println("We are operating on our working directory. Be careful!")
	} else {
		fmt.Println("We are not in our working directory. Go nuts!")
	}

	// Get a new optional with any relative path converted to absolute path (also ensuring it is a valid path)
	abs, err := f.Abs()
	if err != nil {
		fmt.Println("Could not convert path into absolute path. Is it a valid path?")
		return err
	}

	// Stat the file, or just check if it exists if you don't care about other file info
	if abs.Exists() {
		fmt.Println("The file exists!")
	}

	info, err := abs.Stat() // I don't care about the info
	if err != nil {
		fmt.Println("Could not stat the file")
	} else {
		fmt.Println("Got file info: ", info)
	}

	// Check that the file has permissions of at least 0444 (read), but is not 0111 (execute).
	// If those conditions are not fulfilled, we will set perms to 0644.
	valid, err := abs.FilePermsValid(0444, 0111)
	if err != nil {
		fmt.Println("Could not read file permissions!")
		return err
	}

	if !valid {
		err = abs.SetFilePerms(0644)
		if err != nil {
			fmt.Println("Failed to set file perms to 700")
			return err
		}
	}

	return nil
}
