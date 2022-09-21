package lc_api

type lcQueries struct {
	NOTE_QUERY          string
	GRAPHQL_URL         string
	PROGRESS_LIST_QUERY string
}

func GetLcQueries() lcQueries {
	return lcQueries{NOTE_QUERY: `
query QuestionNote($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
        questionId
        note
        __typename
  }
}
`, PROGRESS_LIST_QUERY: `
query progressList($pageNo: Int, $numPerPage: Int, $filters: ProgressListFilterInput) {
	isProgressCalculated
	solvedQuestionsInfo(pageNo: $pageNo, numPerPage: $numPerPage, filters: $filters) {
		currentPage
		pageNum
		totalNum
		data {
			totalSolves
			question {
				questionFrontendId
				questionTitle
				questionDetailUrl          
				difficulty
				topicTags {
					name
					slug
				}        
			}        
			lastAcSession {
				time
				wrongAttempts
			}
		}     
}   
`,
		GRAPHQL_URL: "https://leetcode.com/graphql"}
}
