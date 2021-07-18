package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	memoryStorage := NewMemoryStorage()
	handler := NewHandler(memoryStorage)

	router := gin.Default()

	router.POST("/api/createPoll/", handler.AddPoll)
	router.PUT("/api/poll/:poll_id/:choice_id", handler.VotePoll)
	router.GET("/api/getResult/:poll_id", handler.GetResult)

	router.Run(":8000")
}
