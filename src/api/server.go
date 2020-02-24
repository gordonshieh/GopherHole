package api

import (
	"blocklist"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
)

type DBContext struct {
	echo.Context
	DB *blocklist.Blocklist
}

func getRoot(c echo.Context) error {
	db := c.(*DBContext)
	bytes, _ := json.Marshal(db.DB.GetBlocklists())
	return c.JSONBlob(http.StatusOK, bytes)
}

func postNewBlocklist(c echo.Context) (err error) {
	source := c.FormValue("source")
	db := c.(*DBContext)
	db.DB.AddBlocklist(source)
	return c.NoContent(http.StatusCreated)
}

func StartAPIServer(bl *blocklist.Blocklist) {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &DBContext{c, bl}
			return next(cc)
		}
	})

	e.GET("/", getRoot)
	e.POST("/add", postNewBlocklist)
	e.Logger.Fatal(e.Start(":1323"))
}
