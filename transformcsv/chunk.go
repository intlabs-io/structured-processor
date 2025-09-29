package transformcsv

import (
	"bytes"
	"fmt"
	"os"

	splitCsv "github.com/tolik505/split-csv"
)

/*
	Chunk the content into multiple smaller chunks
*/
func Chunk(content []byte) ([][]byte, error) {
	var chunks [][]byte
	// default batch size of 10MB
	batchSize := 10*1024*1024

	// If the content is smaller than the batch size, return the content as a single chunk
	if len(content) < batchSize {
		return [][]byte{content}, nil
	}

	file, err := StoreTemp(content, "tempfile")
	if err != nil {
		return nil, err
	}
	// Closing and removing the file after the function is finished (including if the function errors)
	defer file.Close()
	defer os.Remove(file.Name())
	
	chunkedFiles, err := ChunkFiles(file.Name(), "./csvDownloads/", batchSize)
	if err != nil {
		return nil, err
	}
	// Checking if the chunks are empty
	if len(chunkedFiles) == 0 {
		return [][]byte{}, fmt.Errorf("Length of chunked files is 0")
	}

	// Looping through the chunkedFiles
	for _, chunkedFile := range chunkedFiles {
		// Remove file after the loop is finished
		defer os.Remove(chunkedFile)
		// Reading the chunk
		chunk, err := os.ReadFile(chunkedFile)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

/*
	Chunk the file into multiple smaller file
*/
func ChunkFiles(filePath string, fileDir string, size int) ([]string, error) {
	splitter := splitCsv.New()
	splitter.Separator = ";"
	splitter.FileChunkSize = size
	result, err := splitter.Split(filePath, fileDir)
	return result, err
}

/*
	Combine the files into a single file
*/
func CombineFiles(file string, combineFile string) error {
	files := []string{file, combineFile}

	var buf bytes.Buffer
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		buf.Write(b)
	}

	err := os.WriteFile(combineFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
