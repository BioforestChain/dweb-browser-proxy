package v1

import "github.com/gogf/gf/v2/frame/g"

// 1.2  App
type CreateTopicReq struct {
	g.Meta    `path:"/pubsub/publish_msg" tags:"CreateTopicReqService" method:"post" summary:"Create A New Chat Topic"`
	TopicName string `v:"required"` //主题名称
	Msg       string `v:""`         //消息内容
}

// ChatCreateTopicRes
// @Description:
type CreateTopicRes struct {
}
type SubscribeMsgReq struct {
	g.Meta    `path:"/pubsub/subscribe_msg" tags:"SubscribeMsgReqService" method:"post" summary:"Subscribe by TopicName"`
	TopicName string `v:"required"` //主题名称
}

type SubscribeMsgRes struct {
	Msg string `json:"msg"` //消息内容
}
