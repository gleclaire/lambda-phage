package main

import "os"
import "github.com/tj/go-debug"
import "archive/zip"
import "io"
import "bytes"
import "strings"
import "fmt"

type zipFile struct {
	f     *os.File
	fName string
	*zip.Writer
}

// creates a new zip file
func newZipFile(fName string) (*zipFile, error) {
	f, err := os.Create(fName)
	if err != nil {
		return nil, err
	}

	return &zipFile{
		f,
		fName,
		zip.NewWriter(f),
	}, nil
}

// adds a file to this archive
func (z *zipFile) AddFile(fName string) (int64, error) {
	debug := debug.Debug("zipFile.AddFile")
	debug("opening source file")
	f, err := os.Open(fName)
	if err != nil {
		return 0, err
	}

	debug("source file opened")

	// want to make sure we get the base name for a file,
	// so fstat it, which is able to do that
	// debug("fstat source file")
	// s, err := f.Stat()
	// if err != nil {
	// 	return 0, err
	// }
	// bName := s.Name()

	// add file to archive
	debug("adding file to archive")
	wr, err := z.Create(f.Name())
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(wr, f)
	if err != nil {
		return 0, err
	}

	err = f.Close()
	if err != nil {
		return 0, err
	}

	return n, nil
}

// adds a file from string data to the archive
func (z *zipFile) AddString(fName string, str []byte) (int64, error) {
	debug := debug.Debug("zipFile.AddString")

	buf := bytes.NewBuffer(str)

	// add file to archive
	debug("adding file to archive")

	wr, err := z.Create(fName)
	if err != nil {
		return 0, err
	}

	debug("copying data for file")
	n, err := io.Copy(wr, buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// adds a directory recursively
func (z *zipFile) AddDirectory(f *os.File, infoCh chan<- string) error {
	debug := debug.Debug("zipFile.AddDirectory")

	dirName := f.Name()

	debug("adding directory %s", dirName)

	fi, err := f.Readdir(-1)

	if err != nil {
		return err
	}

	for _, info := range fi {
		debug("zname is...", z.f.Name())

		fName := info.Name()
		fullName := dirName + "/" + fName

		if info.IsDir() {
			if strings.Index(fName, ".") != 0 {
				debug("found directory, adding file %s", dirName+"/"+fName)

				if infoCh != nil {
					infoCh <- fmt.Sprintf("adding %s", fullName)
				}

				f, err := os.Open(fullName)
				if err != nil {
					// TODO: prob don't want to do this ALWAYS
					return err
				}

				err = z.AddDirectory(f, infoCh)
				if err != nil {
					return err
				}

				err = f.Close()
				if err != nil {
					return err
				}
			}
		} else if info.Name() != z.fName {
			debug("have file %s in directory %s, adding now", info.Name(), dirName)
			if infoCh != nil {
				infoCh <- fmt.Sprintf("adding %s", fullName)
			}

			_, err := z.AddFile(dirName + "/" + info.Name())
			if err != nil {
				// TODO: prob don't want to do this ALWAYS
				return err
			}
		}

	}

	return nil
}

// closes the writer and the file
func (z *zipFile) Close() error {
	err := z.Writer.Close()
	if err != nil {
		return err
	}

	return z.f.Close()
}
