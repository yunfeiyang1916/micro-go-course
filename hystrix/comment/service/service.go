package service

import "context"

// 评论列表视图对象
type CommentListVO struct {
	Id string
	// 评论列表
	CommentList []CommentVo
}

// 评论视图对象
type CommentVo struct {
	Id string
	// 描述
	Desc string
	// 分数
	Score float32
	// 回复id
	ReplyId string
}

// 评论服务接口
type Service interface {
	// 获取评论列表
	GetCommentsList(ctx context.Context, id string) (CommentListVO, error)
}

func NewGoodsServiceImpl() Service {
	return &CommentsServiceImpl{}
}

// 评论服务实现
type CommentsServiceImpl struct{}

// 获取评论列表
func (service *CommentsServiceImpl) GetCommentsList(ctx context.Context, id string) (CommentListVO, error) {
	comment1 := CommentVo{Id: "1", Desc: "comments", Score: 1.0, ReplyId: "0"}
	comment2 := CommentVo{Id: "2", Desc: "comments", Score: 1.0, ReplyId: "1"}

	list := []CommentVo{comment1, comment2}
	detail := CommentListVO{Id: id, CommentList: list}
	return detail, nil
}
