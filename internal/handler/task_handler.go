package handler

import (
	"encoding/json"
	"net/http"

	"task/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

type createTaskRequest struct {
	Title    string          `json:"title" binding:"required"`
	Payload  json.RawMessage `json:"payload"`
	Priority int             `json:"priority"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uint64)
	idStr, err := h.tasks.Create(c.Request.Context(), userID, req.Title, req.Payload, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"task_id": idStr, "status": "queued"})
}

func (h *TaskHandler) Get(c *gin.Context) {
	idStr := c.Param("task_id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
		return
	}
	id := uid[:]
	t, err := h.tasks.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	var payload interface{}
	var result interface{}
	if len(t.Payload) > 0 {
		_ = json.Unmarshal(t.Payload, &payload)
	}
	if len(t.Result) > 0 {
		_ = json.Unmarshal(t.Result, &result)
	}
	c.JSON(http.StatusOK, gin.H{
		"task_id":    t.UUIDString(),
		"status":     t.Status,
		"result":     result,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	})
}
