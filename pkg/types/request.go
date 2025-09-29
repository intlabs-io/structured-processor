package types

type RequestBodyPaginate struct {
	Input  Input  `json:"input" validate:"required"`
	Output Output `json:"output" validate:"required"`
}

type Attributes struct {
	Paths []string `json:"paths"`
}

type PaginationResult struct {
	Message    string       `json:"message"`
	TotalPages int          `json:"totalPages"`
	Attributes Attributes   `json:"attributes"`
}

type RequestBodyTransform struct {
	Input   Input    `json:"input" validate:"required"`
	Output  Output   `json:"output" validate:"required"`
	Rules   []Rule   `json:"rules" validate:"required"`
	Webhook *Webhook `json:"webhook,omitempty"`
}
