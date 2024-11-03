package duckron

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
)

type Diagnostics struct {
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
}

type alertOptions struct {
	ramThreshold  float64
	cpuThreshold  float64
	diskThreshold float64
}

type alertManager struct {
	options   *alertOptions
	Timer     *Timer
	alertChan chan int
	errChan   chan *Error
}

func NewAlertManager(options *alertOptions, alertChan chan int, errChan chan *Error) *alertManager {
	if options == nil {
		options = &alertOptions{
			ramThreshold:  0.8,
			cpuThreshold:  0.8,
			diskThreshold: 0.8,
		}
	}

	if options.ramThreshold == 0 {
		options.ramThreshold = 0.8
	}
	if options.cpuThreshold == 0 {
		options.cpuThreshold = 0.8
	}
	if options.diskThreshold == 0 {
		options.diskThreshold = 0.8
	}

	timer := NewTimer(5 * time.Second)

	return &alertManager{options: options, alertChan: alertChan, errChan: errChan, Timer: timer}
}

func (am *alertManager) monitor() {
	go func(errChan chan *Error) {
		if err := am.Timer.Start(
			func() *Error {
				diagnostics := diagnose()

				if diagnostics.CPUUsage > am.options.cpuThreshold {
					am.alertChan <- CPU_THRESHOLD_EXCEEDED
				}

				if diagnostics.MemoryUsage > am.options.ramThreshold {
					am.alertChan <- RAM_THRESHOLD_EXCEEDED
				}

				if diagnostics.DiskUsage > am.options.diskThreshold {
					am.alertChan <- DISK_THRESHOLD_EXCEEDED
				}
				return nil
			},
		); err != nil {
			errChan <- err
		}
	}(am.errChan)
}

func diagnose() Diagnostics {
	cpuUsage, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		fmt.Println("Error retrieving CPU usage:", err)
		cpuUsage = []float64{0.0}
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	totalMemory := float64(memStats.Sys)
	memoryUsage := float64(memStats.Alloc) / totalMemory

	diskUsageStat, err := disk.Usage("/")
	if err != nil {
		fmt.Println("Error retrieving disk usage:", err)
		diskUsageStat = &disk.UsageStat{UsedPercent: 0}
	}
	diskUsage := diskUsageStat.UsedPercent / 100.0

	diagnostics := Diagnostics{
		CPUUsage:    cpuUsage[0] / 100.0,
		MemoryUsage: memoryUsage,
		DiskUsage:   diskUsage,
	}

	// Output for debugging
	fmt.Println("CPU:", diagnostics.CPUUsage, "Memory:", diagnostics.MemoryUsage, "Disk:", diagnostics.DiskUsage)

	return diagnostics
}
