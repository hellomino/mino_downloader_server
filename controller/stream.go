package controller

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"io"
	"minodl/models"
	"net/http"
	"os/exec"
	"regexp"
	"sync"
	"time"
)

const ZeroProcess = "0"

var (
	progressRe = regexp.MustCompile(`(?i)\[download\]\s+([\d\.]+)%`)
	tasks      = make(map[string]*models.Task) // task_id -> TaskInfo
	tasksMutex sync.RWMutex
)

func HandleStream(c *gin.Context) {
	videoURL := c.Query("url")
	taskID := c.Query("task_id")
	if videoURL == "" || taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url or task_id"})
		return
	}
	tasksMutex.Lock()
	tasks[taskID] = &models.Task{
		VideoURL: videoURL,
		Progress: ZeroProcess,
		Status:   models.StatusPending,
	}
	tasksMutex.Unlock()

	cmd := exec.Command(
		"yt-dlp",
		"-f", "best[ext=mp4]/best",
		"-o", "-",
		videoURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		updateTask(taskID, models.StatusFailed, ZeroProcess)
		c.String(http.StatusInternalServerError, "stdout pipe error: %v", err)
		return
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		updateTask(taskID, models.StatusFailed, ZeroProcess)
		c.String(http.StatusInternalServerError, "start yt-dlp failed: %v", err)
		return
	}

	// 设置视频流头
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", `inline; filename="video.mp4"`)
	c.Header("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	// 边下边推流
	go func() {
		_, _ = io.Copy(c.Writer, stdout)
	}()

	// 解析 stderr，实时进度
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if match := progressRe.FindStringSubmatch(line); len(match) == 2 {
				progress := match[1]
				updateTask(taskID, models.StatusRunning, progress)
				broadcastProgress(taskID, progress)
			}
		}
	}()

	err = cmd.Wait()
	if err == nil {
		updateTask(taskID, models.StatusCompleted, "100")
		broadcastProgress(taskID, "100")
	} else {
		updateTask(taskID, models.StatusFailed, "0")
	}
}

func broadcastProgress(taskID, progress string) {
	progressOfTask := gin.H{
		"task_id":  taskID,
		"progress": progress,
	}
	_ = progressOfTask
}

func updateTask(taskID string, status models.TaskStatus, progress string) {
	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	if t, ok := tasks[taskID]; ok {
		t.Progress = progress
		t.Status = status
		t.UpdatedAt = time.Now()
	}
}
