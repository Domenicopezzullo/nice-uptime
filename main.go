package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := sql.Open("sqlite3", "skibidi.db")
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
		_, err = db.Exec("SELECT * from users")
		if err != nil {
			panic(err.Error())
		}
		ctx.HTML(200, "login.html", gin.H{})
	})
}
