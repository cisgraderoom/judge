package schemas

type Payload struct {
	Mode         string   `json:"mode"`
	SubmissionId int      `json:"submission_id"`
	Language     string   `json:"language"`
	Code         string   `json:"code"`
	Input        []string `json:"input"`
	Output       []string `json:"output"`
	TimeLimit    int      `json:"time_limit"`
	MemLimit     int      `json:"mem_limit"`
	ProblemId    int      `json:"problem_id"`
	Username     string   `json:"username"`
	Testcase     int      `json:"testcase"`
	MaxScore     float64  `json:"max_score"`
	Classcode    string   `json:"classcode"`
	JobId        int      `json:"job_id"`
}
