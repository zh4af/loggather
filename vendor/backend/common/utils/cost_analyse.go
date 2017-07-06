package utils

import (
	"backend/common/clog"
	"fmt"
	"time"
)

type CostAnalyse struct {
	Name string
	At   time.Time
}

func InitCostAnalyse(node string) []CostAnalyse {
	return []CostAnalyse{CostAnalyse{
		Name: node,
		At:   time.Now(),
	}}
}

func AddCostNode(node string, costList []CostAnalyse) []CostAnalyse {
	costNode := CostAnalyse{
		Name: node,
		At:   time.Now(),
	}
	return append(costList, costNode)
}

func PrintCostAnalyse(costList []CostAnalyse) {
	if len(costList) < 2 {
		return
	}

	var str string
	str = fmt.Sprintf("[total:%d ms] ", costList[len(costList)-1].At.Sub(costList[0].At)/1000000)
	for i, costNode := range costList {
		if i == len(costList)-1 {
			break
		}
		nextNode := costList[i+1]
		costTime := nextNode.At.Sub(costNode.At).Nanoseconds() / 1000000
		str += fmt.Sprintf("[%s~%s:%d ms]", costNode.Name, nextNode.Name, costTime)
	}
	clog.Noticef("%s", str)
}
