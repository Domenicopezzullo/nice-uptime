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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, password TEXT)")
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

		rows := db.QueryRow("SELECT * FROM users where username = '?'", username)
		fmt.Printf("SELECT * FROM users where username = '%s'\n", username)
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

	s.Run(":8080")
}
