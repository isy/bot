package main

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/callback", func(c *gin.Context) {
		proxyURL, _ := url.Parse(os.Getenv("FIXIE_URL"))
		client := &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		}

		bot, err := linebot.NewClient(os.Getenv("CHANNEL_ID"), os.Getenv("CHANNEL_SECRET"), os.Getenv("MID") linebot.WithHTTPClient(client))
		if err != nil {
			fmt.Println(err)
			return
		}

		received, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				fmt.Println(err)
			}
			return
		}

		for _, result := range received.Results {
			content := result.Context()
			if content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText {
				text, err := content.TextContent()
				res, err := bot.SendText([]string{content.Form}, "OK "+text.Text)
				if err != nil {
					fmt.Println(res)
				}
			}
		}
	})

	router.Run(":" + port)
}
