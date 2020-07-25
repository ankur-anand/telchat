package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"text/tabwriter"
	"time"
)

type optionType uint8

const (
	writeTimeout                = 10 * time.Second
	helpCommand                 = "/h"
	roomPrefix                  = "/room"
	clientPrefix                = "/client"
	clientOptionType optionType = iota
	roomOptionType
	msgOptionType
)

// formatDM format's the display message that include timestamp, name of the client and msg
func formatDM(name, room, msg string) string {
	return fmt.Sprintf("\0337\r \u001b[36m%s \u001b[35m%s\u001b[0m@\u001b[34m%s\u001b[0m \u001B[33m:\u001B[0m  %s\n\0338", time.Now().UTC().Format(time.Stamp), name, room, msg)
}

// formatCMDErr format's the display message that indicate the command err.
func formatCMDErr(cmd string) string {
	return fmt.Sprintf("\u001b[31m[Error]:\u001b[0m \u001b[34minvalid command\u001b[0m `%s`\n", cmd)
}

var (
	welcomeMsg        = "Hi There! Welcome to TELCHAT! Please Enter Your Chatter Name: \n>>"
	blankTime         = time.Time{}
	errInvalidCommand = errors.New("invalid command")
)

// disHelpCommand returns string output for the help command
func disHelpCommand() string {
	wr := new(bytes.Buffer)
	w := new(tabwriter.Writer)
	wr.WriteString("Thanks for Joining!. You can type /h for help anytime. Quick guide.\n\r")
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(wr, 0, 8, 4, '\t', 0)
	fmt.Fprintf(w, "\n SERIAL\tCOMMAND\tOPTION\tARGS\tDESCRIPTION")                                 // Header
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "------", "-------", "------", "----", "-----------") // row separator
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "1", "/room", "change", "[name]", "join to [name] room")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "2", "/room", "current", "", "display current room")                   //
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "3", "/client", "ignore", "[name]", "ignore [name] client's messages") //
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "4", "/client", "allow", "[name]", "allow [name] client's messages")   //
	err := w.Flush()
	if err != nil {
		panic(err)
	}
	wr.WriteString("\n\r")
	wr.WriteString("\nExamples\n\r")
	fmt.Fprintf(w, "\n %s\t%s\t", "1", "/room change myroom3")
	fmt.Fprintf(w, "\n %s\t%s\t", "2", "/room current")
	fmt.Fprintf(w, "\n %s\t%s\t", "3", "/client ignore annoyignone")
	fmt.Fprintf(w, "\n %s\t%s\t", "4", "/client allow annoyignore")
	err = w.Flush()
	if err != nil {
		panic(err)
	}
	wr.WriteString("\n\r")
	wr.WriteString("\nSend your typed message to the current room by entering enter")
	wr.WriteString("\n\r")
	return wr.String()
}
