package wechat

import (
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
)

var messageQueue chan *Message

const messageQueueLength = 400

// NewContext start wechat message queue service
func NewContext() {
	if messageQueue != nil {
		return
	}

	messageQueue = make(chan *Message, messageQueueLength)
	go processMessageQueue()
}

// processMessageQueue 处理消息队列中的消息
func processMessageQueue() {
	for msg := range messageQueue {
		log.Trace("New wechat message sending request %s: %20s", msg.openid, msg.content)

		token, err := getCachedToken(setting.Wechat.AppId, setting.Wechat.AppSecret)
		if err == nil {
			err = sendMessage(token, msg.openid, msg.content, msg.url)
		}

		if err != nil {
			log.Error("Failed to send wechat message %s: %.10s... - %v", msg.openid, msg.content, err)
		}
	}
}
