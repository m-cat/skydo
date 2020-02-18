package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	binaryName = "Skydo"

	displayAmount = 5

	quitMsg = `
That's a cryin' shame.
`
	welcomeMsg = `
Welcome to ` + binaryName + `, the first todo list app on Skynet.
Type 'help' to see available commands.
`
)

var (
	helpMsg = strings.ReplaceAll(`
Available commands:

General:
- help
- quit
- save

Sharing:
- load [skylink]: \t open the given skylink
- share: \t display the skylink for the current list

Editing:
- add [entry]
- delete [position]
- insert [position] [entry]: \t add an entry at the given position
- move [src] [dest]: \t move the selected entry to the new position

Display:
- all: \t displays all available lists
- list: \t displays the entire current list
- ls: \t alias for 'list'

Lists:
- new [name]: \t create a new list
- open [name]: \t open the given list by name
- rename [name]: \t set a new name for the current list
`, `\t`, "\t")

	scanner = bufio.NewScanner(os.Stdin)

	skydoDir       = userHomeDir + "/.skydo/"
	skydoFile      = userHomeDir + "/.skydo.save"
	userHomeDir, _ = os.UserHomeDir()
)

func main() {
	fmt.Print(welcomeMsg)

	// Get the list of lists.
	ll := getLists()

	// Get the current todo list.
	tl := getTodoList(&ll)

	// Print the current todo list.
	printList(tl, 0, displayAmount*2)

	// Handle commands in a loop.
	for {
		if !tl.saved {
			fmt.Print("*")
		}
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		tokens := strings.SplitN(strings.TrimSpace(input), " ", 2)
		command := tokens[0]
		if command == "" {
			// Whitespace-only input.
			continue
		}
		// Remainder of the input. Can be further split up if a particular
		// command requires it.
		var args string
		if len(tokens) > 1 {
			args = tokens[1]
		} else {
			args = ""
		}

		var quit bool
		var err error
		tl, quit, err = handleCommand(tl, &ll, command, args)
		if quit {
			break
		}
		if err != nil {
			fmt.Println("\nerror:", err, "\n")
		}
		// fmt.Printf("\ntl: %v\nll: %v\n\n", tl, ll)
	}
}

func getLists() listList {
	ll, err := loadLists()
	if err != nil {
		fmt.Println("error:", err)
		return listList{}
	}

	return ll
}

func getTodoList(ll *listList) *todoList {
	if len(ll.lists) == 0 {
		return makeFirstList(ll)
	}

	cur := ll.current
	tl := ll.lists[cur]
	tl, err := downloadList(tl.name, tl.skylink)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	ll.lists[cur] = tl

	return &ll.lists[cur]
}

func makeFirstList(ll *listList) *todoList {
	name := ""
	for name == "" {
		fmt.Print("\nNo lists exist yet. Name your first list: ")
		scanner.Scan()
		name = scanner.Text()
		name = strings.TrimSpace(name)
	}
	tl := todoList{name: name}
	ll.lists = append(ll.lists, tl)
	return &ll.lists[len(ll.lists)-1]
}

func handleCommand(tl *todoList, ll *listList, command string, args string) (*todoList, bool, error) {
	var err error

	// TODO: auto command expansions, e.g. "ins" -> "insert"
	switch command {
	case "help":
		err = commandHelp(tl, ll, command, args)
	case "quit":
		quit := commandQuit(tl, ll, command, args)
		return tl, quit, nil
	case "save":
		err = commandSave(tl, ll, command, args)
	// Sharing
	case "load":
		tl, err = commandLoad(tl, ll, command, args)
	case "share":
		err = commandShare(tl, ll, command, args)
	// Editing
	case "add":
		err = commandAdd(tl, ll, command, args)
	case "delete":
		err = commandDelete(tl, ll, command, args)
	case "insert":
		err = commandInsert(tl, ll, command, args)
	case "move":
		err = commandMove(tl, ll, command, args)
	// Display
	case "all":
		err = commandAll(tl, ll, command, args)
	case "list":
		err = commandList(tl, ll, command, args)
	case "ls":
		err = commandList(tl, ll, command, args)
	// Lists
	case "new":
		tl, err = commandNew(tl, ll, command, args)
	case "open":
		tl, err = commandOpen(tl, ll, command, args)
	case "rename":
		err = commandRename(tl, ll, command, args)

	default:
		return tl, false, errors.New("unrecognized command")
	}

	if err != nil {
		return tl, false, err
	}

	return tl, false, nil
}

