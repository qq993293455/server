package system_info

import (
	"os"
	"runtime"
	"time"

	"coin-server/common/proto/broadcast"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func MemoryUsedPercent() float64 {
	vms, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	if vms.Total <= 0 {
		return 0
	}
	return float64(vms.Used) / float64(vms.Total)
}

func CPUUsedPercent() float64 {
	perS, err := cpu.Percent(time.Second, false)
	if err != nil || len(perS) == 0 {
		return 0
	}
	return perS[0] / 100
}

func DiskUsedInfo() map[string]float64 {
	m := make(map[string]float64)
	parS, err := disk.Partitions(true)
	if err != nil {
		return m
	}
	for _, v := range parS {
		u, err := disk.Usage(v.Mountpoint)
		if err != nil {
			continue
		}
		if u.Total <= 0 {
			m[v.Mountpoint] = 0
		} else {
			m[v.Mountpoint] = float64(u.Used) / float64(u.Total)
		}
	}
	return m
}

type NetworkIOData struct {
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
	Errin       uint64
	Errout      uint64
}

func NetworkIOInfo() *broadcast.Stats_NetworkIOInfo {
	cs, err := net.IOCounters(false)
	if err != nil || len(cs) == 0 {
		return nil
	}
	c := cs[0]
	sni := &broadcast.Stats_NetworkIOInfo{
		BytesSent:   c.BytesSent,
		BytesRecv:   c.BytesRecv,
		PacketsSent: c.PacketsSent,
		PacketsRecv: c.PacketsRecv,
		ErrIn:       c.Errin,
		ErrOut:      c.Errout,
	}

	return sni
}

func DiskIOInfo() map[string]*broadcast.Stats_DiskIoInfo {
	parS, err := disk.Partitions(true)
	if err != nil {
		return nil
	}
	mps := make([]string, 0, len(parS))
	for _, v := range parS {
		mps = append(mps, v.Mountpoint)
	}
	ds, err := disk.IOCounters(mps...)
	if err != nil {
		return nil
	}
	out := make(map[string]*broadcast.Stats_DiskIoInfo, len(parS))
	for k, v := range ds {
		out[k] = &broadcast.Stats_DiskIoInfo{
			ReadCount:      v.ReadCount,
			WriteCount:     v.WriteCount,
			ReadBytes:      v.ReadBytes,
			WriteBytes:     v.WriteBytes,
			IopsInProgress: v.IopsInProgress,
		}
	}

	return out
}

func NumGoroutine() uint64 {
	return uint64(runtime.NumGoroutine())
}

func SelfInfo() *broadcast.Stats_ProgressInfo {
	pid := os.Getpid()
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil
	}
	ssi := &broadcast.Stats_ProgressInfo{}
	ssi.Pid = int64(pid)
	ssi.CpuPercent, _ = p.CPUPercent()
	ssi.CpuPercent /= 100
	mp, _ := p.MemoryPercent()
	ssi.MemPercent = float64(mp)
	ssi.Name, _ = p.Name()
	s, _ := p.Status()
	ssi.Status = s[0]
	ssi.StartTime, _ = p.CreateTime()
	nt, _ := p.NumThreads()
	ssi.NumThreads = int64(nt)
	nf, _ := p.NumFDs()
	ssi.NumFds = int64(nf)
	ssi.NumGoroutines = int64(NumGoroutine())
	ssi.Exe, _ = p.Exe()
	return ssi
}

func StatsInfo() *broadcast.Stats_StatsInfo {
	si := &broadcast.Stats_StatsInfo{}
	si.CpuPercent = CPUUsedPercent()
	si.MemPercent = MemoryUsedPercent()
	si.DiskPercent = DiskUsedInfo()
	si.NetworkIoInfo = NetworkIOInfo()
	si.DiskIoInfos = DiskIOInfo()
	si.Pi = SelfInfo()
	return si
}
