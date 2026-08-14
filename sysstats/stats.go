package sysstats

import (
	"runtime"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/cpu"
	"fmt"
)

type NodeSpecs struct {
	RAM uint64 `json:"ram"`
	DiskSpace uint64 `json:"disk_space"`
	CPUCores int `json:"cpu_cores"`
}

func GetSpecs() (NodeSpecs, error) {
	os := runtime.GOOS
	path := ""
	if os == "windows" {
		path = "C:" // TODO: add support for multiple drives on windows
	} else {
		path = "/"
	}

	v, err := mem.VirtualMemory()
	if err != nil {
    return NodeSpecs{}, fmt.Errorf("failed to get memory stats: %w", err)
	}
	d, err := disk.Usage(path)
	if err != nil {
    return NodeSpecs{}, fmt.Errorf("failed to get disk stats: %w", err)
	}
	c, err := cpu.Counts(true)
	if err != nil {
    return NodeSpecs{}, fmt.Errorf("failed to get cpu stats: %w", err)
	}
	

	ram_GB := v.Total / (1024 * 1024 * 1024)
	disk_GB := d.Free / (1024 * 1024 * 1024)

	stats := NodeSpecs{
		RAM: ram_GB,
		DiskSpace: disk_GB,
		CPUCores: c,
	}

	return stats, nil
}