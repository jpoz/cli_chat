package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	clichat "github.com/jpoz/clichat/pkg"
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("Starting up...")

	msgChan := make(chan clichat.MessageContext, 100)
	backendChan := make(chan clichat.MessageContext, 100)
	p := tea.NewProgram(clichat.InitialModel(msgChan, backendChan))

	go clichat.NewAIClient(p, msgChan, backendChan).Run()
	go clichat.NewBackend(p, msgChan, backendChan).Run()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
