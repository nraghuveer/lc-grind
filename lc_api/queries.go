package lc_api

type lcQueries struct {
	NOTE_QUERY          string
	GRAPHQL_URL         string
	PROGRESS_LIST_QUERY string
}

func GetLcQueries() lcQueries {
	return lcQueries{
		NOTE_QUERY: `
query QuestionNote($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
        questionId
        note
        __typename
  }
}
`,
// PROGRESS_LIST_QUERY: "query progressList($pageNo: Int, $numPerPage: Int, $filters: ProgressListFilterInput) { isProgressCalculated solvedQuestionsInfo(pageNo: $pageNo, numPerPage: $numPerPage, filters: $filters) { currentPage pageNum totalNum data { totalSolves question { questionFrontendId questionTitle questionDetailUrl difficulty topicTags { name slug } } lastAcSession { time wrongAttempts } } }",
PROGRESS_LIST_QUERY: "\n query progressList($pageNo: Int, $numPerPage: Int, $filters: ProgressListFilterInput) {\n isProgressCalculated\n solvedQuestionsInfo(pageNo: $pageNo, numPerPage: $numPerPage, filters: $filters) {\n currentPage\n pageNum\n totalNum\n data {\n totalSolves\n question {\n questionFrontendId\n questionTitle\n questionDetailUrl\n difficulty\n topicTags {\n name\n slug\n }\n }\n lastAcSession {\n time\n wrongAttempts\n }\n }\n }\n}\n",
GRAPHQL_URL: "https://leetcode.com/graphql"}
}
