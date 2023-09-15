package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func BuildJson(checks *JsonOutput) SlackMessage {

	field := Field{Type: "plain_text", Text: "SSL Certificate Problem Detected"}
	section := Section{Type: "header", Text: field}
	blocks := []Section{section}

	for _, v := range checks.Results {
		field := Field{Type: "mrkdwn", Text: "URL: " + "<" + v.Resource + ">" + "\n\n" + "- TIME ELAPSED: " + v.Elapsed + "\n\n" + "- ERROR: " + v.Error}
		section := Section{Type: "section", Text: field}
		blocks = append(blocks, section)
	}
	message := SlackMessage{blocks}
	return message
}

func PostSlack(checks *JsonOutput) {
	slackmessage := BuildJson(checks)
	finalJson, err := json.MarshalIndent(slackmessage, "", " ")
	if err != nil {
		fmt.Print("Error Marshalling JSON")
	}

	req, err := http.NewRequest("POST", Cfg.SlackUrl, bytes.NewBuffer(finalJson))
	if err != nil {
		fmt.Print("Error making post request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
