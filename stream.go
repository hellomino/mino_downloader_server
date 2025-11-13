package main

import (
	"bufio"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------- 数据结构 ----------

const (
	TaskRunning  TaskStatus = "running"
	TaskFinished TaskStatus = "finished"
	TaskFailed   TaskStatus = "failed"
)

type TaskInfo struct {
	TaskID   string     `json:"task_id"`
	URL      string     `json:"url"`
	Progress string     `json:"progress"`
	Status   TaskStatus `json:"status"`
	Started  time.Time  `json:"started"`
	Updated  time.Time  `json:"updated"`
}

var (
	progressRe = regexp.MustCompile(`(?i)\[download\]\s+([\d\.]+)%`)
	taskMu     sync.RWMutex
	tasks      = make(map[string]*TaskInfo) // task_id -> TaskInfo
)

// ---------- 主函数 ----------

func main1() {
	r := gin.Default()

	// 视频流: /stream?url=xxx&task_id=xxx
	r.GET("/stream", handleStream)

	// 获取所有任务状态
	r.GET("/api/tasks", func(c *gin.Context) {
		taskMu.RLock()
		defer taskMu.RUnlock()

		list := []*TaskInfo{}
		for _, t := range tasks {
			list = append(list, t)
		}
		c.JSON(200, list)
	})

	r.Run(":8080")
}

func handleStream(c *gin.Context) {
	videoURL := c.Query("url")
	taskID := c.Query("task_id")
	if videoURL == "" || taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url or task_id"})
		return
	}

	// 初始化任务
	taskMu.Lock()
	tasks[taskID] = &TaskInfo{
		TaskID:   taskID,
		URL:      videoURL,
		Progress: "0",
		Status:   TaskRunning,
		Started:  time.Now(),
		Updated:  time.Now(),
	}
	taskMu.Unlock()

	cmd := exec.Command(
		"yt-dlp",
		"-f", "best[ext=mp4]/best",
		"-o", "-",
		videoURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		updateTask(taskID, TaskFailed, "0")
		c.String(http.StatusInternalServerError, "stdout pipe error: %v", err)
		return
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		updateTask(taskID, TaskFailed, "0")
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
				updateTask(taskID, TaskRunning, progress)
				broadcastProgress(taskID, progress)
			}
		}
	}()

	err = cmd.Wait()
	if err == nil {
		updateTask(taskID, TaskFinished, "100")
		broadcastProgress(taskID, "100")
	} else {
		updateTask(taskID, TaskFailed, "0")
	}
}

func broadcastProgress(taskID, progress string) {
	progressOfTask := gin.H{
		"task_id":  taskID,
		"progress": progress,
	}
	_ = progressOfTask
}

func updateTask(taskID string, status TaskStatus, progress string) {
	taskMu.Lock()
	defer taskMu.Unlock()

	if t, ok := tasks[taskID]; ok {
		t.Progress = progress
		t.Status = status
		t.Updated = time.Now()
	} else {
		tasks[taskID] = &TaskInfo{
			TaskID:   taskID,
			URL:      "",
			Progress: progress,
			Status:   status,
			Started:  time.Now(),
			Updated:  time.Now(),
		}
	}
}
