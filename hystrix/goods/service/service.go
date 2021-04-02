package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/yunfeiyang1916/micro-go-course/hystrix/goods/common"
)

// 商品详情视图对象
type GoodsDetailVO struct {
	Id   string
	Name string
	// 评论列表
	Comments common.CommentListVO
}

// 服务接口
type Service interface {
	// 获取商品详情
	GetGoodsDetail(ctx context.Context, id string) (GoodsDetailVO, error)
	// 初始化配置
	InitConfig(ctx context.Context)
}

func NewGoodsServiceImpl() Service {
	return &GoodsDetailServiceImpl{}
}

// 商品详情服务实现
type GoodsDetailServiceImpl struct {
	// 是否开启了降级，降级后不在调用评论服务
	callCommentService int
}

// 获取商品详情
func (g *GoodsDetailServiceImpl) GetGoodsDetail(ctx context.Context, id string) (GoodsDetailVO, error) {
	detail := GoodsDetailVO{Id: id, Name: "商品A"}
	if g.callCommentService != 0 {
		commentResult, err := GetGoodsComments(id)
		if err != nil {
			return detail, err
		}
		detail.Comments = commentResult.Detail
	}
	return detail, nil
}

// 初始化配置
func (g *GoodsDetailServiceImpl) InitConfig(ctx context.Context) {
	log.Printf("InitConfig")
	cli, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	// get
	resp, _ := cli.Get(ctx, "call_service_d")
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
		if string(ev.Key) == "call_service_d" {
			service.callCommentService, _ = strconv.Atoi(string(ev.Value))
		}
	}

	rch := cli.Watch(context.Background(), "call_service_d") // <-chan WatchResponse
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			if string(ev.Kv.Key) == "call_service_d" {
				service.callCommentService, _ = strconv.Atoi(string(ev.Kv.Value))
			}
		}
	}
}

// 获取商品评论集合,使用断路器控制熔断
func GetGoodsComments(id string) (common.CommentResult, error) {
	var result common.CommentResult
	serviceName := "Comments"
	err := hystrix.Do(serviceName, func() error {
		reqUrl := url.URL{
			Scheme:   "http",
			Host:     "127.0.0.1:10087",
			Path:     "/comments/detail",
			RawQuery: "id=" + id,
		}
		resp, err := http.Get(reqUrl.String())
		if err != nil {
			return err
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			return jsonErr
		}
		return nil
	}, func(err error) error {
		return err
	})
	return result, err
}
