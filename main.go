package main

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	DeploymentType  = "Deployment"
	StatefulSetType = "StatefulSet"
	DaemonSetType   = "DaemonSet"
	JobSetType      = "Job"
)

type Workload struct {
	Name      string
	Type      string
	Namespace string
}

func main() {
	var wg sync.WaitGroup
	workloadsChan := make(chan Workload, 50)
	clientset := newClient()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	meta(clientset, &wg, ctx, workloadsChan)
	wg.Wait()
	close(workloadsChan)

	workloads := make(map[string][]Workload)
	for workload := range workloadsChan {
		workloads[workload.Namespace] = append(workloads[workload.Namespace], workload)
	}

	for namespace, workloadsList := range workloads {
		slices.SortFunc(workloadsList, func(a, b Workload) int {
			return strings.Compare(a.Name, b.Name)
		})
		workloads[namespace] = workloadsList
	}

	buildGraph(workloads)
}

func insertNewline(s string) *string {
	if len(s) == 0 {
		emptyStr := ""
		return &emptyStr
	}
	chars := strings.Split(s, "")
	result := strings.Join(chars, "\n")
	return &result
}

func stringFormatter(s string) string {
	var b strings.Builder
	if !strings.Contains(s, "-") {
		runes := []rune(s)
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteRune(runes[1])
	} else {
		parts := strings.Split(s, "-")
		if len(parts) >= 2 {
			for index, part := range parts[:2] {
				runes := []rune(part)
				if index == 0 {
					b.WriteRune(unicode.ToUpper(runes[0]))
				} else {
					if index == 1 {
						b.WriteRune(unicode.ToLower(runes[0]))
					}
				}
			}
		}
	}
	return b.String()

}
