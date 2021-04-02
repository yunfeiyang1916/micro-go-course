package loadbalancer

// 负载均衡器
type LoadBalancer interface {
	// 选择服务器
	SelectService(services []string) (string, error)
	SelectServiceByKey(service []string, key string) (string, error)
}

// 随机负载
type RandomLoadBalancer struct {
}

// 选择服务器
func (r *RandomLoadBalancer) SelectService(services []string) (string, error) {
	if len(services) == 0 {
		return
	}
}
