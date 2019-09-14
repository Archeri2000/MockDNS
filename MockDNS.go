package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Windows DNS Hosts file
var hosts = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"

// Appdata changes
var exe, _ = os.Executable()
var path = filepath.Dir(exe)
var changes = path + "\\changes"

func main() {
	// Create Appdata changes if it doesn't exist
	if _, err := os.Stat(changes); os.IsNotExist(err) {
		err = ioutil.WriteFile(changes, []byte{}, 0777)
		if err != nil {
			fmt.Println("Unable to create changes file!")
		}
	}

	// Get command args
	Args := os.Args[1:]
	if len(Args) >= 1 {
		// Get command keyword
		keyword := strings.ToLower(Args[0])
		switch keyword {
		case "add", "remove":
			if len(Args) == 2 {
				// Get filename
				filename := Args[1]
				// Check if filename is a valid file
				if isValidFilepath(filename) {
					switch keyword {
					// Adds to DNS
					case "add":
						AddToDNS(filename)
					// Removes from DNS
					case "remove":
						RemoveFromDNS(filename)
					}
				} else {
					fmt.Println("Invalid filepath!")
				}
			} else {
				fmt.Println("Invalid number of arguments!")
			}
		case "add-now", "remove-now":
			if len(Args) == 3 {
				str := ParseStrings([]string{strings.Join(Args[1:3], " ")})
				if len(str) == 1 {
					switch keyword {
					case "add-now":
						result, writeErr := CompareAndWrite(hosts, str, false, false)
						if writeErr == nil {
							_, _ = CompareAndWrite(changes, result, false, true)
						}
					case "remove-now":
						_, writeErr := CompareAndWrite(hosts, str, true, false)
						if writeErr == nil {
							_, _ = CompareAndWrite(changes, str, true, true)
						}
					}
				}
			} else {
				fmt.Println("Invalid number of arguments!")
			}
		case "reset", "show", "show-all":
			if len(Args) == 1 {
				switch keyword {
				// Removes all appdata changes from DNS host file
				case "reset":
					ResetDNS()
				// Shows all mocked DNS entries
				case "show":
					ShowMocked()
				// Displays the contents of the DNS host file
				case "show-all":
					ShowDNS()
				}
			} else {
				fmt.Println(keyword + " cannot have any other arguments!")
			}
		default:
			fmt.Println("Invalid method! Method can only be add, remove, reset, show, show-all!")
		}
	} else {
		fmt.Println("No Arguments detected! Please supply a method! Method can be add, remove, reset, show, show-all!")
	}
}

// Adds to DNS
func AddToDNS(filename string) {
	file, err := FormatFileContents(filename)
	if err == nil {
		fmt.Println("Adding DNS entries from: " + filename + "...")
		// Attempts to add differing entries to DNS
		addedEntries, writeErr := CompareAndWrite(hosts, file, false, false)
		if writeErr == nil {
			// Writes the added entries to the appdata changes if successful
			_, _ = CompareAndWrite(changes, addedEntries, false, true)
		}
	}
}

// Remove from DNS
func RemoveFromDNS(filename string) {
	file, err := FormatFileContents(filename)
	if err == nil {
		fmt.Println("Removing DNS entries from: " + filename + "...")
		// Attempts to remove entries from DNS
		_, writeErr := CompareAndWrite(hosts, file, true, false)
		if writeErr == nil {
			// Remove said entries from appdata changes if successful
			_, _ = CompareAndWrite(changes, file, true, true)
		}
	}
}

// Resets the DNS
func ResetDNS() {
	fmt.Println("Resetting DNS...")
	// Removes appdata changes from DNS
	RemoveFromDNS(changes)
}

// Show Mocked Entries
func ShowMocked() {
	contents, err := FileToList(changes)
	if err == nil {
		contents = StripEmptyAndComments(contents)
		if len(contents) > 0 {
			fmt.Println("Mocked DNS Entries:")
			fmt.Println(ArrayToString(contents))
		} else {
			fmt.Println("No DNS Entries are currently mocked!")
		}
	} else {
		fmt.Println("Could not read existing changes!")
	}
}

