package loadbalancer

import (
	"errors"
	"math/rand"
)

var (
	ErrNotExistService = errors.New("service instances are not exist")
)

// 服务实例结构体
type InstanceInfo struct {
	// 当前权重
	CurWeight int
	// 权重
	Weight int
	// 地址
	Address string
}

// 负载均衡器
type LoadBalancer interface {
	// 选择服务器
	SelectService(services []*InstanceInfo) (*InstanceInfo, error)
	// 根据key选择服务器
	SelectServiceByKey(service []*InstanceInfo, key string) (*InstanceInfo, error)
}

// 随机负载均衡器
type RandomLoadBalancer struct {
}

// 选择服务器
func (r *RandomLoadBalancer) SelectService(services []*InstanceInfo) (*InstanceInfo, error) {
	if len(services) == 0 {
		return nil, ErrNotExistService
	}
	return services[rand.Intn(len(services))], nil
}

// 权重平滑负载均衡器
type WeightRoundRobinLoadBalancer struct {
}

// 选择服务器
// 每次当请求到来，选取服务实例时，该策略会遍历服务实例队列中的所有服务实例。
// 对于每个服务实例，让它的CurWeight 值加上 Weight 值；同时累加所有服务实例的Weight 值，将其保存为Total。
// 遍历完所有服务实例之后，如果某个服务实例的CurWeight最大，就选择这个服务实例处理本次请求，最后把该服务实例的 CurWeight 减去 Total 值
func (w *WeightRoundRobinLoadBalancer) SelectService(services []*InstanceInfo) (*InstanceInfo, error) {
	if len(services) == 0 {
		return nil, ErrNotExistService
	}
	total := 0
	var best *InstanceInfo
	for i := 0; i < len(services); i++ {
		w := services[i]
		if w == nil {
			continue
		}
		// 累加当前权重值
		w.CurWeight += w.Weight
		total += w.Weight
		if best == nil || w.CurWeight > best.CurWeight {
			best = w
		}
	}
	if best == nil {
		return nil, nil
	}
	best.CurWeight -= total
	return best, nil
}

// 一致性哈希负载均衡器
type HashLoadBalancer struct {
}

// 根据key选择服务器
func (h *HashLoadBalancer) SelectServiceByKey(services []*InstanceInfo, key string) (*InstanceInfo, error) {
	if len(services) == 0 {
		return nil, ErrNotExistService
	}
	nodeWeight := make(map[string]int)
	instanceMap := make(map[string]*InstanceInfo)
	for i := 0; i < len(services); i++ {
		instance := services[i]
		nodeWeight[instance.Address] = i
		instanceMap[instance.Address] = instance
	}
	// sort.Sort()
	// 建立哈希环
	hash := NewHashRing()
	// 添加各个服务实例到环上
	hash.AddNodes(nodeWeight)
	// 根据请求的key来获取对缘的服务实例
	host := hash.GetNode(key)
	return instanceMap[host], nil
}
