package pkg

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"text/tabwriter"
	"time"
)

type optionType uint8

const (
	writeTimeout                = 10 * time.Second
	helpCommand                 = "/h"
	infoCommand                 = "/info"
	roomPrefix                  = "/room"
	clientPrefix                = "/client"
	clientOptionType optionType = iota
	roomOptionType
	msgOptionType
)

// formatDM format's the display message that include timestamp, name of the client and msg in terminal format
func formatDM(name, room, msg string) string {
	return fmt.Sprintf("\0337\r \u001b[36m%s \u001b[35m%s\u001b[0m@\u001b[34m%s\u001b[0m \u001B[33m:\u001B[0m  %s\n\0338", time.Now().UTC().Format(time.Stamp), name, room, msg)
}

// formatCMDErr format's the display message that indicate the command err in terminal format.
func formatCMDErr(cmd string) string {
	return fmt.Sprintf("\u001b[31m[Error]:\u001b[0m \u001b[34minvalid command\u001b[0m `%s`\n", cmd)
}

// infoDisplay decorate the name and room information in terminal format
func infoDisplay(name, room string) string {
	return fmt.Sprintf("\u001B[35m%s\u001B[0m: \u001B[34m[%s]\u001B[0m \n\r", name, room)
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
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "1", "/info", "", "", "display username & current room")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "2", "/room", "change", "[name]", "join to [name] room")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "3", "/client", "ignore", "[name]", "ignore [name] client's messages") //
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t%s\t", "4", "/client", "allow", "[name]", "allow [name] client's messages")   //
	err := w.Flush()
	if err != nil {
		panic(err)
	}
	wr.WriteString("\n\r")
	wr.WriteString("\nExamples\n\r")
	fmt.Fprintf(w, "\n %s\t%s\t", "1", "/info")
	fmt.Fprintf(w, "\n %s\t%s\t", "2", "/room change myroom3")
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

// msgWriter writes given msg to the connection with write deadline
func msgWriter(conn net.Conn, msg string) error {
	err := conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		log.Printf("SetWriteDeadline failed: %v\n", err)
		return err
	}
	_, err = io.WriteString(conn, msg)
	if err != nil {
		log.Println("conn write failed, err: ", err)
		return err
	}
	err = conn.SetWriteDeadline(blankTime)
	if err != nil {
		log.Printf("SetWriteDeadline failed: %v\n", err)
		return err
	}
	return nil
}

func commandType(m string) optionType {
	if strings.HasPrefix(m, clientPrefix) {
		return clientOptionType
	}

	if strings.HasPrefix(m, roomPrefix) {
		return roomOptionType
	}

	return msgOptionType
}

// telnetHandler handles the accepted telnet connection's
type telnetHandler struct {
	mWriter   io.Writer
	chatStore *chatDataStore
	helpDMsg  string
	hook      func() // hook is a test noop in live code
}

func newTelnetS(lw io.Writer) *telnetHandler {
	return &telnetHandler{
		mWriter:   lw,
		chatStore: newChatDataStore(lw),
		helpDMsg:  disHelpCommand(),
		hook:      func() {}, // noop function
	}
}

// cmdErrWriter writes error in formatted form when any wrong command is provided.
func (ts *telnetHandler) cmdErrWriter(conn net.Conn, cmd string) error {
	err := msgWriter(conn, formatCMDErr(cmd))
	if err != nil {
		return err
	}
	return errInvalidCommand
}

// infoPrompt writes the information back to user when requested
func (ts *telnetHandler) infoPrompt(conn net.Conn, name, room string) error {
	return msgWriter(conn, infoDisplay(name, room))
}

func (ts *telnetHandler) displayHelp(conn net.Conn, name, room string) error {
	err := msgWriter(conn, ts.helpDMsg)
	if err != nil {
		return err
	}
	return ts.infoPrompt(conn, name, room)
}

// roomCommandOps handles all room command operation
func (ts *telnetHandler) roomCommandOps(conn net.Conn, cmd, name string, roomName *string) error {
	cmds := strings.Split(cmd, " ")
	if len(cmds) != 3 {
		return ts.cmdErrWriter(conn, cmd)
	}
	option := strings.TrimSpace(cmds[1])
	arg := strings.TrimSpace(cmds[2])
	switch option {
	case "change": // change
		if len(arg) == 0 {
			return ts.cmdErrWriter(conn, cmd)
		}
		// remove from current room
		ts.chatStore.removeClientFromRoom(name, *roomName)
		// add the client to the new room
		ts.chatStore.addClientToRoom(name, arg)
		*roomName = arg
		return ts.infoPrompt(conn, name, *roomName)
	default:
		return ts.cmdErrWriter(conn, cmd)
	}
}

