package types

type SourceReference struct {
	// ONE DRIVE && GOOGLE DRIVE
	Id string `json:"id,omitempty"`
	// S3
	Bucket string `json:"bucket,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Region string `json:"region,omitempty"`
	// REST
	Host string `json:"host,omitempty"`
	// RDB
	Server   string `json:"server,omitempty"`
	SSL      bool   `json:"ssl,omitempty"`
	Database string `json:"database,omitempty"`
	Table    string `json:"table,omitempty"`
	Query    string `json:"query,omitempty"`
}

type Secrets struct {
	// S3
	Secret string `json:"secret,omitempty"`
	// RDB
	Password string `json:"password,omitempty"`
	// REST
	ApiBearerToken string `json:"apiBearerToken,omitempty"`
	// ONE DRIVE && GOOGLE DRIVE
	AccessToken string `json:"accessToken,omitempty"`
}

type Resources struct {
	// ONE DRIVE && GOOGLE DRIVE
	Id string `json:"id,omitempty"`
	// RDB
	Username string `json:"username,omitempty"`
}

type SourceCredential struct {
	Secrets   Secrets   `json:"secrets"`
	Resources Resources `json:"resources"`
}

type Input struct {
	StorageType string           `json:"storageType"`
	DataType    string           `json:"dataType"` // CSV, JSON, JSONL, SQL
	Reference   SourceReference  `json:"reference"`
	Credential  SourceCredential `json:"credential"`
}

type Output struct {
	StorageType string           `json:"storageType"`
	DataType    string           `json:"dataType"` // CSV, JSON, JSONL, SQL
	Reference   SourceReference  `json:"reference"`
	Credential  SourceCredential `json:"credential"`
}
