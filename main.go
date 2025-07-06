package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Domenicopezzullo/skibidi-uptime/checker"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Uptime struct {
	URL    string
	Status string
}

type User struct {
	ID       int
	Username string
	Password string
}

func main() {
	db, err := sql.Open("sqlite", "skibidi.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		panic(err.Error())
	}
	checker.StartMonitoringFromDB(db, 30*time.Second)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		password TEXT
	)`)
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Uptimes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT,
		status TEXT DEFAULT 'UNKNOWN',
		userId INTEGER,
		FOREIGN KEY(userId) REFERENCES Users(id)
	)`)
	if err != nil {
		panic(err.Error())
	}

	s := gin.Default()
	s.LoadHTMLGlob("templates/html/*")

	s.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", gin.H{})
	})

	s.POST("/", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")

		var user User
		err := db.QueryRow("SELECT id, username, password FROM Users WHERE username = ?", username).
			Scan(&user.ID, &user.Username, &user.Password)

		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"authorized": false,
				"message":    "Wrong username or password",
			})
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"authorized": false,
				"message":    "A problem occurred",
			})
			return
		}

		if password != user.Password {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"authorized": false,
				"message":    "Wrong username or password",
			})
			return
		}

		rows, err := db.Query("SELECT url, status FROM Uptimes WHERE userId = (SELECT id FROM Users WHERE username = ?)", username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load uptimes"})
			return
		}
		defer rows.Close()

		var uptimes []Uptime
		for rows.Next() {
			var u Uptime
			if err := rows.Scan(&u.URL, &u.Status); err == nil {
				uptimes = append(uptimes, u)
			}
		}

		ctx.HTML(http.StatusOK, "dashboard.html", gin.H{
			"username": username,
			"uptimes":  uptimes,
		})
	})

	// API endpoint
	s.POST("/api/addUptime", func(ctx *gin.Context) {
		url := ctx.PostForm("url")
		user := ctx.PostForm("user")

		var userID int
		err := db.QueryRow("SELECT id FROM Users WHERE username = ?", user).Scan(&userID)
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "User not found"})
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
			return
		}

		_, err = db.Exec("INSERT INTO Uptimes (url, status, userId) VALUES (?, ?, ?)", url, "UNKNOWN", userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to insert uptime"})
			return
		}
		ctx.JSON(200, gin.H{
			"abc": "test",
		})
	})

	s.Run(":8080")
}
