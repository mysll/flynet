package master

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	//"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func GetInfo() map[string]interface{} {
	path := "/"
	if runtime.GOOS == "windows" {
		file, _ := exec.LookPath(os.Args[0])
		path = filepath.VolumeName(file)
	}

	diskinfo := disk.UsageStat{}
	if v, err := disk.Usage(path); err == nil {
		var i interface{}
		i = v
		switch inst := i.(type) {
		case disk.UsageStat:
			diskinfo = inst
			break
		case *disk.UsageStat:
			if inst != nil {
				diskinfo = *inst
			}
			break
		}

	}
	diskinfo.Total /= 1024 * 1024
	diskinfo.Used /= 1024 * 1024

	vm := mem.VirtualMemoryStat{}
	sm := mem.SwapMemoryStat{}

	if v, err := mem.VirtualMemory(); err == nil {
		vm = *v
	}
	vm.Total /= 1024 * 1024
	vm.Used /= 1024 * 1024
	if v, err := mem.SwapMemory(); err == nil {
		sm = *v
	}
	sm.Total /= 1024 * 1024
	sm.Used /= 1024 * 1024

	var cpupercent float64

	ps, err := cpu.Percent(time.Millisecond, true)
	if err == nil && len(ps) > 0 {
		for _, v := range ps {
			cpupercent += v
		}
		cpupercent /= float64(len(ps))
	}

	return map[string]interface{}{
		"DiskInfo":      diskinfo,
		"SwapMemory":    sm,
		"VirtualMemory": vm,
		"CpuPercent":    cpupercent * 100,
	}
}
