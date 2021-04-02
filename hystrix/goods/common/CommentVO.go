package common

// 评论列表视图对象
type CommentListVO struct {
	Id          string      `json:"Id"`
	CommentList []CommentVo `json:"CommentList"`
}

type CommentVo struct {
	Id      string  `json:"Id"`
	Desc    string  `json:"Desc"`
	Score   float32 `json:"Score"`
	ReplyId string  `json:"ReplyId"`
}

// 评论结果
type CommentResult struct {
	Detail CommentListVO `json:"detail"`
}
