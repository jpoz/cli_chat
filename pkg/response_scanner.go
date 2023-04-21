package clichat

import (
	"bufio"
	"strings"
)

type AiResponse struct {
	Action      string `json:"action"`
	actionIdx   int
	ActionInput string `json:"action_input"`
	Agent       string `json:"agent"`
	Thought     string `json:"thought"`
}

func ParseResponse(text string) AiResponse {
	scanner := bufio.NewScanner(strings.NewReader(text))

	response := AiResponse{}
	idx := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		event := parts[0]
		data := strings.TrimSpace(strings.Join(parts[1:], ":"))

		switch event {
		case "Action":
			response.Action = data
			response.actionIdx = idx
		case "Action Input":
			if idx == response.actionIdx+1 {
				response.ActionInput = data
				return response
			} else {
				response.ActionInput = ""
				response.actionIdx = 0
			}
		case "Agent":
			response.Agent = data
		case "Thought":
			response.Thought += data + "\n"
		}

		idx++
	}

	return response
}