func (ts *telnetHandler) clientCommandOps(conn net.Conn, name, cmd string) error {
	cmds := strings.Split(cmd, " ")
	if len(cmds) != 3 {
		return ts.cmdErrWriter(conn, cmd)
	}
	option := strings.TrimSpace(cmds[1])
	arg := strings.TrimSpace(cmds[2])
	switch option {
	case "ignore": // ignore
		if len(arg) == 0 {
			return ts.cmdErrWriter(conn, cmd)
		}
		// add to the ignore list.
		ts.chatStore.ignoreNamedClient(name, arg)
		ts.hook()
	case "allow": // allow
		if len(arg) == 0 {
			return ts.cmdErrWriter(conn, cmd)
		}
		// remove from the ignore list.
		ts.chatStore.allowNamedClient(name, arg)
		ts.hook()
	default:
		return ts.cmdErrWriter(conn, cmd)
	}
	return nil
}

// serveConn serve all of the net.Conn
func (ts *telnetHandler) serveConn(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println("conn close failed, err: ", err)
		}
	}()
	// Welcome user on the screen.
	err := msgWriter(conn, welcomeMsg)
	if err != nil {
		log.Println("unable to welcome user on screen, err: ", err)
		return
	}

	// split read each line from conn
	connScan := bufio.NewScanner(conn)
	var name string
	// split scan on new line
	// get user name

	// clientReg marks that this client has been registered
	// this prevent the case when the client terminate the connection
	// before it get registered with the store,
	clientReg := false
	for connScan.Scan() {
		if err := connScan.Err(); err != nil {
			log.Println("username scan failed", err)
			return
		}
		name = connScan.Text()
		if name == "" {
			err = msgWriter(conn, "name cannot be empty \n>>")
			if err != nil {
				log.Println("conn write failed, err: ", err)
				return
			}
			continue
		}

		// if name is already taken ask for new name.
		if err := ts.chatStore.registerClient(name, conn); err != nil {
			err = msgWriter(conn, fmt.Sprintf("name %s Taken, try new name \n>>", name))
			if err != nil {
				log.Println("conn write failed, err: ", err)
				return
			}
			continue
		}
		clientReg = true
		break
	}
	if !clientReg {
		return
	}
	defer ts.chatStore.deleteClient(name)
	currentRoom := metaRoom
	err = ts.displayHelp(conn, name, currentRoom)
	if err != nil {
		return
	}
	log.Printf("new client connected name: %s, remoteAddr: %s", name, conn.RemoteAddr())
	defer func() {
		log.Printf("client disconnected name: %s, remoteAddr: %s", name, conn.RemoteAddr())
	}()
	for connScan.Scan() {
		if err := connScan.Err(); err != nil {
			log.Println("conn state err", err)
			return
		}
		command := strings.TrimSpace(connScan.Text())
		switch command {
		// single command
		case helpCommand:
			err := ts.displayHelp(conn, name, currentRoom)
			if err != nil {
				return
			}
		case infoCommand:
			err := ts.infoPrompt(conn, name, currentRoom)
			if err != nil {
				return
			}
		default:
			// check if command query
			switch commandType(command) {
			case msgOptionType:
				ts.chatStore.broadcastMsg(context.TODO(), name, currentRoom, []byte(formatDM(name, currentRoom, command)))
				ts.logWriter(command)
			case roomOptionType:
				err := ts.roomCommandOps(conn, command, name, &currentRoom)
				if err != nil && !errors.Is(err, errInvalidCommand) {
					return
				}
			case clientOptionType:
				err := ts.clientCommandOps(conn, name, command)
				if err != nil && !errors.Is(err, errInvalidCommand) {
					return
				}
			}
		}
	}
}

func (ts *telnetHandler) logWriter(command string) {
	_, err := ts.mWriter.Write([]byte(command + "\n\r")) // write message to the log file
	if err != nil {
		log.Println("error writing message to the log file")
	}
}
