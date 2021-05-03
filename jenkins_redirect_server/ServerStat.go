package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	//"github.com/shirou/gopsutil/net"
	"net/http"
	//"runtime"
	"strconv"
	//"syscall"
)

// Можно заюзать этот пакет
// https://github.com/ricochet2200/go-disk-usage/blob/master/du/diskusage.go

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

/*type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// disk usage of path/disk
func DiskUsage(path string) (DiskStatus, error) {
	var disk DiskStatus = DiskStatus{0, 0, 0}

	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return disk, err
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return disk, nil
}

	// Example
	disk, err := DiskUsage("/")
	if err == nil {
		fmt.Printf("All: %.2f GB\n", float64(disk.All)/float64(GB))
		fmt.Printf("Used: %.2f GB\n", float64(disk.Used)/float64(GB))
		fmt.Printf("Free: %.2f GB\n", float64(disk.Free)/float64(GB))
	} else {

	}
*/

func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
		//os.Exit(-1)
	}
}

func GetHardwareData(w http.ResponseWriter, r *http.Request) {
	//runtimeOS := runtime.GOOS

	// memory
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)

	// disk - start from "/" mount point for Linux
	// might have to change for Windows!!
	// don't have a Window to test this out, if detect OS == windows
	// then use "\" instead of "/"

	diskStat, err := disk.Usage("/")
	dealwithErr(err)

	// cpu - get CPU number of cores and speed
	cpuStat, err := cpu.Info()
	dealwithErr(err)
	percentage, err := cpu.Percent(0, true)
	dealwithErr(err)

	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	dealwithErr(err)

	// get interfaces MAC/hardware address
	//interfStat, err := net.Interfaces()
	//dealwithErr(err)

	html := "<html>"

	html = html + "<p><a href=\"/\">Main page</a></p>"

	//html = html + "OS : " + runtimeOS + "<br>"
	//html = html + "<br>"

	html = html + fmt.Sprintf("Total RAM memory: %.2f GB <br>", float64(vmStat.Total)/float64(GB))
	html = html + fmt.Sprintf("Free RAM memory: %.2f GB <br>", float64(vmStat.Free)/float64(GB))
	html = html + fmt.Sprintf("Percentage used RAM memory: %.2f%% <br>", vmStat.UsedPercent)
	html = html + "<br>"

	// get disk serial number.... strange... not available from disk package at compile time
	// undefined: disk.GetDiskSerialNumber
	//serial := disk.GetDiskSerialNumber("/dev/sda")

	//html = html + "Disk serial number: " + serial + "<br>"
	html = html + fmt.Sprintf("Total disk space: %.2f GB <br>", float64(diskStat.Total)/float64(GB))
	html = html + fmt.Sprintf("Used disk space: %.2f GB <br>", float64(diskStat.Used)/float64(GB))
	html = html + fmt.Sprintf("Free disk space: %.2f GB <br>", float64(diskStat.Free)/float64(GB))
	html = html + fmt.Sprintf("Percentage used disk memory: %.2f%% <br>", diskStat.UsedPercent)
	html = html + "<br>"

	// since my machine has one CPU, I'll use the 0 index
	// if your machine has more than 1 CPU, use the correct index
	// to get the proper data
	html = html + "CPU index number: " + strconv.FormatInt(int64(cpuStat[0].CPU), 10) + "<br>"
	html = html + "VendorID: " + cpuStat[0].VendorID + "<br>"
	html = html + "Family: " + cpuStat[0].Family + "<br>"
	html = html + "Number of cores: " + strconv.FormatInt(int64(cpuStat[0].Cores), 10) + "<br>"
	html = html + "Model Name: " + cpuStat[0].ModelName + "<br>"
	html = html + "Speed: " + strconv.FormatFloat(cpuStat[0].Mhz, 'f', 2, 64) + " MHz <br>"
	html = html + "<br>"

	for idx, cpupercent := range percentage {
		html = html + "Current CPU utilization: [" + strconv.Itoa(idx) + "] " + strconv.FormatFloat(cpupercent, 'f', 2, 64) + "%<br>"
	}
	html = html + "<br>"

	html = html + "Hostname: " + hostStat.Hostname + "<br>"
	html = html + "Uptime: " + strconv.FormatUint(hostStat.Uptime, 10) + "<br>"
	html = html + "Number of processes running: " + strconv.FormatUint(hostStat.Procs, 10) + "<br>"

	// another way to get the operating system name
	// both darwin for Mac OSX, For Linux, can be ubuntu as platform
	// and linux for OS

	html = html + "OS: " + hostStat.OS + "<br>"
	html = html + "Platform: " + hostStat.Platform + "<br>"

	// the unique hardware id for this machine
	html = html + "Host ID(uuid): " + hostStat.HostID + "<br>"
	html = html + "<br>"

	/*
		html = html + "<br>"

		for _, interf := range interfStat {
			html = html + "------------------------------------------------------<br>"
			html = html + "Interface Name: " + interf.Name + "<br>"

			if interf.HardwareAddr != "" {
				html = html + "Hardware(MAC) Address: " + interf.HardwareAddr + "<br>"
			}

			for _, flag := range interf.Flags {
				html = html + "Interface behavior or flags: " + flag + "<br>"
			}

			for _, addr := range interf.Addrs {
				html = html + "IPv6 or IPv4 addresses: " + addr.String() + "<br>"

			}
		}
	*/

	html = html + "</html>"

	w.Write([]byte(html))
}
