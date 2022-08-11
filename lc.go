package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"git.mills.io/prologic/bitcask"
)

var db *bitcask.Bitcask
var dbErr error
var once sync.Once

const (
	CSRF       = "BUzkbtjIQSt3mCiO33MZeeN6CF8S8Ww4pxf7Kn8aWWKtcVM9lHrP5kNw1MyuU2Wq"
	LC_SESSION = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJfcGFzc3dvcmRfcmVzZXRfa2V5IjoiNjFvLWY5MTM1ZGE1MDFkYjY1NDZmYTBhIiwiX2F1dGhfdXNlcl9pZCI6IjU3NTk1MiIsIl9hdXRoX3VzZXJfYmFja2VuZCI6ImFsbGF1dGguYWNjb3VudC5hdXRoX2JhY2tlbmRzLkF1dGhlbnRpY2F0aW9uQmFja2VuZCIsIl9hdXRoX3VzZXJfaGFzaCI6IjYzMGUwMjBhMDBlYTk1NTE4MTI1YWVlNjgzMWI5ZTc4ZDkzMWE1ZDUiLCJpZCI6NTc1OTUyLCJlbWFpbCI6InJhZ2h1dmVlcm5hcmFoYXJpc2V0dGlAZ21haWwuY29tIiwidXNlcm5hbWUiOiJuYXJhaGFyaXNldHRpIiwidXNlcl9zbHVnIjoibmFyYWhhcmlzZXR0aSIsImF2YXRhciI6Imh0dHBzOi8vYXNzZXRzLmxlZXRjb2RlLmNvbS91c2Vycy9uYXJhaGFyaXNldHRpL2F2YXRhcl8xNTU0NzI1NTI1LnBuZyIsInJlZnJlc2hlZF9hdCI6MTY1OTk0MTE5OSwiaXAiOiIxMTcuMjIxLjc4LjM4IiwiaWRlbnRpdHkiOiI0MWQ2NTYxNGE5ZjhlMzRiMjZlNjFhZGU4MDIzZmZmZCIsIl9zZXNzaW9uX2V4cGlyeSI6MTIwOTYwMCwic2Vzc2lvbl9pZCI6MjM1NzM3OTV9.TYmbGIlP6i6KMtmXZSPI_ysU9XydKuw1SCMWPoDP9ZY"

	GRAPHQL_URL     = "https://leetcode.com/graphql"
	SUBMISSIONS_URL = "https://leetcode.com/api/submissions/?offset=%d&limit=%d&lastkey=%s"
	NOTE_QUERY      = `
query QuestionNote($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
        questionId
        note
        __typename
  }
}
`
)

type noteQueryVariables struct {
	TitleSlug string `json:"titleSlug"`
}

type noteQueryPayload struct {
	OperationName string             `json:"operationName"`
	Variables     noteQueryVariables `json:"variables"`
	Query         string             `json:"query"`
}

type submission struct {
	Id           int64  `json:"id"`
	Lang         string `json:"lang"`
	Time         string `json:"time"`
	Timestamp    int64  `json:"timestamp"`
	Url          string `json:"url"`
	ProblemTitle string `json:"title"`
	Title_slug   string `json:"title_slug"`
}

type submissions struct {
	Submissions []submission `json:"submissions_dump"`
	HasNext     bool         `json:"has_key"`
	LastKey     string       `json:"last_key"`
}

type noteQueryResponse struct {
	Data struct {
		Question struct {
			QuestionId string `json:"questionId"`
			Note       string `json:"note"`
		} `json:"question"`
	} `json:"data"`
}

func (s *submissions) getUniqueSubmissions() []submission {
	var result []submission

	for _, sub := range s.Submissions {
		if len(result) == 0 || (len(result) > 0 && result[len(result)-1].ProblemTitle != sub.ProblemTitle) {
			result = append(result, sub)
		}
	}
	return result
}

func GetAllSubmissions(thresholdTime time.Time) ([]submission, error) {
	if thresholdTime.After(time.Now()) {
		return nil, errors.New("Invalid thresholdTime, given time is in future")
	}
	// we never know how many submissions were there in a time period
	// so give each day as step in progress
	// if we have 150 days, each day gets an increment of 150/100
	// so from the now the lastTime is 20 days before,
	// set the progress values as => 20 * 100/150
	progress := 0.0
	diff := time.Now().Sub(thresholdTime)
	totalDays := diff.Hours() / 24

	curTime := time.Now() // time where we are right now in loading the data, we go backwards
	offset := 0
	limit := 20
	lastKey := ""
	var result []submission

	for curTime.After(thresholdTime) {
		curSubmissions, curLastKey, err := fetchSubmissions(offset, limit, lastKey)
		if err != nil {
			log.Fatalln("Error while getting submissions", err.Error())
		}
		lastKey = curLastKey
		offset += len(curSubmissions)
		result = append(result, curSubmissions...)
		if len(curSubmissions) > 0 {
			curTime = time.Unix(curSubmissions[len(curSubmissions)-1].Timestamp, 0)
		}

		// set the progress variable
		diff := time.Now().Sub(curTime)
		days := (diff.Hours() / 24)
		progress = days * (totalDays / 100)
	}

	for _, sub := range result {
		saveSubmissionRecordDB(&sub)
	}
	return result, nil
}

func fetchSubmissions(offset, limit int, lastKey string) ([]submission, string, error) {
	client := http.Client{}
	requestBody, _ := json.Marshal(map[string]string{})
	url := fmt.Sprintf(SUBMISSIONS_URL, offset, limit, lastKey)
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRFToken", CSRF)
	cookie := fmt.Sprintf("CSRF=%s; LEETCODE_SESSION=%s", CSRF, LC_SESSION)
	request.Header.Set("cookie", cookie)
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var result submissions
	json.NewDecoder(resp.Body).Decode(&result)
	return result.getUniqueSubmissions(), result.LastKey, nil
}

func getNote(titleSlug string) string {
	payload := noteQueryPayload{Query: NOTE_QUERY, OperationName: "QuestionNote", Variables: noteQueryVariables{TitleSlug: titleSlug}}
	client := http.Client{}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Fatalln(err)
	}
	request, err := http.NewRequest("POST", GRAPHQL_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-csrftoken", CSRF)
	request.Header.Set("csrftoken", CSRF)
	request.Header.Set("Referer", "https://leetcode.com")
	cookie := fmt.Sprintf("csrftoken=%s; LEETCODE_SESSION=%s", CSRF, LC_SESSION)
	request.Header.Set("cookie", cookie)
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	var result noteQueryResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.Question.Note
}

func GetDB() (*bitcask.Bitcask, error) {
	once.Do(func() { db, dbErr = bitcask.Open("/Users/raghuveernaraharisetti/Documents/lc-db/") })
	return db, dbErr
}

func saveSubmissionRecordDB(sub *submission) error {
	value, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	// db is always initialized in the main
	db.Put([]byte(sub.Title_slug), []byte(value))
	return nil
}
