package lc_api

type noteQueryVariables struct {
	TitleSlug string `json:"titleSlug"`
}

type noteQueryPayload struct {
	OperationName string             `json:"operationName"`
	Variables     noteQueryVariables `json:"variables"`
	Query         string             `json:"query"`
}

type noteQueryResponse struct {
	Data struct {
		Question struct {
			QuestionId string `json:"questionId"`
			Note       string `json:"note"`
		} `json:"question"`
	} `json:"data"`
}

func GetNote(title string) string { return "xyz" }
