package collector

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// TopProcess represents the process consuming the most CPU at the moment.
type TopProcess struct {
	PID     int32   `json:"pid"`
	Name    string  `json:"name"`
	CPUPerc float64 `json:"cpu_percent"`
}

// SystemSnapshot holds all hardware metrics collected in a single cycle.
type SystemSnapshot struct {
	// CPU usage percentage (0-100).
	CPUPercent float64 `json:"cpu_percent"`

	// RAM metrics.
	RAMTotal       uint64  `json:"ram_total_bytes"`
	RAMUsed        uint64  `json:"ram_used_bytes"`
	RAMAvailable   uint64  `json:"ram_available_bytes"`
	RAMUsedPercent float64 `json:"ram_used_percent"`

	// Disk metrics (root partition).
	DiskTotal       uint64  `json:"disk_total_bytes"`
	DiskUsed        uint64  `json:"disk_used_bytes"`
	DiskFree        uint64  `json:"disk_free_bytes"`
	DiskUsedPercent float64 `json:"disk_used_percent"`

	// Host uptime in seconds.
	UptimeSeconds uint64 `json:"uptime_seconds"`

	// The process with the highest CPU usage right now.
	TopProcess TopProcess `json:"top_process"`

	// Timestamp of this snapshot.
	CollectedAt time.Time `json:"collected_at"`
}

// diskRoot returns the root path based on the operating system.
func diskRoot() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

// collectCPU reads the average CPU usage over a 1-second interval.
func collectCPU() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("cpu: %w", err)
	}
	if len(percentages) == 0 {
		return 0, nil
	}
	return percentages[0], nil
}

// collectRAM reads virtual memory statistics.
func collectRAM() (*mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("ram: %w", err)
	}
	return v, nil
}

// collectDisk reads disk usage for the root partition.
func collectDisk() (*disk.UsageStat, error) {
	d, err := disk.Usage(diskRoot())
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	return d, nil
}

// collectUptime reads the host uptime in seconds.
func collectUptime() (uint64, error) {
	info, err := host.Info()
	if err != nil {
		return 0, fmt.Errorf("uptime: %w", err)
	}
	return info.Uptime, nil
}

// collectTopProcess scans all running processes and returns the one
// with the highest CPU consumption at this instant.
func collectTopProcess() (TopProcess, error) {
	procs, err := process.Processes()
	if err != nil {
		return TopProcess{}, fmt.Errorf("processes: %w", err)
	}

	type scored struct {
		pid  int32
		name string
		cpu  float64
	}

	candidates := make([]scored, 0, len(procs))
	for _, p := range procs {
		cpuPerc, err := p.CPUPercent()
		if err != nil {
			continue
		}
		name, _ := p.Name()
		candidates = append(candidates, scored{
			pid:  p.Pid,
			name: name,
			cpu:  cpuPerc,
		})
	}

	if len(candidates) == 0 {
		return TopProcess{}, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].cpu > candidates[j].cpu
	})

	top := candidates[0]
	return TopProcess{
		PID:     top.pid,
		Name:    top.name,
		CPUPerc: top.cpu,
	}, nil
}

// Collect gathers all system metrics in a single cycle and returns
// a complete SystemSnapshot ready to be consumed by the engine.
func Collect() (SystemSnapshot, error) {
	cpuPerc, err := collectCPU()
	if err != nil {
		return SystemSnapshot{}, err
	}

	ram, err := collectRAM()
	if err != nil {
		return SystemSnapshot{}, err
	}

	dsk, err := collectDisk()
	if err != nil {
		return SystemSnapshot{}, err
	}

	uptime, err := collectUptime()
	if err != nil {
		return SystemSnapshot{}, err
	}

	top, err := collectTopProcess()
	if err != nil {
		return SystemSnapshot{}, err
	}

	return SystemSnapshot{
		CPUPercent:      cpuPerc,
		RAMTotal:        ram.Total,
		RAMUsed:         ram.Used,
		RAMAvailable:    ram.Available,
		RAMUsedPercent:  ram.UsedPercent,
		DiskTotal:       dsk.Total,
		DiskUsed:        dsk.Used,
		DiskFree:        dsk.Free,
		DiskUsedPercent: dsk.UsedPercent,
		UptimeSeconds:   uptime,
		TopProcess:      top,
		CollectedAt:     time.Now(),
	}, nil
}
