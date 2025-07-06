package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "skibidi.db")
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, password TEXT)")
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Uptimes (id INTEGER PRIMARY KEY AUTOINCREMENT, url TEXT, time TEXT, userId INTEGER, FOREIGN KEY(userId) REFERENCES Users(id))")
	if err != nil {
		panic(err.Error())
	}
	s := gin.Default()
	s.LoadHTMLGlob("templates/*")
	s.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "login.html", gin.H{})
	})
	s.POST("/", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		_ = ctx.PostForm("password")

		fmt.Println(username)

		rows := db.QueryRow("SELECT * FROM Users where username = '?'", username)
		fmt.Printf("SELECT * FROM Users where username = '%s'\n", username)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println(rows)

		err := rows.Err()
		if err != nil && err == sql.ErrNoRows {
			ctx.JSON(404, gin.H{
				"authorized": false,
				"message":    "I cannot let you in, stranger",
			})
			return
		} else if err != nil {
			ctx.JSON(500, gin.H{
				"authorized": false,
				"message":    "A problem occurred",
			})
			return
		}

		ctx.HTML(200, "dashboard.html", gin.H{
			"username": username,
		})

	})

	// api endpoints

	s.POST("/api/addUptime", func(ctx *gin.Context) {
		url := ctx.PostForm("url")
		time := ctx.PostForm("time")
		user := ctx.PostForm("user")

		fmt.Printf("time: %v\n", time)

		id := db.QueryRow("SELECT id FROM Users where username = '?'", user)

		fmt.Printf("id: %v\n", id)

		if id.Err() != nil {
			panic(id.Err().Error())
		}

		db.Exec("INSERT INTO Uptimes VALUES (?, ?)", &url, time)
	})

	s.Run(":8080")
}
