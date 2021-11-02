package goMerge

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Put file path and content into fileContent
type fileContent struct {
	path string
	content string
	err error
}

//getFilesContent using parallelism, a goroutine walks the directory filesPath and returns the files contents
func getFilesContent(done <-chan struct{}, filesPath string, ext string) (<-chan fileContent, <-chan error){
	ch := make(chan fileContent)
	errCh := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(filesPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {
				var (
					data []byte
					err error
				)

				if filepath.Ext(path) == ext {
					data, err = ioutil.ReadFile(path)
				}

				select {
				case ch <- fileContent{path, string(data), err}:
				case <-done:
				}
				wg.Done()
			}()

			select {
			case <-done:
				return errors.New("walk into directory canceled")
			default:
				return nil
			}
		})
		go func() {
			wg.Wait()
			close(ch)
		}()
		errCh <- err
	}()
	return ch, errCh
}

// Merge get files directory, extension, outMergedName and remove flag for merge files to one file
func Merge(path string, extension string, outMergedName string, remove bool) error {
	done := make(chan struct{})
	defer close(done)

	// get files content from channel
	chFiles, errCh := getFilesContent(done, path, extension)

	var contents []string

	// append files content to data string
	for f := range chFiles {
		if f.err != nil {
			return f.err
		}
		// append content to contents slice
		contents = append(contents, f.content)

		// remove files unmerged files if true
		if remove {
			// remove file after append file content
			if err := os.Remove(f.path); err != nil {
				return err
			}
		}
	}

	if err := <-errCh; err != nil {
		return err
	}

	// create new merged files
	newMergedFile, err := os.Create(outMergedName)
	if err != nil {
		return err
	}
	defer newMergedFile.Close()

	// write content to merged file
	for _, content := range contents {
		if _, err := fmt.Fprintln(newMergedFile, content); err != nil {
			return err
		}
	}

	return nil
}
