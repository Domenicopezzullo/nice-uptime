package checker

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"
)

type UptimeEntry struct {
	URL    string
	UserID int
}

var (
	client = &http.Client{Timeout: 10 * time.Second}
	mutex  = sync.Mutex{}
)

func StartMonitoringFromDB(db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			entries := loadUptimeEntries(db)
			for _, entry := range entries {
				go checkURL(db, entry)
			}
		}
	}()
}

func loadUptimeEntries(db *sql.DB) []UptimeEntry {
	rows, err := db.Query("SELECT url, userId FROM Uptimes")
	if err != nil {
		log.Println("Failed to query uptimes:", err)
		return nil
	}
	defer rows.Close()

	var entries []UptimeEntry
	for rows.Next() {
		var e UptimeEntry
		if err := rows.Scan(&e.URL, &e.UserID); err == nil {
			entries = append(entries, e)
		}
	}
	return entries
}

func checkURL(db *sql.DB, entry UptimeEntry) {
	start := time.Now()
	resp, err := client.Get(entry.URL)
	duration := time.Since(start)

	status := "DOWN"
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
		status = "UP"
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	log.Printf("[Check] %s | Status: %s | Time: %v", entry.URL, status, duration)

	mutex.Lock()
	defer mutex.Unlock()
	_, err = db.Exec("UPDATE Uptimes SET status = ? WHERE url = ? AND userId = ?", status, entry.URL, entry.UserID)
	if err != nil {
		log.Println("Failed to update status:", err)
	}
}
