package storage

import (
	"testing"
	"fmt"
	"strings"
	"encoding/csv"

	"github.com/go-playground/assert/v2"
	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
)

func TestRDBPostgres(t *testing.T) {
	t.Run("postgres CSV data seeing if client works properly", func(t *testing.T) {
		p := postgres.Preset(
			postgres.WithUser("gnomock", "gnomick"),
			postgres.WithDatabase("mydb"),
			postgres.WithQueriesFile("../assets/goldenFiles/testRDBPostgres.sql"),
		)
		container, err := gnomock.Start(p)
		if err != nil {
			t.Fatalf("failed to start gnomock container: %v", err)
		}
		t.Cleanup(func() { _ = gnomock.Stop(container) })
	
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			container.Host, container.DefaultPort(),
			"gnomock", "gnomick", "mydb",
		)
	
		input := Test_Input{
			Server:        "postgres",
			ConnectionStr: connStr,
			Query:         "select * from users",
		}
	
		bytesContent, err := Test_DownloadFromRDB(input)
		if err != nil {
			t.Fatalf("failed to download proxy: %v", err)
		}
	
		reader := csv.NewReader(strings.NewReader(string(bytesContent)))
		reader.Comma = ','
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1
		reader.TrimLeadingSpace = true
	
		lines, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("failed to read CSV content: %v", err)
		}
	
		var responseCSV [][]string
		responseCSV = append(responseCSV, []string{"id", "full_name", "phone", "created_at", "updated_at"})
		responseCSV = append(responseCSV, []string{"1", "John Doe", "123-456-7890", "2023-10-01T00:00:00Z", "2023-10-01T00:00:00Z"})
		assert.Equal(t, lines, responseCSV)
	})
	
	t.Run("postgres fail", func(t *testing.T) {
		p := postgres.Preset(
			postgres.WithUser("gnomock", "gnomick"),
			postgres.WithDatabase("mydb"),
			postgres.WithQueriesFile("../assets/goldenFiles/testRDBPostgres.sql"),
		)
		container, err := gnomock.Start(p)
		if err != nil {
			t.Fatalf("failed to start gnomock container: %v", err)
		}
		t.Cleanup(func() { _ = gnomock.Stop(container) })
	
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			container.Host, container.DefaultPort(),
			"gnomock", "gnomick", "mydb",
		)
	
		input := Test_Input{
			Server:        "postgres",
			ConnectionStr: connStr,
			Query:         "select * from tabledne",
		}
	
		_, err = Test_DownloadFromRDB(input)
		assert.MatchRegex(t, err.Error(), `pq: relation "tabledne" does not exist`)
	})
	
}
