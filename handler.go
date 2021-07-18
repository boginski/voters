package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) AddPoll(c *gin.Context) {
	var poll Poll

	if err := c.BindJSON(&poll); err != nil {
		fmt.Printf("failed to bind poll: %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	new_id, err := h.storage.Add(&poll)
	if err != nil {
		fmt.Println(new_id, err)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"poll_id": new_id,
	})
}

func (h *Handler) VotePoll(c *gin.Context) {
	poll_id, err1 := strconv.Atoi(c.Param("poll_id"))
	choice_id, err2 := strconv.Atoi(c.Param("choice_id"))
	fmt.Println(poll_id, choice_id)
	if err1 != nil {
		fmt.Printf("failed to convert poll_id param to int: %s\n", err1.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err1.Error(),
		})
		return
	}

	if err2 != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err2.Error(),
		})
		return
	}

	voters_now, err := h.storage.Vote(poll_id, choice_id)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"Count of voters now": voters_now,
	})
}

func (h *Handler) GetResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("poll_id"))
	if err != nil {
		fmt.Printf("failed to convert id param to int: %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	report, err := h.storage.ReportBack(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report.result)
}
