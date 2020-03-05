package main

import (
	"api"
	"blocklist"
	"dns"
	"io/ioutil"
	"net/http"
	"strings"
)

func importBlocklistFromHTTP(db *blocklist.Blocklist) {
	for _, url := range db.GetBlocklists() {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		toBlock := strings.Split(string(body), "\n")
		for _, host := range toBlock {
			println(host)
			db.AddHost(host)
		}
	}
}

func main() {
	db := blocklist.GetDatabase()
	blockStreamer := make(chan blocklist.HistoryEntry)
	db.AddBlocklist("https://raw.githubusercontent.com/hectorm/hmirror/master/data/adaway.org/list.txt")
	go importBlocklistFromHTTP(db)
	println("server started")
	go dns.Server(db, blockStreamer)
	api.StartAPIServer(db, blockStreamer)
}
