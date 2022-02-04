package ui

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Console struct{}

// These could probably be collapsed. Not sure if we ever display the job
// without asking for something. Maybe another layer of abstraction, I dunno.
func (c *Console) DisplayText(text string) {
	os.Stdout.WriteString(text)
}

func (c *Console) DisplayNewLine() {
	os.Stdout.WriteString("\n")
}

func (c *Console) PromptForConfirmation(text string) bool {
	os.Stdout.WriteString(text + " ")
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Println("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press enter:")
		return c.PromptForConfirmation(text)
	}
}
