package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var err error

func InitDB(dbname string) (*sql.DB, error) {
	_, err = os.Stat(dbname)
	if os.IsNotExist(err) {
		log.Printf("InitDB: database does not exist %s", dbname)
	}
	db, err = sql.Open("sqlite3", dbname)
	if err != nil {
		return nil, fmt.Errorf("InitDB: Error creating database. %s", err.Error())
	}
	err = syncDB()
	if err != nil {
		return nil, fmt.Errorf("InitDB: Error syncing database. %s", err.Error())
	}
	return db, nil
}

func syncDB() error {
	query := `CREATE TABLE IF NOT EXISTS messages (` +
		`id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,` +
		`uuid char(32) UNIQUE NOT NULL,` +
		`message char(160) NOT NULL,` +
		`mobile char(15) NOT NULL,` +
		`status char(15) NOT NULL,` +
		`retries INTEGER DEFAULT 0,` +
		`created_at TIMESTAMP default CURRENT_TIMESTAMP,` +
		`updated_at TIMESTAMP);`
	_, err = db.Exec(query, nil)
	if err != nil {
		return fmt.Errorf("syncDB: %s", err.Error())
	}
	return nil
}

func InsertMessage(sms *SMS) error {
	log.Printf("InsertMessage: %#v", sms)
	stmt, err := db.Prepare("INSERT INTO messages(uuid, message, mobile, status) VALUES(?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("InsertMessage: Failed to prepare transaction. %s", err.Error())
	}
	_, err = stmt.Exec(sms.UUID, sms.Body, sms.Mobile, sms.Status)
	if err != nil {
		return fmt.Errorf("InsertMessage: Failed to execute transaction. %s", err.Error())
	}
	return nil
}

//TODO: locks for driver.Stmt (stmt) and driver.Conn (db)
func UpdateMessageStatus(sms SMS) error {
	log.Printf("Updating msg status %#v", sms)
	stmt, err := db.Prepare("UPDATE messages SET status=?, retries=?, updated_at=DATETIME('now') WHERE uuid=?")
	defer stmt.Close()
	if err != nil {
		return fmt.Errorf("UpdateMessageStatus: %s", err.Error())
	}
	_, err = stmt.Exec(sms.Status, sms.Retries, sms.UUID)
	if err != nil {
		return fmt.Errorf("UpdateMessageStatus: %s", err.Error())
	}
	return nil
}

func GetMessageByUuid(uuid string) (SMS, error) {
	log.Println("GetMessageById:", uuid)
	var sms SMS
	query := fmt.Sprintf("SELECT uuid, message, mobile, status, retries FROM"+
		" messages WHERE uuid == \"%s\"", uuid)

	rows, err := db.Query(query)
	if err != nil {
		return sms, fmt.Errorf("GetMessageByUuid: %s", err.Error())
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries)
	} else {
		return sms, fmt.Errorf("GetMessageByUuid: Failed to get message %s", uuid)
	}
	return sms, nil
}

// GetAllMessages returns all messages with specified status
func GetAllMessages() ([]SMS, error) {
	var messages []SMS
	query := "SELECT uuid, message, mobile, status, retries FROM messages"
	rows, err := db.Query(query)
	if err != nil {
		return messages, fmt.Errorf("GetAllMessages: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		sms := SMS{}
		rows.Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries)
		messages = append(messages, sms)
	}

	return messages, nil
}

func GetPendingMessages() ([]SMS, error) {
	// log.Println("GetPendingMessages")
	var messages []SMS
	query := "SELECT uuid, message, mobile, status, retries FROM" +
		" messages WHERE status = \"pending\" AND retries < 3"

	rows, err := db.Query(query)
	if err != nil {
		return messages, fmt.Errorf("GetPendingMessages: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		sms := SMS{}
		rows.Scan(&sms.UUID, &sms.Body, &sms.Mobile, &sms.Status, &sms.Retries)
		messages = append(messages, sms)
	}
	//rows.Close()
	db.Exec("UPDATE messages SET status='sending' WHERE status = \"pending\" AND retries < 3")
	return messages, nil
}
