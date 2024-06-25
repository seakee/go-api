package trace

import (
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sk-pkg/util"
)

const (
	initIndex = 10000000 // 初始序列号
	indexBase = 36       // 序列号的基数
)

var (
	hostnameOnce sync.Once // 仅执行一次获取主机名的操作
	hostname     string    // 缓存的主机名
)

// ID 是用于生成唯一标识符的结构体。
type ID struct {
	index  uint64     // 序列号，通过原子操作访问
	prefix string     // 包含时间戳和主机名的前缀
	mu     sync.Mutex // 互斥锁，确保更新前缀时的线程安全
}

// NewTraceID 创建并返回一个新的 ID 实例，使用主机名和时间戳。
// 它在初始化时一次性地检索主机名并缓存它。
func NewTraceID() *ID {
	t := &ID{
		index: initIndex,
	}
	t.updatePrefix()
	return t
}

// updatePrefix 用当前时间戳和缓存的主机名组合前缀。
// 调用此方法时需要外部同步。
func (t *ID) updatePrefix() {
	var err error

	t.mu.Lock()
	defer t.mu.Unlock()

	hostnameOnce.Do(func() {
		hostname, err = os.Hostname()
		if err != nil {
			log.Printf("获取主机名失败: %v", err)
			// 如果获取主机名失败，使用默认值
			hostname = "unknown"
		}
	})

	t.prefix = util.SpliceStr(hostname, "-", strconv.FormatInt(time.Now().UnixNano(), indexBase), "-")
	t.index = initIndex
}

// New 生成并返回一个新的唯一标识符 ID。
func (t *ID) New() string {
	// 原子递增序列号
	newIndex := atomic.AddUint64(&t.index, 1)

	// 如果序列号溢出，加锁后再次检查以更新前缀并重置序列号
	if newIndex == 0 {
		t.mu.Lock()
		defer t.mu.Unlock()
		if atomic.LoadUint64(&t.index) == 0 {
			t.updatePrefix()
		}
	}

	// 将序列号转换为基数为 36 的字符串
	id := strconv.FormatUint(newIndex, indexBase)

	return util.SpliceStr(t.prefix, id)
}
