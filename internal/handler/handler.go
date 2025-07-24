package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alexputin/downloader/config"
	"github.com/alexputin/downloader/internal/task"
	"github.com/alexputin/downloader/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ApiHandler struct {
	taskService task.TaskService
	cfg         *config.Config
}

func NewApiHandler(taskService task.TaskService, config *config.Config) *ApiHandler {
	return &ApiHandler{
		taskService: taskService,
		cfg:         config,
	}
}

func (h *ApiHandler) RegisterRoutes(app *echo.Echo) {
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	app.GET("/archive/:id", h.downloadHandler)

	api := app.Group("api")
	api.POST("/task", h.createTask)
	api.POST("/task/:id/add_file", h.addFile)
	api.GET("/task/:id", h.getTask)

}

func (h *ApiHandler) createTask(ctx echo.Context) error {
	task, err := h.taskService.CreateTask()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, task)
}

func (h *ApiHandler) getTask(ctx echo.Context) error {
	taskId := ctx.Param("id")
	task, err := h.taskService.GetTask(taskId)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, task)
}

func (h *ApiHandler) addFile(ctx echo.Context) error {
	taskId := ctx.Param("id")
	var req AddFileReq
	err := ctx.Bind(&req)
	if err != nil {
		return utils.ErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid body format: %w", err))
	}

	if len(req.URL) == 0 {
		return utils.ErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("url is empty"))
	}

	task, err := h.taskService.AddFile(taskId, req.URL)

	if err != nil {
		return utils.ErrorResponse(ctx, http.StatusBadRequest, err)
	}

	return ctx.JSON(http.StatusOK, task)
}

func (h *ApiHandler) downloadHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, fmt.Errorf("filename is required"))
	}

	task, err := h.taskService.GetTask(id)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err)
	}

	filePath := task.StoredFile

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return utils.ErrorResponse(c, http.StatusNotFound, fmt.Errorf("file not found"))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to open file"))
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to get file info"))
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, fileInfo.Name()))
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", fileInfo.Size()))

	return c.Stream(http.StatusOK, "application/octet-stream", file)
}
