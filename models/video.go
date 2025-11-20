package models

// VideoFormat 定义视频格式的结构体
type VideoFormat struct {
	ID     string `json:"format_id"`
	Ext    string `json:"video_ext"`
	Format string `json:"format"`
	Height int    `json:"height"`
	URL    string `json:"url"`
	Size   int    `json:"filesize"`
}

// VideoInfo 定义视频信息的结构体
type VideoInfo struct {
	Title     string        `json:"title"`
	Formats   []VideoFormat `json:"formats"`
	URL       string        `json:"webpage_url"`
	PlayUrl   string        `json:"play_url"`
	Duration  float64       `json:"duration"`
	Thumbnail string        `json:"thumbnail"` // 新增封面URL字段
}

type VideoResult struct {
	Title     string              `json:"title"`
	Duration  int                 `json:"duration"`
	Thumbnail string              `json:"thumbnail"`
	Qualities []map[string]string `json:"qualities"`
	Ad        string              `json:"ad"`
}
