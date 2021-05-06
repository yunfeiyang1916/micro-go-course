package loadbalancer

import (
	"crypto/sha1"
	"math"
	"sort"
	"strconv"
	"sync"
)

// 一致性哈希算法
const (
	// 默认虚拟结点数
	DefaultVirualSpots = 400
)

// 结点
type node struct {
	// 节点key
	nodeKey string
	// 虚拟结点哈希值
	spotValue uint32
}

// 结点集合
type nodesArray []node

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

// 哈希环
type HashRing struct {
	// 虚拟结点数
	virualSpots int
	// 结点集合
	nodes nodesArray
	// 权重配置，以服务地址为键，权重为值的map
	weights map[string]int
	mu      sync.RWMutex
}

// 构建哈希环
func NewHashRing() *HashRing {
	spots := DefaultVirualSpots
	h := &HashRing{
		virualSpots: spots,
		weights:     make(map[string]int),
	}
	return h
}

// 批量添加服务结点
func (h *HashRing) AddNodes(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for nodeKey, w := range nodeWeight {
		h.weights[nodeKey] = w
	}
	h.generate()
}

// 添加服务结点
func (h *HashRing) AddNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.weights[nodeKey] = weight
	h.generate()
}

// 移除服务结点
func (h *HashRing) RemoveNode(nodeKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.weights, nodeKey)
	h.generate()
}

// 更新服务结点
func (h *HashRing) UpdateNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.weights[nodeKey] = weight
	h.generate()
}

// 生成哈希环
func (h *HashRing) generate() {
	var totalW int
	for _, w := range h.weights {
		totalW += w
	}
	totalVirtualSpots := h.virualSpots * len(h.weights)
	h.nodes = nodesArray{}
	for nodeKey, w := range h.weights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			hash := sha1.New()
			hash.Write([]byte(nodeKey + ":" + strconv.Itoa(i)))
			hashBytes := hash.Sum(nil)
			n := node{
				nodeKey:   nodeKey,
				spotValue: genValue(hashBytes[6:10]),
			}
			h.nodes = append(h.nodes, n)
			hash.Reset()
		}
	}
	h.nodes.Sort()
}
func genValue(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	return v
}

// 获取指定值所在环中的服务节点
func (h *HashRing) GetNode(s string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.nodes) == 0 {
		return ""
	}
	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	v := genValue(hashBytes[6:10])
	i := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].spotValue >= v
	})
	if i == len(h.nodes) {
		i = 0
	}
	return h.nodes[i].nodeKey
}
