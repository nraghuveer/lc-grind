package lc_api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nraghuveer/lc-grind/protocols"
)

type progressListQueryVariables struct {
	PageNo     int         `json:"pageNo"`
	NumPerPage int         `json:"numPerPage"`
	Filters    interface{} `json:"filters"`
}

type topicTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProgressQuestion struct {
	Id            string `json:"questionFrontendId"`
	QuestionTitle string `json:"questionTitle"`
	URL           string `json:"questionDetailUrl"`
	Difficulty    string `json:"difficulty"`
	LastSolved string
}

func (p ProgressQuestion) ParseTitleSlug() string {
	parts := strings.Split(p.URL, "/")
	return parts[2]
}

func (pq *ProgressQuestion) FilterValue() string { return pq.QuestionTitle }
func (pq *ProgressQuestion) Title() string       { return pq.QuestionTitle }
func (pq *ProgressQuestion) Description() string { return fmt.Sprintf("%s | %s", pq.Difficulty, timestampToWord(pq.LastSolved)) }
func (pq *ProgressQuestion) String() string      { return fmt.Sprintf("%s: %s", pq.Id, pq.QuestionTitle) }

type solvedQuestionsInfoDataItem struct {
	TotalSolves int              `json:"totalSolves"`
	Question    ProgressQuestion `json:"question"`
	LastAcSession struct {
		Time string `json:"time"`
		WrongAttempts int `json:"wrongAttempts"`
	} `json:"lastAcSession"`
}

type solvedQuestionsInfo struct {
	CurrentPage    int                           `json:"currentPage"`
	TotalPages     int                           `json:"pageNum"`
	TotalQuestions int                           `json:"totalNum"`
	Data           []solvedQuestionsInfoDataItem `json:"data"`
}

type progressPage struct {
	Data struct {
		SolvedQuestions solvedQuestionsInfo `json:"solvedQuestionsInfo"`
	} `json:"data"`
}

type Progress struct {
	numPerPage     int
	curPageNo      int
	totalPages     int
	questions      []*ProgressQuestion
	curQuestionIdx int
}

func (pc *Progress) Init() error {
	pc.curPageNo = 0
	pc.totalPages = -1
	pc.numPerPage = 10
	pc.questions = make([]*ProgressQuestion, 4000)
	pc.curQuestionIdx = -1
	if err := pc.FetchNext(); err != nil {
		return err
	}
	return nil
}

func (pc *Progress) CompletedPercentage() float32 {
	if pc.curPageNo >= pc.totalPages {
		return 100.0
	}
	return (float32(pc.curPageNo) / float32(pc.totalPages)) * 100.0
}

func (pc *Progress) HasNext() bool { return pc.curPageNo <= pc.totalPages }

func (pc *Progress) FetchNext() error {
	pc.curPageNo += 1
	lcQueries := GetLcQueries()
	nextPage := &progressPage{}
	if err := makeGraphqlRequest(progressListQueryVariables{PageNo: pc.curPageNo, NumPerPage: pc.numPerPage, Filters: struct{}{}}, nextPage, "progressList", lcQueries.PROGRESS_LIST_QUERY); err != nil {
		return err
	}
	pc.totalPages = nextPage.Data.SolvedQuestions.TotalPages
	pc.totalPages = 5 // FIXME: for testing, since loading everything is expensive
	for _, questionItem := range nextPage.Data.SolvedQuestions.Data {
		pc.curQuestionIdx += 1
		curQuestion := questionItem.Question
		curQuestion.LastSolved = questionItem.LastAcSession.Time
		pc.questions[pc.curQuestionIdx] = &curQuestion
	}
	return nil
}

// Implements Aggregate[*ProgressQuestion]
func (p *Progress) CreateIterator() protocols.Iterator[*ProgressQuestion] {
	return &ProgressIterator{curIdx: -1, total: p.curQuestionIdx + 1, elements: p.questions}
}

type ProgressIterator struct {
	curIdx   int // the idx of item that is just served by the iterator
	total    int
	elements []*ProgressQuestion
}

func (pi *ProgressIterator) HasNext() bool { return pi.curIdx < pi.total-1 }

func (pi *ProgressIterator) Next() (*ProgressQuestion, error) {
	if !pi.HasNext() {
		return nil, errors.New("no more items in the progress iter")
	}
	pi.curIdx += 1
	return pi.elements[pi.curIdx], nil
}
