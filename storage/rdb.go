package storage

import (
	"database/sql"
	"fmt"
	"lazy-lagoon/pkg/types"

	_ "github.com/lib/pq"
)

/*
Download the content from rdb
*/
func DownloadFromRDB(input types.Input) ([]byte, error) {
	var connectionStr string
	if input.Reference.Server == "postgres" {
		connectionStr = fmt.Sprintf("postgresql://%s:%s@%s/%s",
			input.Credential.Resources.Username,
			input.Credential.Secrets.Password,
			input.Reference.Host,
			input.Reference.Database)
	}
	db, err := sql.Open(input.Reference.Server, connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(input.Reference.Query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var csvData []byte
	var result string

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Append column headers to result
	for i, col := range columns {
		if i > 0 {
			result += ","
		}
		result += col
	}
	result += "\n"

	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, v := range values {
			if i > 0 {
				result += ","
			}
			if v.Valid {
				result += v.String
			}
		}
		result += "\n"
	}

	csvData = []byte(result)
	return csvData, nil
}

type Test_Input struct {
	Server        string
	ConnectionStr string
	Query         string
}

func Test_DownloadFromRDB(input Test_Input) ([]byte, error) {
	db, err := sql.Open(input.Server, input.ConnectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(input.Query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var csvData []byte
	var result string

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Append column headers to result
	for i, col := range columns {
		if i > 0 {
			result += ","
		}
		result += col
	}
	result += "\n"

	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, v := range values {
			if i > 0 {
				result += ","
			}
			if v.Valid {
				result += v.String
			}
		}
		result += "\n"
	}

	csvData = []byte(result)
	return csvData, nil
}
