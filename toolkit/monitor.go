package toolkit

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// MonitorInfo
type MonitorInfo struct {
	RequestNum int64
	QPSValue   float64
	IPAddr     string
	reqnum     int64
	uptime     int64
	lock       sync.Mutex
}

// MonitorInfo.RequestHook
func (m *MonitorInfo) RequestHook() {
	m.lock.Lock()
	m.RequestNum++
	m.lock.Unlock()
}

// MonitorInfo.UpdateInfo
func (m *MonitorInfo) UpdateInfo() {
	m.lock.Lock()
	nowtime := time.Now().Unix()
	m.QPSValue = float64(m.RequestNum-m.reqnum) / float64(nowtime-m.uptime)
	m.reqnum = m.RequestNum
	m.uptime = nowtime
	m.lock.Unlock()
}

var Monitor *MonitorInfo

// MonitorInit
func MonitorInit() {
	Monitor = &MonitorInfo{
		RequestNum: 0,
		QPSValue:   0,
		IPAddr:     "",
		reqnum:     0,
		uptime:     0,
	}

	ipaddrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("InterfaceAddrs: %s\n", err.Error())
		return
	}
	for _, ipaddr := range ipaddrs {
		if ipnet, ok := ipaddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				Monitor.IPAddr = ipnet.IP.String()
				break
			}
		}
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			Monitor.UpdateInfo()
		}
	}()
}

// MonitorHookHTTP
func MonitorHookHTTP(w http.ResponseWriter, r *http.Request) {
	Monitor.RequestHook()
}

// MonitorHook
func MonitorHook() {
	Monitor.RequestHook()
}
