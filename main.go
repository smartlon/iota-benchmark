package main

import (
	"fmt"
	"github.com/smartlon/iota-benchmark/docker/config"
	"github.com/smartlon/iota-benchmark/docker/logging"
	"github.com/smartlon/iota-benchmark/docker/monitor"
	"github.com/smartlon/iota-benchmark/spammer"
	"github.com/smartlon/iota-benchmark/txmonitor"
	"sync"
)

var logger = lancetlogging.GetLogger()

func main() {
	cfy, err := config.LoadFromFile("./docker/config/config.yaml")
	if err != nil {
		logger.Errorf("LoadFromFile Error: %s", err)
	}
	cf := cfy.GetAllConfig()
	monitor.NewMail(cf.Mail.MailUser, cf.Mail.MailPasswd, cf.Mail.SmtpHost, cf.Mail.ReceiveMail)
	mcs := make([]*monitor.MonitorCli, 0)
	for hostname, host := range cf.Hosts {
		mc, err := monitor.NewMonitorCliFromConf(hostname, host.Address, host.ApiVersion, cf.IntervalTime, cf.Tls.TlsSwitch, cf.Tls.ClientCertPath)
		if err != nil {
			logger.Errorf("NewMonitorCliFromConf Error: %s", err)
			panic(err)
		}
		mcs = append(mcs, mc)
	}
	monitorSwitch := monitor.NewMonitorSwitch(mcs)
	monitor.FinishMonitor = make(chan bool)

	monitorSwitch.StartMonitor()
	logger.Debugf("MonitorTime  is  %s", cf.Time)
	nodes := make([]string,0)
	var wg sync.WaitGroup
	for _,mc := range mcs {
		containerInfo,err := mc.GetContainList()
		if err != nil {
			logger.Errorf("GetContainList Error: %s", err)
		}
		for _,cinfo := range containerInfo {
			for _,port := range cinfo.Port {
				if port.PrivatePort == 14265 {
					nodes = append(nodes,fmt.Sprintf("http://202.117.43.212:%d",port.PublicPort))
					break
				}
			}
		}
		wg.Add(2)
		logger.Debugf("nodes  is  %s", nodes)
		go spammer.Spammer(nodes,&wg)
		go txmonitor.Monitor(&wg)
	}
	wg.Wait()
	monitorSwitch.StopMonitor()
}