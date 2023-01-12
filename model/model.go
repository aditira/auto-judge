package model

type CreateSubmissionResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type LanguangeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GetSubmissionResponse struct {
	SourceCode     string `json:"source_code"`
	LanguangeId    int    `json:"languange_id"`
	Stdin          string `json:"stdin,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	Stdout         string `json:"stdout,omitempty"`
	StatusId       int    `json:"status_id"`
	CreatedAt      string `json:"created_at"`
	FinishedAt     string `json:"finished_at"`
	Time           string `json:"time"`
}

type ReportSpreadSheet struct {
	Name                  string
	Email                 string
	Campus                string
	Code1                 string
	ProgrammingLanguange1 string
	Code2                 string
	ProgrammingLanguange2 string
}
