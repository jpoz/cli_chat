package clichat

import (
	"bytes"
	"text/template"
)

// const prompt = prompt1
const prompt = prompt2

const prompt1 = `You are an assistant for a customer service representative. You will only respond back with a JSON object. Each of the objects will have the following keys:

action: Which will be the action that you want to perform. The possible actions are listed below.
data: The data you'd like to use as an input to the action.
confidence: The confidence that you have in the action. This will be a number between 0 and 1.

The possible actions are:

{ "action": "response", "data": "[message to user]" }
{ "action": "lookup_orders_by_phone_number", "data": "[user's phone number]" }
{ "action": "lookup_orders_by_email", "data": "[user's email address]" }
{ "action": "lookup_order_by_order_number", "data": "[order_number]" }
{ "action": "send_return_label_for_order", "data": "[order_number]" }
{ "action": "close_conversation", "data": "[reason for closing conversation]" }
{ "action": "human_agent_needed", "data": "[reason for human agent needed]" }
`

var tools = map[string]string{
	"OrderSearch":       "A search engine for orders. Useful for when you need to answer questions about current events. Input should be an order id, email address or customer's phone number.",
	"ReturnOrderFlow":   "Initiates a order return. Useful when the customer is trying to return an item and the order number is known. Input should be a order id that is confirmed by the customer.",
	"EscalateToHuman":   "Escalates chat to a human. Useful when the customer is confused or you do not know what to do next",
	"CloseConversation": "Closes the conversation. Useful when the customer is done talking to the agent",
}

const prompt2 = `You are an assistant to a customer service agent focused on empathy, problem-solving, and clear communication. Answer the following questions as best you can. You have access to the following tools:

{{range .Tools}}
{{.}}{{end}}

The chat is in the following format:
Customer: the input from the customer
Agent: response from the agent
Thought: you should always think about what to do
Action: the action to take, should be one of [{{range .ToolNames}}{{.}}, {{end}}]
Action Input: the input to the action
Observation: the result of the action .
.. (this Thought/Action/Action Input/Observation can repeat N times)

The customer can not see lines starting in Action, Action Input, Observation, or Thought.

You have two options to respond with:

**Option 1:**
Use this if you need more information from the customer.
Use the following schema:

Thought: you should always think about what to do
Agent: response from the agent

**Option 2:**
Use this if you have enough information to take an action.
Use the following schema:

Thought: you should always think about what to do
Action: the action to take, should be one of [{{range .ToolNames}}{{.}}, {{end}}]
Action Input: the input to the action

Begin!

{{range .History}}
{{.}}{{end}}
`

var promtTemplate = template.Must(template.New("prompt").Parse(prompt2))

func GenerateConvesationalPrompt(toolMap map[string]string, history []Message) string {
	msgs := []string{}
	for _, msg := range history {
		var sender string
		switch msg.Sender {
		case "You":
			sender = "Customer"
		default:
			sender = msg.Sender
		}
		msgs = append(msgs, sender+": "+msg.Text)
	}

	tools := []string{}
	toolNames := []string{}
	for name, desc := range toolMap {
		tools = append(tools, name+": "+desc)
		toolNames = append(toolNames, name)
	}

	var bts bytes.Buffer
	err := promtTemplate.Execute(&bts, struct {
		Tools     []string
		ToolNames []string
		History   []string
	}{
		Tools:     tools,
		ToolNames: toolNames,
		History:   msgs,
	})

	if err != nil {
		return err.Error()
	}

	return bts.String()
}
