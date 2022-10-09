package lc_api

import (
	// "io"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type graphqlPayload[T any] struct {
	Query         string `json:"query"`
	Variables     T      `json:"variables"`
	OperationName string `json:"operationName"`
}

func makeGraphqlRequest[GQLVariable any, Result any](variables GQLVariable, result *Result, gqlOperationName string, gqlQuery string) error {
	lcConfig := GetLcConfig()
	lcQueries := GetLcQueries()

	payload := graphqlPayload[GQLVariable]{Query: gqlQuery, OperationName: gqlOperationName, Variables: variables}
	client := http.Client{}
	requestBody, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", lcQueries.GRAPHQL_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-csrftoken", lcConfig.CSRF)
	request.Header.Set("csrftoken", lcConfig.CSRF)
	request.Header.Set("Referer", "https://leetcode.com/progress/")
	cookie := fmt.Sprintf("csrftoken=%s; LEETCODE_SESSION=%s", lcConfig.CSRF, lcConfig.LC_SESSION)
	request.Header.Set("cookie", cookie)
	resp, err := client.Do(request)
	// t, err := io.ReadAll(resp.Body)
	// fmt.Println(string(t))
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(result)
	return nil
}


func timestampToWord(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		_ = fmt.Errorf(err.Error())
		return ""
	}
	diff := time.Now().Sub(t).Hours() / 24
	diffInt := int (diff)
	switch diffInt {
	case 0: return "Today"
	case 1: return "Yesterday"
	default: return fmt.Sprintf("%d days ago", diffInt)
	}
}

