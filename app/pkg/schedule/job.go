package schedule

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

const (
	DailyRunType      = "daily"
	PerSecondsRunType = "seconds"
	PerMinuitRunType  = "minuit"
	PerHourRunType    = "hour"

	defaultServerLockTTL = 600 // 默认单服务任务锁时间（10分钟）
)

type (
	Job struct {
		Name                  string          // 任务实例名
		Logger                *logger.Manager // 日志
		Redis                 *redis.Manager  // Redis
		Handler               HandlerFunc     // 执行器
		EnableMultipleServers bool            // 允许多节点执行
		EnableOverlapping     bool            // 允许即使之前的任务实例还在执行，调度内的任务也会执行
		RunTime               *RunTime        // 任务实例运行时参数
		TraceID               *trace.ID
	}

	HandlerFunc interface {
		Exec(ctx context.Context)
		Error() <-chan error
		Done() <-chan struct{}
	}

	RunTime struct {
		Type          string        // 调度时间类型
		Time          interface{}   // 执行时间（时间点、执行间隔时长）
		Locked        bool          // EnableOverlapping 值为 false 时任务执行锁，保证单节点有且只有一个任务在执行
		PerTypeLocked bool          // 间隔固定时长类型任务锁
		Done          chan struct{} // 执行结束
		RandomDelay   *RandomDelay  // 随机延迟执行时间
	}

	RandomDelay struct {
		Min int
		Max int
	}
)

// WithoutOverlapping 避免任务重复
func (j *Job) WithoutOverlapping() *Job {
	j.EnableOverlapping = false
	return j
}

// RandomDelay 设置随机延迟执行时间区间，单位秒。
// 当 min 和 max 值都不为 0 且 max > min 时，任务会在[min,max]秒之后执行
func (j *Job) RandomDelay(min, max int) *Job {
	if max < min {
		panic("must max > min")
	}

	j.RunTime.RandomDelay = &RandomDelay{
		Min: min,
		Max: max,
	}

	return j
}

// DailyAt 每天 time 执行一次任务，可设置多个时间点
// 例子：DailyAt("07:30:00", "12:00:00", "18:00:00")
// 与其他执行时间互斥，每个任务有且只有一个执行时间
func (j *Job) DailyAt(time ...string) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = DailyRunType
		j.RunTime.Time = time
	}
	return j
}

// PerSeconds 每 seconds 秒执行一次任务
// 与其他执行时间互斥，每个任务有且只有一个执行时间
func (j *Job) PerSeconds(seconds int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = PerSecondsRunType
		j.RunTime.Time = time.Duration(seconds) * time.Second
	}
	return j
}

// PerMinuit 每 minuit 分钟执行一次任务
// 与其他执行时间互斥，每个任务有且只有一个执行时间
func (j *Job) PerMinuit(minuit int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = PerMinuitRunType
		j.RunTime.Time = time.Duration(minuit) * time.Minute
	}
	return j
}

// PerHour 每 hour 小时执行一次任务
// 与其他执行时间互斥，每个任务有且只有一个执行时间
func (j *Job) PerHour(hour int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = PerHourRunType
		j.RunTime.Time = time.Duration(hour) * time.Hour
	}
	return j
}

// OnOneServer 任务只运行在一台服务器上
// 需要 Redis 服务支持
func (j *Job) OnOneServer() *Job {
	j.EnableMultipleServers = false
	return j
}

// runWithRecover 为任务运行添加了recover，以防止panic导致的程序崩溃
func (j *Job) runWithRecover() {
	ctx := context.WithValue(context.Background(), logger.TraceIDKey, j.TraceID.New())

	defer func() {
		// 如果有panic，记录错误并继续
		if r := recover(); r != nil {
			j.Logger.Error(ctx, "job has a panic error", zap.Any("error", r))
		}
	}()

	j.handler(ctx)
}