func commandHelp(tl *todoList, ll *listList, command string, args string) error {
	if args != "" {
		return errors.New("too many arguments to help")
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', 0)
	fmt.Fprintln(w, helpMsg)
	w.Flush()
	return nil
}

func commandQuit(tl *todoList, ll *listList, command string, args string) bool {
	if !ll.saved {
		fmt.Print("\nQuit with unsaved changes (y/N)? ")
		scanner.Scan()
		input := scanner.Text()
		input = strings.TrimSpace(input)
		if strings.ToLower(input) != "y" {
			fmt.Println()
			return false
		}
	}

	// Ignore errors here.
	_ = saveLists(ll)

	fmt.Println(quitMsg)
	return true
}

func commandSave(tl *todoList, ll *listList, command string, args string) error {
	if args != "" {
		return errors.New("too many arguments to help")
	}

	for i := range ll.lists {
		tl := &ll.lists[i]
		if tl.saved {
			continue
		}

		// Upload the list.
		skylink, err := uploadList(tl)
		if err != nil {
			return err
		}

		// Update the skylink.
		tl.skylink = skylink
		tl.saved = true
	}

	// Write the save file.
	err := saveLists(ll)
	if err != nil {
		return err
	}

	ll.saved = true

	return nil
}

func commandLoad(tl *todoList, ll *listList, command string, args string) (*todoList, error) {
	argsv := strings.Split(args, " ")
	if len(argsv) != 1 {
		return tl, errors.New("1 argument expected for " + command)
	}
	skylink := argsv[0]

	// See if we already have the list.
	for i, list := range ll.lists {
		if list.skylink == skylink {
			return openList(ll, i), nil
		}
	}

	tl2, err := downloadList("load", skylink)
	if err != nil {
		return tl, err
	}
	tl = &tl2

	printList(tl, 0, displayAmount)
	return tl, nil
}

func commandShare(tl *todoList, ll *listList, command string, args string) error {
	fmt.Println()

	if tl.saved {
		fmt.Println(tl.skylink)
	} else {
		if tl.skylink != "" {
			fmt.Println(tl.skylink)
			fmt.Println("warning: this list hasn't been saved")
		} else {
			fmt.Println("skylink doesn't exist (try saving first)")
		}
	}

	fmt.Println()

	return nil
}

func commandAdd(tl *todoList, ll *listList, command string, args string) error {
	if args == "" {
		return errors.New("not enough arguments to " + command)
	}
	tl.entries = append(tl.entries, todoEntry{args, false})
	tl.saved = false
	ll.saved = false

	l := len(tl.entries)
	printListAtIndex(tl, l-1)
	return nil
}

func commandDelete(tl *todoList, ll *listList, command string, args string) error {
	l := len(tl.entries)
	if args == "" {
		return errors.New("not enough arguments to " + command)
	}
	position, err := strconv.Atoi(args)
	if err != nil {
		return err
	}
	if position < 0 || position >= l {
		return errors.New("cannot delete from position " + args)
	}

	fmt.Printf("\nAre you sure you want to delete entry %v (y/N)? ", position)
	scanner.Scan()
	input := scanner.Text()
	input = strings.TrimSpace(input)
	if strings.ToLower(input) != "y" {
		fmt.Println()
		return nil
	}

	deleteEntry(tl, position)
	tl.saved = false
	ll.saved = false

	printListAtIndex(tl, position)
	return nil
}

func commandInsert(tl *todoList, ll *listList, command string, args string) error {
	l := len(tl.entries)
	argsv := strings.SplitN(args, " ", 2)
	if len(argsv) < 2 {
		return errors.New("not enough arguments to " + command)
	}
	position, err := strconv.Atoi(argsv[0])
	if err != nil {
		return err
	}
	if position < 0 || position >= l {
		return errors.New("cannot insert at position " + argsv[0])
	}
	note := argsv[1]

	insertEntry(tl, todoEntry{note, false}, position)
	tl.saved = false
	ll.saved = false

	printListAtIndex(tl, position)
	return nil
}

func commandMove(tl *todoList, ll *listList, command string, args string) error {
	l := len(tl.entries)
	argsv := strings.Split(args, " ")
	if len(argsv) != 2 {
		return errors.New("2 arguments expected for " + command)
	}
	src, err := strconv.Atoi(argsv[0])
	if err != nil {
		return err
	}
	if src < 0 || src >= l {
		return errors.New("cannot move from position " + argsv[0])
	}
	dest, err := strconv.Atoi(argsv[1])
	if err != nil {
		return err
	}
	if dest < 0 || dest >= l {
		return errors.New("cannot move to position " + argsv[1])
	}

	entry := tl.entries[src]
	deleteEntry(tl, src)
	insertEntry(tl, entry, dest)
	tl.saved = false
	ll.saved = false

	printListAtIndex(tl, dest)
	return nil
}

func commandAll(tl *todoList, ll *listList, command string, args string) error {
	if args != "" {
		return errors.New("too many arguments to " + command)
	}
	printLists(ll)
	return nil
}

func commandList(tl *todoList, ll *listList, command string, args string) error {
	if args != "" {
		return errors.New("too many arguments to " + command)
	}
	printList(tl, 0, len(tl.entries))
	return nil
}

func commandNew(tl *todoList, ll *listList, command string, args string) (*todoList, error) {
	name := args

	for _, list := range ll.lists {
		if list.name == name {
			return tl, errors.New("a list named '" + name + "' already exists")
		}
	}

	list := todoList{name: name, entries: []todoEntry{}}
	ll.lists = append(ll.lists, list)
	i := len(ll.lists) - 1
	ll.current = i
	tl = &ll.lists[i]
	tl.saved = false
	ll.saved = false

	printList(tl, 0, displayAmount)
	return tl, nil
}

func commandOpen(tl *todoList, ll *listList, command string, args string) (*todoList, error) {
	name := args

	for i, list := range ll.lists {
		if list.name == name {
			return openList(ll, i), nil
		}
	}

	return tl, errors.New("no list named '" + name + "'")
}

func commandRename(tl *todoList, ll *listList, command string, args string) error {
	name := args

	tl.name = name
	tl.saved = false
	ll.saved = false

	printList(tl, 0, displayAmount)
	return nil
}
