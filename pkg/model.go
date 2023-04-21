package clichat

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type Model struct {
	viewport         viewport.Model
	internalViewport viewport.Model

	messages    []Message
	messageChan chan MessageContext
	backendChan chan MessageContext

	textarea      textarea.Model
	agentTextarea textarea.Model

	senderStyle  lipgloss.Style
	agentStyle   lipgloss.Style
	aiStyle      lipgloss.Style
	backendStyle lipgloss.Style
	thoughtStyle lipgloss.Style

	err error
	gr  *glamour.TermRenderer
}

func InitialModel(messageChan chan MessageContext, backendChan chan MessageContext) Model {
	ta := textarea.New()
	ta.Placeholder = "Send a user message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(VIEW_WIDTH)
	ta.SetHeight(3)
	// ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	ata := textarea.New()
	ata.Placeholder = "Send an agent message..."
	ata.Prompt = "┃ "
	ata.CharLimit = 280
	ata.SetWidth(VIEW_WIDTH)
	ata.SetHeight(3)
	// ata.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ata.ShowLineNumbers = false
	ata.KeyMap.InsertNewline.SetEnabled(false)
	ata.Blur()

	vp := viewport.New(VIEW_WIDTH, HEIGHT)
	vp.SetContent(`Type a message as a user to get started.    	      `)

	ivp := viewport.New(VIEW_WIDTH, HEIGHT)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(WIDTH),
	)
	if err != nil {
		panic(err)
	}

	return Model{
		textarea:         ta,
		agentTextarea:    ata,
		messages:         []Message{},
		messageChan:      messageChan,
		backendChan:      backendChan,
		viewport:         vp,
		internalViewport: ivp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		agentStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		aiStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		backendStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		thoughtStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		err:              nil,
		gr:               renderer,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd  tea.Cmd
		vpCmd  tea.Cmd
		ivpCmd tea.Cmd
	)

	if m.textarea.Focused() {
		m.textarea, tiCmd = m.textarea.Update(msg)
	} else {
		m.agentTextarea, tiCmd = m.agentTextarea.Update(msg)
	}
	m.internalViewport, ivpCmd = m.internalViewport.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// log.Printf("Model.Update: %#v", msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyTab:
			if m.textarea.Focused() {
				m.textarea.Blur()
				m.agentTextarea.Focus()
			} else {
				m.agentTextarea.Blur()
				m.textarea.Focus()
			}
		case tea.KeyEnter:
			var outMsg Message

			if m.textarea.Focused() {
				outMsg = Message{
					Sender: "You",
					Text:   m.textarea.Value(),
				}
				m.textarea.Reset()
			} else {
				outMsg = Message{
					Sender: "Agent",
					Text:   m.agentTextarea.Value(),
				}
				m.agentTextarea.Reset()
			}

			m.messages = append(m.messages, outMsg)
			m.messageChan <- MessageContext{
				Current: outMsg,
				History: m.messages,
			}

			m.viewport.SetContent(m.messageContent())
			m.viewport.GotoBottom()
		}
	case errMsg:
		m.err = msg
		return m, nil
	case Messages:
		log.Printf("Received messages: %v", msg)
		for _, a := range msg {
			log.Printf("Processing message: %v", a)
			m.messages = append(m.messages, a)
			m.messageChan <- MessageContext{
				Current: a,
				History: m.messages,
			}
			m.backendChan <- MessageContext{
				Current: a,
				History: m.messages,
			}
		}
		m.internalViewport.SetContent(m.internalContent())
		m.internalViewport.GotoBottom()
		m.viewport.SetContent(m.messageContent())
		m.viewport.GotoBottom()
	case AIMsg:
		outMsg := Message{
			Sender: "AI",
			Text:   msg.text,
		}
		m.messages = append(m.messages, outMsg)
		m.backendChan <- MessageContext{
			Current: outMsg,
			History: m.messages,
		}
		m.internalViewport.SetContent(m.internalContent())
		m.internalViewport.GotoBottom()
	case BackendMsg:
		outMsg := Message{
			Sender: "Backend",
			Text:   msg.text,
		}
		m.messages = append(m.messages, outMsg)
		m.messageChan <- MessageContext{
			Current: outMsg,
			History: m.messages,
		}
		m.internalViewport.SetContent(m.internalContent())
		m.internalViewport.GotoBottom()
	case AgentMsg:
		m.messages = append(m.messages, Message{
			Sender: "Agent",
			Text:   msg.text,
		})
		m.viewport.SetContent(m.messageContent())
		m.viewport.GotoBottom()
	}

	// log.Printf("Model.Returning:\n\t%#v\n\t%#v\n\t%#v", tiCmd, vpCmd, ivpCmd)

	return m, tea.Batch(tiCmd, vpCmd, ivpCmd)
}

func (m Model) messageContent() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		if msg.Sender == "You" || msg.Sender == "Agent" {
			var ssb strings.Builder
			if msg.Sender == "You" {
				ssb.WriteString(m.senderStyle.Render(msg.Sender))
			} else {
				ssb.WriteString(m.agentStyle.Render(msg.Sender))
			}
			ssb.WriteString(": ")
			str, err := m.gr.Render(msg.Text)
			if err != nil {
				ssb.WriteString("error rendering message")
			}
			ssb.WriteString(str)

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		}
	}

	return sb.String()
}

func (m Model) internalContent() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		if msg.Sender == "AI" {
			var ssb strings.Builder
			ssb.WriteString(m.aiStyle.Render(msg.Sender))
			ssb.WriteString(": ")
			str, err := m.gr.Render(msg.Text)
			if err != nil {
				ssb.WriteString("error rendering message")
			}
			ssb.WriteString(str)

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		} else if msg.Sender == "Action" {
			var ssb strings.Builder
			ssb.WriteString(m.backendStyle.Render(msg.Sender))
			ssb.WriteString(": ")
			ssb.WriteString(msg.Text)
			ssb.WriteString("\n")
			ssb.WriteString(m.backendStyle.Render("ActionInput"))
			ssb.WriteString(": ")
			ssb.WriteString(msg.Input)
			ssb.WriteString("\n")

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		} else if msg.Sender == "Agent" {
			var ssb strings.Builder
			ssb.WriteString(m.aiStyle.Render(msg.Sender))
			ssb.WriteString(": ")
			str, err := m.gr.Render(msg.Text)
			if err != nil {
				ssb.WriteString("error rendering message")
			}
			ssb.WriteString(str)

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		} else if msg.Sender == "Backend" {
			var ssb strings.Builder
			ssb.WriteString(m.backendStyle.Render("Observation"))
			ssb.WriteString(": ")
			str, err := m.gr.Render(msg.Text)
			if err != nil {
				ssb.WriteString("error rendering message")
			}
			ssb.WriteString(str)

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		} else if msg.Sender == "Thought" {
			var ssb strings.Builder
			ssb.WriteString(m.thoughtStyle.Render(msg.Sender))
			ssb.WriteString(": ")
			ssb.WriteString(msg.Text)

			sb.WriteString(wordwrap.String(ssb.String(), VIEW_WIDTH))
		}

	}

	return sb.String()
}

func (m Model) View() string {
	out := fmt.Sprintf(
		"%s\n\n%s",
		lipgloss.JoinHorizontal(lipgloss.Top, m.viewport.View(), m.internalViewport.View()),
		lipgloss.JoinHorizontal(lipgloss.Top, m.textarea.View(), m.agentTextarea.View()),
	) + "\n\n"
	if m.err != nil {
		out += m.err.Error()
	}
	return out
}
