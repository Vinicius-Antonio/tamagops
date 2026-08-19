package collector

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type TopProcess struct {
	PID     int32   `json:"pid"`
	Name    string  `json:"name"`
	CPUPerc float64 `json:"cpu_percent"`
}

type SystemSnapshot struct {
	CPUPercent float64 `json:"cpu_percent"`

	CPUTemp          float64 `json:"cpu_temp_celsius"`
	BoardTemp        float64 `json:"board_temp_celsius"`
	SensorsAvailable bool    `json:"sensors_available"`

	RAMTotal       uint64  `json:"ram_total_bytes"`
	RAMUsed        uint64  `json:"ram_used_bytes"`
	RAMAvailable   uint64  `json:"ram_available_bytes"`
	RAMUsedPercent float64 `json:"ram_used_percent"`

	SwapTotal       uint64  `json:"swap_total_bytes"`
	SwapUsed        uint64  `json:"swap_used_bytes"`
	SwapUsedPercent float64 `json:"swap_used_percent"`

	DiskTotal       uint64  `json:"disk_total_bytes"`
	DiskUsed        uint64  `json:"disk_used_bytes"`
	DiskFree        uint64  `json:"disk_free_bytes"`
	DiskUsedPercent float64 `json:"disk_used_percent"`

	Load1         float64 `json:"load_avg_1min"`
	Load5         float64 `json:"load_avg_5min"`
	Load15        float64 `json:"load_avg_15min"`
	LoadAvailable bool    `json:"load_available"`

	UptimeSeconds uint64 `json:"uptime_seconds"`

	TopProcess TopProcess `json:"top_process"`

	CollectedAt time.Time `json:"collected_at"`
}

func diskRoot() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

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

func collectRAM() (*mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("ram: %w", err)
	}
	return v, nil
}

func collectSwap() (*mem.SwapMemoryStat, error) {
	s, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("swap: %w", err)
	}
	return s, nil
}

func collectDisk() (*disk.UsageStat, error) {
	d, err := disk.Usage(diskRoot())
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	return d, nil
}

func collectUptime() (uint64, error) {
	info, err := host.Info()
	if err != nil {
		return 0, fmt.Errorf("uptime: %w", err)
	}
	return info.Uptime, nil
}

func collectLoad() (avg1, avg5, avg15 float64, available bool) {
	stats, err := load.Avg()
	if err != nil {
		return 0, 0, 0, false
	}
	return stats.Load1, stats.Load5, stats.Load15, true
}

func isCPUSensor(key string) bool {
	lowered := strings.ToLower(key)
	return strings.Contains(lowered, "cpu") ||
		strings.Contains(lowered, "core") ||
		strings.Contains(lowered, "package") ||
		strings.Contains(lowered, "tctl") ||
		strings.Contains(lowered, "tdie")
}

func collectTemperatures() (cpuTemp, boardTemp float64, available bool) {
	readings, err := host.SensorsTemperatures()
	if err != nil && len(readings) == 0 {
		return 0, 0, false
	}

	var cpuMax, boardMax float64
	var sawAny bool

	for _, reading := range readings {
		if reading.Temperature <= 0 {
			continue
		}
		sawAny = true

		if isCPUSensor(reading.SensorKey) {
			if reading.Temperature > cpuMax {
				cpuMax = reading.Temperature
			}
		} else {
			if reading.Temperature > boardMax {
				boardMax = reading.Temperature
			}
		}
	}

	if !sawAny {
		return 0, 0, false
	}

	return cpuMax, boardMax, true
}

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

	swap, swapErr := collectSwap()
	cpuTemp, boardTemp, sensorsOK := collectTemperatures()
	load1, load5, load15, loadOK := collectLoad()

	snap := SystemSnapshot{
		CPUPercent: cpuPerc,

		CPUTemp:          cpuTemp,
		BoardTemp:        boardTemp,
		SensorsAvailable: sensorsOK,

		RAMTotal:       ram.Total,
		RAMUsed:        ram.Used,
		RAMAvailable:   ram.Available,
		RAMUsedPercent: ram.UsedPercent,

		DiskTotal:       dsk.Total,
		DiskUsed:        dsk.Used,
		DiskFree:        dsk.Free,
		DiskUsedPercent: dsk.UsedPercent,

		Load1:         load1,
		Load5:         load5,
		Load15:        load15,
		LoadAvailable: loadOK,

		UptimeSeconds: uptime,
		TopProcess:    top,
		CollectedAt:   time.Now(),
	}

	if swapErr == nil && swap != nil {
		snap.SwapTotal = swap.Total
		snap.SwapUsed = swap.Used
		snap.SwapUsedPercent = swap.UsedPercent
	}

	return snap, nil
}