// Show Full DNS
func ShowDNS() {
	contents, err := FileToList(hosts)
	if err == nil {
		fmt.Println(ArrayToString(contents))
	} else {
		fmt.Println("Could not read existing changes!")
	}
}

// Compares a file of strings with a string array and performs a write operation
func CompareAndWrite(filenameTO string, from []string, remove bool, suppressPrint bool) ([]string, error) {
	// Retrieve string list from destination file
	to, err := FileToList(filenameTO)
	// Check for errors in opening file
	if err == nil {
		// Check operation type
		if remove {
			// REMOVE operation
			// Loop over destination to remove strings that exist in from
			for i := 0; i < len(to); i++ {
				for j := 0; j < len(from); j++ {
					if to[i] == from[j] {
						to = Remove(to, i)
						if !suppressPrint {
							fmt.Println("REMOVE: " + from[j])
						}
						i -= 1
						break
					}
				}
			}
			// Rewrite entire destination file with what is left
			err := ioutil.WriteFile(filenameTO, []byte(strings.TrimRight(ArrayToString(to), "\n")), 0777)
			// Check for errors
			if err == nil {
				if !suppressPrint {
					fmt.Println("Successfully removed elements from DNS")
				}
				return []string{}, nil
			} else {
				fmt.Println("REMOVE was unsuccessful...")
				return nil, err
			}

		} else {
			// ADD Operation
			// Loop over destination to find un-added entries from FROM
			for i := 0; i < len(to); i++ {
				for j := 0; j < len(from); j++ {
					if to[i] == from[j] {
						from = Remove(from, j)
						j -= 1
					}
				}
			}
			if !suppressPrint {
				for k := 0; k < len(from); k++ {
					fmt.Println("ADD: " + from[k])
				}
			}

			// Append un-added entries from FROM
			f, _ := os.OpenFile(filenameTO, os.O_APPEND|os.O_WRONLY, 0644)
			defer f.Close()

			_, err := f.WriteString("\n" + ArrayToString(from))

			// Check for errors
			if err == nil {
				if !suppressPrint {
					fmt.Println("Successfully added elements to DNS")
				}
				return from, nil
			} else {
				fmt.Println("ADD was unsuccessful...")
				return nil, err
			}
		}
	}
	return nil, err
}

// Opens file to add and formats its input
func FormatFileContents(filename string) ([]string, error) {
	//Attempt to read file
	file, err := FileToList(filename)
	if err == nil {
		// Parse strings
		file = ParseStrings(file)
		return file, nil
	} else {
		fmt.Println("File format is invalid!")
		return nil, err
	}
}

// Checks if a string is a valid filepath
func isValidFilepath(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Converts a filename to an array of strings based on \n separator
func FileToList(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		return strings.Split(string(content), "\n"), nil
	} else {
		return nil, err
	}
}

// Converts an array of strings to a single string separated with \n
func ArrayToString(arr []string) string {
	return strings.Join(arr, "\n")
}

// Removes a given index from an array of strings
func Remove(arr []string, ind int) []string {
	if ind >= len(arr) {
		return arr
	}
	return append(arr[:ind], arr[ind+1:]...)
}

// Strips empty array entries from string array
func StripEmptyAndComments(arr []string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == "" || string(arr[i][0]) == "#" {
			arr = Remove(arr, i)
			i -= 1
		}
	}
	return arr
}

// Parse array of strings to replace localhost with 127.0.0.1
func ParseStrings(entry []string) []string {
	entry = StripEmptyAndComments(entry)
	for i := 0; i < len(entry); i++ {
		entry[i] = strings.ReplaceAll(entry[i], "localhostv6", "::1")
		entry[i] = strings.ReplaceAll(entry[i], "localhost", "127.0.0.1")
		if ent := strings.Split(entry[i], " "); len(ent) != 2 || !isIP(ent[0]) {
			fmt.Println(entry[i] + " omitted due to invalid format!")
			entry = Remove(entry, i)
			i -= 1
		}
	}
	return entry
}

func isIP(str string) bool {
	str = strings.ToLower(str)
	isIPv4, _ := regexp.MatchString(`^\d+.\d+.\d+.\d+$`, str)
	isIPv6, _ := regexp.MatchString(`^([a-f\d]*:)+\d*$`, str)
	return isIPv4 || isIPv6
}
