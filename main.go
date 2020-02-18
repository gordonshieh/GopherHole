package main

import "dns"

var blocklist []string = make([]string, 0)

func main() {
	blocklist = append(blocklist, "google.ca")
	dns.Server(blocklist)
}
