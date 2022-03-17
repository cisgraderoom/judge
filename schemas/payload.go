package schemas

type Payload struct {
	Language  string   `json:"language"`
	Code      string   `json:"code"`
	Input     []string `json:"input"`
	Output    []string `json:"output"`
	TimeLimit int      `json:"time_limit"`
	MemLimit  int      `json:"mem_limit"`
	ProblemId int      `json:"problem_id"`
	Username  string   `json:"username"`
	Testcase  int      `json:"testcase"`
	MaxScore  float64  `json:"max_score"`
}
