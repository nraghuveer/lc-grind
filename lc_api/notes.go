package lc_api

type noteQueryVariables struct {
	TitleSlug string `json:"titleSlug"`
}

type noteQueryResponse struct {
	Data struct {
		Question struct {
			QuestionId string `json:"questionId"`
			Note       string `json:"note"`
		} `json:"question"`
	} `json:"data"`
}

func GetNote(title string) (string, error) {
	lcQueries := GetLcQueries()
	variables := noteQueryVariables{TitleSlug: title}
	result := noteQueryResponse{}
	if err := makeGraphqlRequest(variables, &result, "QuestionNote", lcQueries.NOTE_QUERY); err != nil {
		return "", err
	}
	return result.Data.Question.Note, nil
}
