package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"

	"github.com/auto-judge/model"
)

// https://ce.judge0.com/languages
// https://ce.judge0.com/statuses

func init() {
	godotenv.Load()
}

func main() {
	res, err := getSheetReport()
	if err != nil {
		panic(err.Error())
	}

	for _, v := range *res {
		fmt.Println("name: ", v.Name)
		fmt.Println("code: ", v.Code1)
		fmt.Println("programming languange: ", v.ProgrammingLanguange1)
		progID, _ := getLangID(v.ProgrammingLanguange1)
		fmt.Println("languange code: ", *progID)

		resTkn, err := createSubmission(*progID, v.Code1, "50196510036")
		if err != nil {
			panic(err.Error())
		}

		resSubmission, err := getSubmission(resTkn.Token)

		sourceCode, err := base64.StdEncoding.DecodeString(resSubmission.SourceCode)
		output, err := base64.StdEncoding.DecodeString(resSubmission.Stdout)
		expectedOutput, err := base64.StdEncoding.DecodeString(resSubmission.ExpectedOutput)

		if err != nil {
			panic(err)
		}

		fmt.Println("Source Code: ", string(sourceCode))
		fmt.Println("Output: ", string(output))
		fmt.Println("Expected Code: ", string(expectedOutput))

		fmt.Println("-------------------------------------------------")
	}

}

func getLangID(progLang string) (*int, error) {
	res, err := getLanguanges()
	if err != nil {
		return nil, err
	}

	var progID int
	for _, v := range *res {
		lowerName := strings.ToLower(v.Name)
		if strings.Contains(lowerName, progLang) {
			progID = v.ID
		}
	}

	return &progID, nil
}

func getSheetReport() (*[]model.ReportSpreadSheet, error) {
	data, err := ioutil.ReadFile("config/sheet.json")
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	client := conf.Client(context.TODO())
	srv, err := sheets.New(client)
	if err != nil {
		return nil, err
	}

	resp, err := srv.Spreadsheets.Values.Get(os.Getenv("SPREAD_SHEET_ID"), "Responses-Dev!A:U").Do()
	if err != nil {
		return nil, err
	}

	var reportResults []model.ReportSpreadSheet

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("Sheet is empty !")
	} else {
		for _, row := range resp.Values {
			fmt.Println("-------------------------")

			var code1, code2, lang1, lang2 string

			re := regexp.MustCompile("^```(.*)")
			match := re.FindStringSubmatch(row[17].(string))
			if len(match) == 2 {
				lang1 = match[1]
				re = regexp.MustCompile("^```" + regexp.QuoteMeta(lang1) + "\n|\n```$")
				newCode := re.ReplaceAllString(row[17].(string), "")
				code1 = newCode
			}

			re = regexp.MustCompile("^```(.*)")
			match = re.FindStringSubmatch(row[19].(string))
			if len(match) == 2 {
				lang2 = match[1]
				re = regexp.MustCompile("^```" + regexp.QuoteMeta(lang1) + "\n|\n```$")
				newCode := re.ReplaceAllString(row[19].(string), "")
				code2 = newCode
			}

			reportResult := model.ReportSpreadSheet{
				Name:                  row[3].(string),
				Email:                 row[1].(string),
				Campus:                row[4].(string),
				Code1:                 code1,
				ProgrammingLanguange1: lang1,
				Code2:                 code2,
				ProgrammingLanguange2: lang2,
			}

			reportResults = append(reportResults, reportResult)
		}
	}

	return &reportResults, nil
}

func createSubmission(languangeID int, sourceCode, expectedOutput string) (*model.CreateSubmissionResponse, error) {
	url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=true&fields=*"
	postBody, _ := json.Marshal(map[string]interface{}{
		"language_id":     languangeID,
		"source_code":     base64.StdEncoding.EncodeToString([]byte(sourceCode)),
		"expected_output": base64.StdEncoding.EncodeToString([]byte(expectedOutput)),
	})

	payload := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-RapidAPI-Key", os.Getenv("X_RAPID_API_KEY"))
	req.Header.Add("X-RapidAPI-Host", "judge0-ce.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

	var jsonResponse model.CreateSubmissionResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return nil, err
	}

	return &jsonResponse, nil
}

func getSubmission(token string) (*model.GetSubmissionResponse, error) {
	url := "https://judge0-ce.p.rapidapi.com/submissions/" + token + "?base64_encoded=true&fields=*"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", os.Getenv("X_RAPID_API_KEY"))
	req.Header.Add("X-RapidAPI-Host", "judge0-ce.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("resp-string: ", string(body))

	var jsonResponse model.GetSubmissionResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		panic(err)
	}

	return &jsonResponse, nil
}

func getStatuses() (*[]model.StatusResponse, error) {
	resp, err := http.Get("https://ce.judge0.com/statuses")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponses []model.StatusResponse
	err = json.Unmarshal(body, &jsonResponses)
	if err != nil {
		return nil, err
	}

	return &jsonResponses, nil
}

func getLanguanges() (*[]model.LanguangeResponse, error) {
	resp, err := http.Get("https://ce.judge0.com/languages/")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponses []model.LanguangeResponse
	err = json.Unmarshal(body, &jsonResponses)
	if err != nil {
		return nil, err
	}

	return &jsonResponses, nil
}
