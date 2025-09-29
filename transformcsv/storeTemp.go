package transformcsv

import (
	"os"
	"path/filepath"
)

/*
	Store the content into a temporary file
*/
func StoreTemp(content []byte, fileName string) (*os.File, error) {
	// Writing CSV to DIR
	localPath := filepath.Join("./csvDownloads/", fileName)
	err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	// Creating a temporary file
	file, err := os.CreateTemp(filepath.Dir(localPath), "temp-*.csv")
	if err != nil {
		return nil, err
	}
	_, err = file.Write(content)
	if err != nil {
		return nil, err
	}
	return file, nil
}
