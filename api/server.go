package api

import (
	"github.com/gordonshieh94/GopherHole/blocklist"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "http://localhost:3000"
		}}
)

type CustomContext struct {
	echo.Context
	DB   *blocklist.Blocklist
	pool *WebSocketPool
}

func getRoot(c echo.Context) error {
	db := c.(*CustomContext)
	bytes, _ := json.Marshal(db.DB.GetBlocklists())
	return c.JSONBlob(http.StatusOK, bytes)
}

func getHistory(c echo.Context) error {
	db := c.(*CustomContext)
	bytes, _ := json.Marshal(db.DB.GetHistory())
	return c.JSONBlob(http.StatusOK, bytes)
}

func postNewBlocklist(c echo.Context) (err error) {
	source := c.FormValue("source")
	db := c.(*CustomContext)
	db.DB.AddBlocklist(source)
	return c.NoContent(http.StatusCreated)
}

func historyStreamer(c echo.Context) (err error) {
	r := c.Request()
	ws, err := upgrader.Upgrade(c.Response(), r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	pool := c.(*CustomContext).pool
	client := toPooledClient(ws, pool)
	go client.SendToClient()
	pool.register <- &client
	return c.NoContent(http.StatusCreated)
}

func StartAPIServer(bl *blocklist.Blocklist, blockStream chan []byte) {
	e := echo.New()
	pool := newWebSocketPool(blockStream)
	go pool.run()
	// Attach context for DB access object
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c, bl, pool}
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
	e.GET("/history-stream", historyStreamer)
	e.Logger.Fatal(e.Start(":1323"))
}
