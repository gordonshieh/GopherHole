package blocklist

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Blocklist struct {
	_db *sql.DB
}

// GetDatabase Get the database access object
// Initializes the db if not exists in the same cwd of the exectuable
func GetDatabase() *Blocklist {
	database, _ := sql.Open("sqlite3", "data.db")
	blocklistStatement, _ := database.Prepare(`CREATE TABLE IF NOT EXISTS
											   blocklist (hostname TEXT PRIMARY KEY,
														  ip TEXT)`)
	blocklistStatement.Exec()

	sourcesStatement, _ := database.Prepare(`CREATE TABLE IF NOT EXISTS
											 sources (id INTEGER PRIMARY KEY AUTOINCREMENT, 
											  		  source TEXT UNIQUE)`)
	sourcesStatement.Exec()
	// type is an int because gopacket.dns.DNSType is defined as uint16
	historyStatement, _ := database.Prepare(`CREATE TABLE IF NOT EXISTS
											  history (type INTEGER, source TEXT, 
													   host TEXT, timestamp DATE, block BOOLEAN)`)
	historyStatement.Exec()
	return &Blocklist{database}
}

// AddHost Add a host name to block.
func (bl *Blocklist) AddHost(host string) {
	statement, _ := bl._db.Prepare(`INSERT OR IGNORE INTO
									blocklist (hostname, ip)
                                    VALUES (?, '0.0.0.0')`)
	statement.Exec(host)
}

// ShouldBlockHost Remap a host to an IP that it returns if it exists
// Returns (ip, err)
func (bl *Blocklist) ShouldBlockHost(host string) bool {
	statement, _ := bl._db.Prepare(`SELECT ip
                                    FROM blocklist
                                    WHERE hostname = ?`)
	row := statement.QueryRow(host)
	var ip string
	err := row.Scan(&ip)
	return err == nil
}

func (bl *Blocklist) GetBlocklists() []string {
	rows, _ := bl._db.Query(`SELECT source FROM sources`)
	defer rows.Close()

	sources := make([]string, 0)
	for rows.Next() {
		var source string
		rows.Scan(&source)
		sources = append(sources, source)
	}
	return sources
}

func (bl *Blocklist) AddBlocklist(source string) {
	statement, _ := bl._db.Prepare(`INSERT OR IGNORE INTO
									sources (source)
									VALUES (?)`)
	statement.Exec(source)
}

func (bl *Blocklist) RecordHistory(history *HistoryEntry) {
	statement, _ := bl._db.Prepare(`INSERT INTO history
									(type, source, host, timestamp, block)
									VALUES (?, ?, ?, ?, ?)`)
	statement.Exec(history.ResourceType, history.Source, history.Host,
		history.Timestamp.Unix(), history.Block)
}

func (bl *Blocklist) GetHistory() []HistoryEntry {

	rows, _ := bl._db.Query(`SELECT type, source, host, timestamp, block
							FROM history LIMIT 100`)
	defer rows.Close()

	entries := make([]HistoryEntry, 0)
	for rows.Next() {
		var entry HistoryEntry
		err := rows.Scan(&entry.ResourceType, &entry.Source, &entry.Host, &entry.Timestamp, &entry.Block)
		if err != nil {
			panic(err)
		}
		entries = append(entries, entry)
	}
	return entries
}
