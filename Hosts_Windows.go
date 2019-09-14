// +build windows

package main

import "os"

// Windows DNS Hosts file
var hosts = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
