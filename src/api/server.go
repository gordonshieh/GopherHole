package api

import (
	"blocklist"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
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

func getHistory(c echo.Context) error {
	db := c.(*DBContext)
	bytes, _ := json.Marshal(db.DB.GetHistory())
	return c.JSONBlob(http.StatusOK, bytes)
}

func postNewBlocklist(c echo.Context) (err error) {
	source := c.FormValue("source")
	db := c.(*DBContext)
	db.DB.AddBlocklist(source)
	return c.NoContent(http.StatusCreated)
}

func historyStreamer(blockStream chan blocklist.HistoryEntry) func(echo.Context) error {

	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				stream := <-blockStream
				bytes, _ := json.Marshal(stream)
				err := websocket.Message.Send(ws, bytes)
				if err != nil {
					c.Logger().Error(err)
					break
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	}

}

func StartAPIServer(bl *blocklist.Blocklist, blockStream chan blocklist.HistoryEntry) {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &DBContext{c, bl}
			return next(cc)
		}
	})

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", getRoot)
	e.POST("/add", postNewBlocklist)
	e.GET("/history", getHistory)
	e.GET("/history-stream", historyStreamer(blockStream))
	e.Logger.Fatal(e.Start(":1323"))
}
