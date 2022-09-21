package lc_api

import (
	"errors"

	"github.com/nraghuveer/lc-grind/protocols"
)

type progressListQueryVariables struct {
	filters    map[string]string
	pageNo     int
	numPerPage int
}

type topicTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProgressQuestion struct {
	Id         string `json:"questionFrontendId"`
	Title      string `json:"questionTitle"`
	URL        string `json:"questionDetailUrl"`
	Difficulty string `json:"difficulty"`
}

func (pq *ProgressQuestion) FilterValue() string { return pq.Title }

type solvedQuestionsInfoDataItem struct {
	TotalSolves int              `json:"totalSolves"`
	Question    ProgressQuestion `json:"question"`
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

type progressPageIterator struct {
	curIdx    int
	total     int
	questions *solvedQuestionsInfo
}

func (ppi progressPageIterator) HasNext() bool { return ppi.curIdx < ppi.total-1 }
func (ppi progressPageIterator) Next() (*ProgressQuestion, error) {
	if !ppi.HasNext() {
		return nil, errors.New("No items in the progress page iterator")
	}
	ppi.curIdx += 1
	return &ppi.questions.Data[ppi.curIdx].Question, nil
}

func (pp *progressPage) CreateIterator() protocols.Iterator[*ProgressQuestion] {
	return progressPageIterator{}
}

type Progress struct {
	numPerPage     int
	curPageNo      int
	totalPages     int
	totalQuestions int
	pages          []*progressPage
}

func (pc Progress) Init() error {
	pc.curPageNo = 0
	pc.totalPages = 1000 // to make the first HasNexxt call happy
	pc.numPerPage = 10
	_, err := pc.FetchNext()
	if err != nil {
		return err
	}
	if len(pc.pages) <= 0 {
		return errors.New("Failed to fetch first page from the progress")
	}
	pc.totalPages = pc.pages[0].Data.SolvedQuestions.TotalPages
	pc.totalQuestions = pc.pages[0].Data.SolvedQuestions.TotalQuestions
	return nil
}

func (pc Progress) CompletedPercentage() float32 {
	if pc.curPageNo >= pc.totalPages {
		return 100.0
	}
	return (float32(pc.curPageNo) / float32(pc.totalPages)) * 100.0
}

func (pc Progress) HasNext() bool { return pc.curPageNo <= pc.totalPages }

func (pc Progress) addNewPage(page *progressPage) { pc.pages = append(pc.pages, page) }

func (pc Progress) FetchNext() (*progressPage, error) {
	pc.curPageNo += 1
	lcQueries := GetLcQueries()
	if !pc.HasNext() {
		return nil, errors.New("No Pages to fetch from progress")
	}
	nextPage := &progressPage{}
	err := makeGraphqlRequest(progressListQueryVariables{pageNo: pc.curPageNo, numPerPage: pc.numPerPage, filters: make(map[string]string)}, nextPage, "progressList", lcQueries.PROGRESS_LIST_QUERY)
	if err != nil {
		return nil, err
	}
	pc.addNewPage(nextPage)
	return nextPage, nil
}

// Implements protocols.Iterator
type ProgressIterator struct {
	curIdx   int // we have read till curIdx
	totalLen int
	iters    []protocols.Iterator[*ProgressQuestion]
}

func (pi ProgressIterator) HasNext() bool {
	// assume the curIdx is always on the right index to read now
	if pi.curIdx >= pi.totalLen {
		return false
	}
	// If on the last iter
	if pi.curIdx == pi.totalLen-1 && !pi.iters[pi.curIdx].HasNext() {
		return false
	}
	if !pi.iters[pi.curIdx].HasNext() {
		pi.curIdx += 1
		return pi.HasNext()
	}
	return true
}
func (pi ProgressIterator) Next() (*ProgressQuestion, error) {
	if !pi.HasNext() {
		return nil, errors.New("No items to iter")
	}
	// We are not on the last iter
	return pi.iters[pi.curIdx].Next()
}

func (p Progress) CreateIterator() protocols.Iterator[*ProgressQuestion] {
	iters := make([]protocols.Iterator[*ProgressQuestion], len(p.pages))
	for i := 0; i < len(p.pages); i++ {
		iters[i] = p.pages[i].CreateIterator()
	}
	return ProgressIterator{curIdx: -1, iters: iters, totalLen: len(iters)}
}
