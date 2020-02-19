package blocklist

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Blocklist object for managing the DNS blocklist in a sqlite database
type Blocklist struct {
	_db *sql.DB
}

// GetDatabase Get the database access object
// Initializes the db if not exists in the same cwd of the exectuable
func GetDatabase() *Blocklist {
	database, _ := sql.Open("sqlite3", "data.db")
	statement, _ := database.Prepare(`CREATE TABLE IF NOT EXISTS
                                      blocklist (hostname TEXT PRIMARY KEY, ip TEXT)`)
	statement.Exec()
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
