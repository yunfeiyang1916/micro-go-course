package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// 注册/发现服务实例
type InstanceInfo struct {
	// 服务实例ID，用来唯一标识服务实例
	ID string `json:"id"`
	// 服务发现时返回的服务名
	Service string `json:"service,omitempty"`
	// 服务名,代表服务实例归属的服务集群
	Name string `json:"name"`
	// 标签，可用于进行服务过滤
	Tags []string `json:"tags,omitempty"`
	// 服务实例地址
	Address string `json:"address"`
	// 服务实例端口
	Port int `json:"port"`
	// 元数据
	Meta map[string]string `json:"meta,omitempty"`
	// 是否允许标签覆盖
	EnableTagOverride bool `json:"enable_tag_override"`
	// 健康检查相关配置
	Check `json:"check,omitempty"`
	// 权重
	Weights `json:"weights,omitempty"`
}

// 健康检查相关配置
type Check struct {
	// 多久之后注销服务
	DeregisterCriticalServiceAfter string `json:"deregister_critical_service_after"`
	// 请求参数
	Args []string `json:"args,omitempty"`
	// 健康检查地址
	HTTP string `json:"http"`
	// Consul 主动检查间隔
	Interval string `json:"interval,omitempty"`
	// 服务实例主动维持心跳间隔，与Interval只存其一
	TTL string `json:"ttl,omitempty"`
}

// 权重
type Weights struct {
	Passing int `json:"passing"`
	Warning int `json:"warning"`
}

// 服务发现客户端
type DiscoveryClient struct {
	// consul的host
	host string
	// consul的端口
	port int
}

// 实例化服务发现客户端
func NewDiscoveryClient(host string, port int) *DiscoveryClient {
	return &DiscoveryClient{
		host: host,
		port: port,
	}
}

// 服务注册
func (consulClient *DiscoveryClient) Register(ctx context.Context, serviceName, instanceId, healthCheckUrl string, instanceHost string, instancePort int, meta map[string]string, weights *Weights) error {
	instanceInfo := &InstanceInfo{
		ID:                instanceId,
		Name:              serviceName,
		Address:           instanceHost,
		Port:              instancePort,
		Meta:              meta,
		EnableTagOverride: false,
		Check: Check{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "https//" + instanceHost + ":" + strconv.Itoa(instancePort) + healthCheckUrl,
			Interval:                       "15s",
		},
	}
	if weights != nil {
		instanceInfo.Weights = *weights
	} else {
		instanceInfo.Weights = Weights{
			Passing: 10,
			Warning: 1,
		}
	}
	byteData, err := json.Marshal(instanceInfo)
	if err != nil {
		log.Printf("json format err:%s", err)
		return err
	}
	req, err := http.NewRequest("PUT", "https://"+consulClient.host+":"+strconv.Itoa(consulClient.port)+"/v1/agent/service/register", bytes.NewReader(byteData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	client.Timeout = time.Second * 2
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("register service err : %s", err)
		return err
	}
	defer req.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("register service http request errCode : %v", resp.StatusCode)
		return fmt.Errorf("register service http request errCode : %v", resp.StatusCode)
	}
	log.Println("register service success")
	return nil
}

// 注销服务注册
func (consulClient *DiscoveryClient) Deregister(ctx context.Context, instanceId string) error {
	req, err := http.NewRequest("PUT",
		"http://"+consulClient.host+":"+strconv.Itoa(consulClient.port)+"/v1/agent/service/deregister/"+instanceId, nil)

	if err != nil {
		log.Printf("req format err: %s", err)
		return err
	}

	client := http.Client{}
	client.Timeout = time.Second * 2

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("deregister service err : %s", err)
		return err
	}

	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("deresigister service http request errCode : %v", resp.StatusCode)
		return fmt.Errorf("deresigister service http request errCode : %v", resp.StatusCode)
	}

	log.Println("deregister service success")

	return nil
}

// 服务发现
func (consulClient *DiscoveryClient) DiscoverServices(ctx context.Context, serviceName string) ([]*InstanceInfo, error) {
	req, err := http.NewRequest("GET",
		"http://"+consulClient.host+":"+strconv.Itoa(consulClient.port)+"/v1/health/service/"+serviceName, nil)

	if err != nil {
		log.Printf("req format err: %s", err)
		return nil, err
	}

	client := http.Client{}
	client.Timeout = time.Second * 2

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("discover service err : %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("discover service http request errCode : %v", resp.StatusCode)
		return nil, fmt.Errorf("discover service http request errCode : %v", resp.StatusCode)
	}
	var serviceList []struct {
		Service InstanceInfo `json:"service"`
	}
	err = json.NewDecoder(resp.Body).Decode(&serviceList)
	if err != nil {
		log.Printf("format service info err : %s", err)
		return nil, err
	}
	instances := make([]*InstanceInfo, len(serviceList))
	for i := 0; i < len(instances); i++ {
		instances[i] = &serviceList[i].Service
	}
	return instances, nil
}
