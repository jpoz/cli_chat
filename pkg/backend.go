package clichat

import (
	"encoding/json"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

var orders = []order{
	{
		OrderNumber: "123456",
		Items: []orderItem{
			{
				ProductDesc: "Brown Pants",
				Quantity:    1,
			},
			{
				ProductDesc: "Purple shirt",
				Quantity:    1,
			},
		},
		OrderDate: "2021-01-01",
	},
	{
		OrderNumber: "654321",
		Items: []orderItem{
			{
				ProductDesc: "Green underwear",
				Quantity:    1,
			},
			{
				ProductDesc: "Red socks",
				Quantity:    1,
			},
		},
		OrderDate: "2023-04-10",
	},
}

var me = user{
	FirstName: "James",
	LastName:  "Pozdena",
	Email:     "jpozdena@gmail.com",
	Phone:     "5033480170",
}

type Backend struct {
	program     *tea.Program
	msgChan     chan MessageContext
	backendChan chan MessageContext
}

func NewBackend(p *tea.Program, msgChan chan MessageContext, backendChan chan MessageContext) *Backend {
	return &Backend{
		program:     p,
		msgChan:     msgChan,
		backendChan: backendChan,
	}
}

func (a *Backend) Run() {
	for msgCtx := range a.backendChan {
		log.Printf("Backend received message: %#v %d\n", msgCtx.Current, len(a.backendChan))
		if msgCtx.Current.Sender == "Action" {
			a.Chat(msgCtx.Current)
		}
	}
}

func (a *Backend) Chat(msg Message) {
	action := msg.Text
	input := msg.Input

	log.Println("Backend Action:", action)

	switch action {
	case "Chat":
		a.program.Send(AgentMsg{text: input})
	case "OrderSearch":
		if input == "jpozdena@gmail.com" {
			orderJson, err := json.MarshalIndent(orderResults{
				ResultsFor: "lookup_orders_by_email",
				LookupBy:   input,
				Orders:     orders,
			}, "", "  ")
			if err != nil {
				a.program.Send(errMsg(err))
				return
			}

			a.program.Send(BackendMsg{text: string(orderJson)})

			return
		}

		orderJson, err := json.MarshalIndent(orderResults{
			ResultsFor: "lookup_orders_by_email",
			LookupBy:   input,
			Orders:     []order{},
		}, "", "  ")
		if err != nil {
			a.program.Send(errMsg(err))
			return
		}

		a.program.Send(BackendMsg{text: string(orderJson)})
	case "lookup_orders_by_phone_number":
		if input == "5033480170" {
			orderJson, err := json.MarshalIndent(orderResults{
				ResultsFor: "lookup_orders_by_phone_number",
				LookupBy:   input,
				Orders:     orders,
			}, "", "  ")
			if err != nil {
				a.program.Send(errMsg(err))
				return
			}

			a.program.Send(BackendMsg{text: string(orderJson)})

			return
		}

		orderJson, err := json.MarshalIndent(orderResults{
			ResultsFor: "lookup_orders_by_phone_number",
			LookupBy:   input,
			Orders:     []order{},
		}, "", "  ")
		if err != nil {
			a.program.Send(errMsg(err))
			return
		}

		a.program.Send(BackendMsg{text: string(orderJson)})
	case "lookup_order_by_order_number":
		for _, o := range orders {
			if o.OrderNumber == input {
				orderJson, err := json.MarshalIndent(orderResults{
					ResultsFor: "lookup_order_by_order_number",
					LookupBy:   input,
					Orders:     []order{o},
				}, "", "  ")
				if err != nil {
					a.program.Send(errMsg(err))
					return
				}

				a.program.Send(BackendMsg{text: string(orderJson)})
				return
			}
		}
	case "ReturnOrderFlow":
		if input == "123456" {
			a.program.Send(BackendMsg{text: fmt.Sprintf(`Retrun instructions sent for order %s`, input)})
			return
		}
		a.program.Send(BackendMsg{text: fmt.Sprintf(`Invalid order number: %s`, input)})
	case "CloseConversation":
		a.program.Send(AgentMsg{text: "Goodbye"})
	case "EscalateToHuman":
		a.program.Send(AgentMsg{text: "I'll transfer you to a human agent"})
	}
}