// run 方法根据不同的运行类型执行任务。
// 对于每日运行类型，它将在指定的时间点执行任务。
// 对于每秒、每分钟、每小时运行类型，它将基于定时器定期执行任务。
func (j *Job) run() {
	switch j.RunTime.Type {
	case DailyRunType:
		// 遍历每日运行的时间点，若当前时间匹配，则启动任务执行
		times := j.RunTime.Time.([]string)
		for _, t := range times {
			if time.Now().Format("15:04:05") == t {
				go j.runWithRecover()
			}
		}
	case PerSecondsRunType, PerMinuitRunType, PerHourRunType:
		// 对于定期运行，确保仅执行一次，避免重复执行
		if j.RunTime.PerTypeLocked {
			return
		}

		j.RunTime.PerTypeLocked = true
		go func() {
			ticker := time.NewTicker(j.RunTime.Time.(time.Duration))
			for range ticker.C {
				go j.runWithRecover()
			}
		}()
	}
}

// handler 方法负责处理任务的执行逻辑。
// 它首先会检查是否允许任务重叠执行，然后根据是否启用多个服务器执行任务的模式进行加锁处理。
// 最后，它会异步执行任务并处理可能的错误。
func (j *Job) handler(ctx context.Context) {
	if !j.EnableOverlapping {
		// 任务重叠执行不允许时的加锁逻辑
		if j.RunTime.Locked {
			return
		}
		j.RunTime.Locked = true
	}

	if !j.EnableMultipleServers {
		// 单服务器执行逻辑，包括加锁和锁续期
		if !j.lock("Server", defaultServerLockTTL, false) {
			j.RunTime.Locked = false
			return
		}

		go j.renewalServerLock(ctx)
	}

	// 随机休眠
	j.randomDelay()

	j.Logger.Info(ctx, util.SpliceStr("The scheduled job: ", j.Name, " starts execution."))

	// 错误处理和任务完成后的清理逻辑
	go func(ctx context.Context) {
	Exit:
		for {
			select {
			case err := <-j.Handler.Error():
				if err != nil {
					j.Logger.Error(ctx, fmt.Sprintf("An error occurred while executing the %s.", j.Name), zap.Error(err))
				}
			case <-j.Handler.Done():
				// 任务完成后的清理逻辑
				if !j.EnableMultipleServers {
					j.RunTime.Done <- struct{}{}
				}

				j.RunTime.Locked = false

				j.Logger.Info(ctx, util.SpliceStr("The scheduled job: ", j.Name, " has done."))

				break Exit
			}
		}
	}(ctx)

	j.Handler.Exec(ctx)
}

// randomDelay 在设定的时间区间内随机生成一个时间段（秒），休眠
func (j *Job) randomDelay() {
	if j.RunTime.RandomDelay == nil {
		return
	}

	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	delay := generator.Intn(j.RunTime.RandomDelay.Max) + j.RunTime.RandomDelay.Min
	time.Sleep(time.Duration(delay) * time.Second)
}

// lock 方法用于加锁，如果指定的 renewal 为 true，则为续期操作。
// 返回值表示加锁（或续期）操作是否成功。
func (j *Job) lock(name string, ttl int, renewal bool) bool {
	prefix := j.Redis.Prefix
	key := util.SpliceStr(prefix, "schedule:jobLock:", j.Name, ":", name)

	// 根据是否续期执行不同的 Redis 操作
	if renewal {
		_, err := j.Redis.Do("EXPIRE", key, ttl)
		if err == nil {
			return true
		}
	} else {
		ok, err := j.Redis.Do("SET", key, "locked", "EX", ttl, "NX")
		if ok != nil && err == nil {
			return true
		}
	}

	return false
}

// unLock 方法用于释放指定名称的锁。
func (j *Job) unLock(ctx context.Context, name string) {
	key := util.SpliceStr("schedule:jobLock:", j.Name, ":", name)

	ok, err := j.Redis.Del(key)
	if !ok && err != nil {
		j.Logger.Error(ctx, util.SpliceStr("unLock job:", name, "failed"), zap.Error(err))
	}
}

// renewalServerLock 方法用于定期续期服务器锁，确保锁不会过期。
func (j *Job) renewalServerLock(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
Exit:
	for {
		select {
		case <-ticker.C:
			// 定期执行锁续期操作
			j.lock("Server", defaultServerLockTTL, true)
		case <-j.RunTime.Done:
			// 任务完成或取消时释放服务器锁
			j.unLock(ctx, "Server")
			break Exit
		}
	}
}
