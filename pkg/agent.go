package clichat

import (
	"context"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

type AIClient struct {
	program         *tea.Program
	client          *openai.Client
	msgChan         chan MessageContext
	lastUserMessage string
}

func NewAIClient(p *tea.Program, msgChan chan MessageContext, backendChan chan MessageContext) *AIClient {
	return &AIClient{
		program: p,
		client:  openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		msgChan: msgChan,
	}
}

func (a *AIClient) Run() {
	for msgCtx := range a.msgChan {
		if msgCtx.Current.Sender == "You" || msgCtx.Current.Sender == "Backend" {
			a.Chat(msgCtx)
		}
	}
}

func (a *AIClient) Chat(msgCtx MessageContext) {
	lastUserMessage := getLastUserMessage(msgCtx.History)
	if lastUserMessage == a.lastUserMessage {
		return
	}

	a.lastUserMessage = lastUserMessage

	prmt := GenerateConvesationalPrompt(tools, msgCtx.History)

	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			// Model: openai.GPT3Dot5Turbo,
			Model:       openai.GPT4,
			Temperature: 0.3,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prmt,
				},
			},
		},
	)

	log.Printf("AI prompt: %q", prmt)

	if err != nil {
		a.program.Send(errMsg(err))
	}

	aiResp := resp.Choices[0].Message.Content
	log.Printf("AI response: %q", aiResp)

	nextAction := ParseResponse(aiResp)
	log.Printf("AI next action: %#v", nextAction)

	msgs := Messages{}

	if nextAction.Agent != "" {
		msgs = append(msgs, Message{
			Sender: "Agent",
			Text:   nextAction.Agent,
		})
	}

	if nextAction.Action != "" {
		msgs = append(msgs, Message{
			Sender: "Action",
			Text:   nextAction.Action,
			Input:  nextAction.ActionInput,
		})
	}

	if nextAction.Thought != "" {
		msgs = append(msgs, Message{
			Sender: "Thought",
			Text:   nextAction.Thought,
		})
	}

	log.Printf("Sending messages: %#v", msgs)

	a.program.Send(msgs)
}

func getLastUserMessage(msgs []Message) string {
	out := ""
	for _, msg := range msgs {
		if msg.Sender == "You" {
			out = msg.Text
		}
	}
	return out
}
