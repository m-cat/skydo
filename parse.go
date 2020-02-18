package main

import (
	"os"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	skynet "github.com/NebulousLabs/go-skynet"
)

func loadLists() (listList, error) {
	el := listList{}

	// Load user's file list if it exists.
	saved, err := ioutil.ReadFile(skydoFile)
	if err != nil {
		return el, nil
	}

	lists := []todoList{}
	lines := strings.Split(string(saved), "\n")

	if len(lines) < 2 {
		return el, errors.New("malformed save file")
	}

	// Get the last-opened file.
	current, err := strconv.Atoi(lines[0])
	if err != nil {
		return el, err
	}

	// Get the lists.
	for _, line := range lines[1:] {
		tokens := strings.SplitN(line, "\f", 2)
		if len(tokens) != 2 {
			return el, errors.New("malformed save file")
		}
		lists = append(lists, todoList{saved: true, name: tokens[0], skylink: tokens[1]})
	}

	if current >= len(lists) {
		return el, errors.New("malformed save file")
	}

	return listList{true, current, lists}, nil
}

func saveLists(ll *listList) error {
	saveData := strconv.Itoa(ll.current)

	for _, list := range ll.lists {
		// Use form feed as separator.
		saveData += "\n" + list.name + "\f" + list.skylink
	}

	err := ioutil.WriteFile(skydoFile, []byte(saveData), 0644)
	return err
}

func downloadList(name, skylink string) (todoList, error) {
	filename := skydoDir + name

	err := skynet.DownloadFile(filename, skylink, skynet.DefaultDownloadOptions)
	if err != nil {
		return todoList{}, err
	}

	rawList, err := ioutil.ReadFile(filename)
	if err != nil {
		return todoList{}, err
	}

	tl, err := parseList(string(rawList), skylink)
	if err != nil {
		return todoList{}, err
	}

	return tl, nil
}

func uploadList(tl *todoList) (string, error) {
	name := tl.name
	rawList := writeList(tl)
	filename := skydoDir + name

	err := os.MkdirAll(skydoDir, 0777)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filename, []byte(rawList), 0600)
	if err != nil {
		return "", err
	}

	// Upload new file
	skylink, err := skynet.UploadFile(filename, skynet.DefaultUploadOptions)
	if err != nil {
		return "", err
	}

	return skylink, nil
}

func parseList(rawList, skylink string) (todoList, error) {
	entries := []todoEntry{}
	lines := strings.Split(rawList, "\n")

	// fmt.Printf("%v\n", lines)

	// // Get rid of headers.
	// lines = lines[4:len(lines)-2]

	if len(lines) < 1 {
		return todoList{}, errors.New("malformed list")
	}

	name := lines[0]

	for _, line := range lines[1:] {
		entries = append(entries, todoEntry{text: line})
	}
	return todoList{true, name, skylink, entries}, nil
}

func writeList(tl *todoList) string {
	output := tl.name
	for _, entry := range tl.entries {
		output += "\n"+entry.text
	}
	return output
}
