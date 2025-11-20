package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"minodl/log"
	"minodl/models"
	"os/exec"
	"strings"
)

const (
	BILIBILI = "bilibili.com"
	YouTube  = "youtube.com"
	DouYin   = "douyin.com"
	XHS      = "xiaohongshu.com"
)

func ParseVideoInfo(videoUrl string) (*models.VideoResult, error) {
	cmd := exec.Command("yt-dlp", "--dump-single-json", videoUrl)
	/*if strings.Contains(videoUrl, "youtube.com") {
		cmd = exec.Command("yt-dlp", "--add-header", forwarded, "--cookies", youtubeCookies, "--dump-single-json", videoUrl)
	}*/
	// 捕获标准输出和标准错误
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	// 运行命令
	err := cmd.Run()
	if err != nil {
		log.Error(fmt.Sprint(err) + ": " + stderr.String())
		return nil, err
	}
	var videoInfo models.VideoInfo
	// 解析JSON输出
	err = json.Unmarshal(out.Bytes(), &videoInfo)
	if err != nil {
		log.Error("解析JSON失败:%v", err)
		return nil, err
	}
	// 需要单独保持封面
	if strings.Contains(videoUrl, YouTube) ||
		strings.Contains(videoUrl, BILIBILI) {
		/*working, _ := os.Getwd()
		ImageId := uuid.New().String()
		day := time.Now().Format("2006-01-02")
		tempDir := fmt.Sprintf("%s/web/%s/", working, day)
		os.MkdirAll(tempDir, os.ModePerm)
		tempImage := fmt.Sprintf("%s%s.webp", tempDir, ImageId)
		dlCmd := fmt.Sprintf("wget -O %s %s", tempImage, videoInfo.Thumbnail)
		_, err = util.Cmd(dlCmd, true)
		if err != nil {
			log.Error("download video cover %s error: %s", videoInfo.Thumbnail, err)
		} else {
			h.Image = fmt.Sprintf("https://%s/statics/%s/%s.webp", ictx.Request().Host, day, ImageId)
			videoInfo.Thumbnail = h.Image
			ret.Thumbnail = h.Image
			log.Release("download image to %s, url %s", tempImage, h.Image)
		}*/
	}
	// 输出视频标题
	log.Info("视频标题:%s", videoInfo.Title)
	log.Info("视频URL:%s", videoInfo.URL)
	log.Info("视频时长:%d", videoInfo.Duration)
	log.Info("封面URL:%s", videoInfo.Thumbnail)
	return &models.VideoResult{Title: videoInfo.Title, Thumbnail: videoInfo.Thumbnail, Qualities: make([]map[string]string, 0)}, nil
}
