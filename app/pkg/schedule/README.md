# Distributed Job Scheduling Framework Usage Guide

## 📌 Overview
This framework provides job scheduling capabilities in distributed environments with the following features:  
✅ Multiple schedule types (Immediate/Daily/Interval execution)  
✅ Single-node execution locking mechanism  
✅ Task overlap prevention  
✅ Random execution delay  
✅ Distributed Redis locks  
✅ Execution trace tracking

## 🧩 Core Concepts
```go
type Schedule // Main scheduler managing all jobs 
type Job // Job instance with execution configuration 
type RunType // Job execution type enumeration
```
## 🚀 Quick Start
```go 
// Initialize scheduler 
scheduler := schedule.New(logger, redisClient, traceID)
// Add scheduled job 
scheduler.AddJob("daily_cleanup", cleanupHandler) .DailyAt("23:30:00") .OnOneServer() .WithoutOverlapping()
// Add interval job 
scheduler.AddJob("health_check", checkHandler) .PerSeconds(30) .RandomDelay(10, 30)
// Start scheduler 
scheduler.Start()
```
## ⚙️ Job Configuration Options

### Execution Policies
| Method               | Description                      | Example                     |
|----------------------|----------------------------------|-----------------------------|
| OnOneServer()        | Single-node cluster execution    | job.OnOneServer()          |
| WithoutOverlapping() | Prevent overlapping execution    | job.WithoutOverlapping()   |
| RandomDelay()        | Random execution delay (seconds) | job.RandomDelay(10, 30)    |

### Schedule Types
```go
// Immediate execution 
job.Immediate()
// Daily fixed-time execution (multiple times supported) 
job.DailyAt("09:00:00", "18:00:00")
// Interval execution 
job.PerSeconds(30) // Every 30 seconds 
job.PerMinuit(15) // Every 15 minutes 
job.PerHour(2) // Every 2 hours
```
## 🔧 Advanced Configuration

### Handler Interface Requirements
```go
type HandlerFunc interface { 
	Exec(ctx context.Context) // Main execution logic 
	Error() <-chan error // Error channel 
	Done() <-chan struct{} // Completion signal channel 
}
```
### Redis Lock Configuration
- Default lock TTL: 600 seconds (modify DefaultServerLockTTL)
- Lock key format: `schedule:jobLock:<job_name>:<lock_type>`

## ⚠️ Important Notes
1. **Redis Dependency**: Requires proper Redis client initialization
2. **Error Handling**: Must implement error channel in HandlerFunc
3. **Time Format**: Use "HH:MM:SS" format for DailyAt
4. **Lock Renewal**: Automatic Redis lock renewal for single-node jobs
5. **Random Delay**: Ensure Max > Min to prevent panic

## 🎯 Best Practices
```go
// Typical production configuration 
scheduler.AddJob("order_statistics", statsHandler) 
    .DailyAt("00:05:00") // Execute daily at 00:05 
    .OnOneServer() // Single instance in cluster 
    .WithoutOverlapping() // Prevent overlap 
    .RandomDelay(30, 60) // 30-60s random delay
```
Complete examples can be found in method comments. All configuration methods support chaining calls.