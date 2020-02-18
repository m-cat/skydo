package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type (
	listList struct {
		saved   bool
		current int
		lists   []todoList
	}
	todoEntry struct {
		text string
		// TODO: Implement marking entries as complete
		complete bool
	}
	todoList struct {
		saved         bool
		name, skylink string
		entries       []todoEntry
	}
)

func printLists(ll *listList) {
	fmt.Println()
	for _, list := range ll.lists {
		if list.saved {
			fmt.Print("-  ")
		} else {
			fmt.Print("-* ")
		}
		fmt.Println(list.name)
	}
	fmt.Println()
}

// printList prints the list from the specified indices.
func printList(tl *todoList, i1, i2 int) {
	fmt.Println()
	fmt.Println(tl.name)

	l := len(tl.entries)
	if l == 0 {
		fmt.Println("(The list is empty)\n")
		return
	}

	if i1 < 0 {
		i1 = 0
	}
	if i2 > l {
		i2 = l
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', 0)

	if i1 > 0 {
		fmt.Fprintln(w, "[...]")
	}

	for i := i1; i < i2; i++ {
		fmt.Fprintf(w, "%v: \t %v\n", i, tl.entries[i].text)
	}

	if i2 < len(tl.entries) {
		fmt.Fprintln(w, "[...]")
	}
	fmt.Fprintln(w)

	w.Flush()
}

func printListAtIndex(tl *todoList, i int) {
	window := displayAmount / 2
	i1, i2 := i-window, i+window+1

	if i2 > len(tl.entries) {
		window = i2 - len(tl.entries)
		i1 -= window
		i2 -= window
	}
	if i1 < 0 {
		window = -i1
		i1 += window
		i2 += window
	}

	printList(tl, i1, i2)
}

func deleteEntry(tl *todoList, i int) {
	tl.entries = append(tl.entries[:i], tl.entries[i+1:]...)
}

func insertEntry(tl *todoList, entry todoEntry, i int) {
	tl.entries = append(tl.entries, todoEntry{})
	copy(tl.entries[i+1:], tl.entries[i:])
	tl.entries[i] = entry
}

func openList(ll *listList, i int) *todoList {
	ll.current = i
	tl := &ll.lists[i]

	printList(tl, 0, displayAmount)
	return tl
}
