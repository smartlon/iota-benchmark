package monitor

import (
	"fmt"
	"strings"
	"time"
)

type MonitorSwitch struct {
	MonitorCliList []*MonitorCli
}

var monitorSwitch *MonitorSwitch

func NewMonitorSwitch(mc []*MonitorCli) *MonitorSwitch {
	if monitorSwitch == nil {
		monitorSwitch = &MonitorSwitch{MonitorCliList: mc}
		return monitorSwitch
	}
	return monitorSwitch
}
func (ms *MonitorSwitch) StartMonitor() {
	mcs := ms.MonitorCliList
	for _, mc := range mcs {
		go startOneMonitor(mc)
	}
}

func (ms *MonitorSwitch) StopMonitor() {
	//mcs := ms.MonitorCliList
	FinishMonitor <- true
	close(FinishMonitor)
	//cl, err := getRecordDataList()
	//if err != nil {
	//	logger.Errorf("getRecordDataList Error: %s", err)
	//	return
	//}
	//var intervalTime float64
	//if len(ms.MonitorCliList) > 0 {
	//	intervalTime = ms.MonitorCliList[0].intervalTime.Seconds()
	//}
	for cname,cstats := range cstatsMap {
		var cpuMax,cpuAvg,memMax,memAvg float64
		clen := len(cstats)
		netIN := cstats[len(cstats)-1].NetIN
		netOUT := cstats[len(cstats)-1].NetOUT
		cpuMax = cstats[0].Cpu
		memMax = cstats[0].Memory
		for _,cstat := range cstats {
			if  cpuMax <cstat.Cpu {
				cpuMax = cstat.Cpu
			}
			cpuAvg += cstat.Cpu
			if  memMax <cstat.Memory {
				memMax = cstat.Memory
			}
			memAvg += cstat.Memory
		}
		cpuAvg = cpuAvg/float64(clen)
		memAvg = memAvg/float64(clen)
		fmt.Println("cname=%s, cpuMax=%6.2f, cpuAvg=%6.2f,memMax=%6.2f,memAvg=%6.2f, netIN=%6.2f, netOUT=%6.2f",cname,cpuMax,cpuAvg,memMax,memAvg,netIN,netOUT)
	}
	//HandleData(cl, intervalTime)
	//logger.Debugf("Make the chart completed! please watch in 'Lancet/resultData/ChartFile' Contents !")
	//for _, mc := range mcs {
	//	FinishChart.Add(1)
	//	cl, _ := mc.GetRecordDataList()
	//	go HandleData(cl)
	//}
	//FinishChart.Wait()
}

/*
每次向一个服务器的所有容器获取一次容器状态，每次间隔intervalTime
*/
func startOneMonitor(monCli *MonitorCli) {
	cl, _ := monCli.GetContainList()
	i := 0
	for {
		//每隔1分钟获取一次容器List，防止中途有容器挂掉，还在监控
		if i*(int)(monCli.intervalTime/time.Second)%60 == 0 {
			cl_new, _ := monCli.GetContainList()
			if result, diff := diffContainerlist(cl, cl_new); diff {
				msg := fmt.Sprintf("Contain is shuntDown! please checkout it!\n %v", result)
				if Mail != nil {
					//go Mail.sendMail(msg)
					logger.Debugf(msg)
				}
				cl = cl_new
			}
		}
		for _, c := range cl {
			logger.Debugf("start MonitorContain[%s]!", c.ContainerName)
			go monCli.MonitorContain(monCli.Hostname, c.ContainerName)
		}
		select {
		case <-FinishMonitor:
			logger.Debugf("Finish Monitor Work !")
			return
		default:
			logger.Debugf("Monitor Work RuningTime is %d !", i*(int)(monCli.intervalTime))
		}
		i++
		time.Sleep(monCli.intervalTime)
	}
}

func diffContainerlist(containerList1 []*ContainerInfo, containerList2 []*ContainerInfo) ([]string, bool) {

	if len(containerList1) == len(containerList2) {
		return nil, false
	}
	var allData string
	var result []string

	if len(containerList1) < len(containerList2) {
		for _, contain := range containerList1 {
			allData += contain.ContainerName
		}
		for _, contain := range containerList2 {
			if !strings.Contains(allData, contain.ContainerName) {
				result = append(result, contain.HostName+"-"+contain.ContainerName)
			}
		}

	} else {
		for _, contain := range containerList2 {
			allData += contain.ContainerName
		}
		for _, contain := range containerList1 {
			if !strings.Contains(allData, contain.ContainerName) {
				result = append(result, contain.HostName+"-"+contain.ContainerName)
			}
		}
	}

	return result, true
}
