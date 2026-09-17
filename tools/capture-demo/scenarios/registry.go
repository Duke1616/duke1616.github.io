package scenarios

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Scenario)
	order      []string
)

// Register 将一个场景注册到全局注册表（通常在各个 scenario 文件的 init() 中调用）
func Register(s Scenario) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[s.Name]; !exists {
		order = append(order, s.Name)
	}
	registry[s.Name] = s
}

// Get 根据名称获取场景定义
func Get(name string) (Scenario, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	s, ok := registry[name]
	return s, ok
}

// All 按注册顺序返回所有场景
func All() []Scenario {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return lo.Map(order, func(name string, _ int) Scenario {
		return registry[name]
	})
}

// ListNames 返回所有已注册的场景名称列表
func ListNames() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return append([]string(nil), order...)
}

// MustGet 获取指定名称的场景，若不存在则 panic
func MustGet(name string) Scenario {
	s, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("场景 %q 未注册，可用场景: %v", name, ListNames()))
	}
	return s
}
