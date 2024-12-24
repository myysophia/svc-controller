package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WecomMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

func sendWecomAlert(webhook string, message string) error {
	msg := WecomMessage{
		MsgType: "markdown",
	}

	// 构建markdown格式消息
	msg.Markdown.Content = fmt.Sprintf(`
### US vnnox-rabbitmq service切换通知
> 时间: %s

%s
`, time.Now().Format("2006-01-02 15:04:05"), message)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal wecom message failed: %v", err)
	}

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("send wecom message failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom api returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
