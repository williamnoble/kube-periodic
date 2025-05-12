package main

import (
	"context"
	"fmt"
	"k8s.io/utils/ptr"
	"log"
	"maps"
	"os"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	tlog "oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"slices"
	"strconv"
	"strings"
)

var theme = map[string]string{
	"Deployment":  "#ACDDDE",
	"StatefulSet": "#E1F8DC",
	"DaemonSet":   "#FEF8DD",
	"Job":         "#F7D8BA",
}

func buildGraph(workloads map[string][]Workload) {
	empty := ""
	graph := newGraph()

	types := slices.Collect(maps.Keys(theme))
	// <!-- Classes: Base -->
	// we need a `base` theme to inherit for `block` based classes e.g. [base; block], we then implement a hack.
	// this isn't ideal, for themes we'll _add_ the base class to each workload-specific theme.
	types = append(types, "base")
	for _, t := range types {
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("classes.%s.style.fill", t), nil, ptr.To("#fff"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("classes.%s.style.font-size", t), nil, ptr.To("40"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("classes.%s.style.font", t), nil, ptr.To("Mono"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("classes.%s.style.border-radius", t), nil, ptr.To("6"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("classes.%s.style.stroke-width", t), nil, ptr.To("0"))
	}

	//<!-- Classes: Workload Specific -->
	graph, _ = d2oracle.Set(graph, nil, "classes.Deployment.style.fill", nil, ptr.To(theme["Deployment"]))
	graph, _ = d2oracle.Set(graph, nil, "classes.StatefulSet.style.fill", nil, ptr.To(theme["StatefulSet"]))
	graph, _ = d2oracle.Set(graph, nil, "classes.DaemonSet.style.fill", nil, ptr.To(theme["DaemonSet"]))
	graph, _ = d2oracle.Set(graph, nil, "classes.Job.style.fill", nil, ptr.To(theme["Job"]))

	//<!-- Parent Container (Workloads) -->
	graph, _, _ = d2oracle.Create(graph, nil, "table")
	graph, _ = d2oracle.Set(graph, nil, "table.label", nil, ptr.To(""))
	graph, _ = d2oracle.Set(graph, nil, "table.grid-columns", nil, ptr.To(strconv.Itoa(len(workloads))))
	graph, _ = d2oracle.Set(graph, nil, "table.grid-gap", nil, ptr.To("0"))
	graph, _ = d2oracle.Set(graph, nil, "table.style.stroke-width", nil, ptr.To("1"))
	graph, _ = d2oracle.Set(graph, nil, "table.style.stroke-dash", nil, ptr.To("5"))
	graph, _ = d2oracle.Set(graph, nil, "table.style.stroke-width", nil, ptr.To("0"))
	graph, _ = d2oracle.Set(graph, nil, "table.style.fill", nil, ptr.To("white"))

	//<!-- Namespace Container -->
	for namespace := range workloads {
		namespaceContainerKey := fmt.Sprintf("table.%s", namespace)
		graph, _, _ = d2oracle.Create(graph, nil, namespaceContainerKey)
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.label", namespaceContainerKey), nil, &empty)
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.grid-columns", namespaceContainerKey), nil, ptr.To("1"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.grid-gap", namespaceContainerKey), nil, ptr.To("5"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.fill", namespaceContainerKey), nil, ptr.To("white"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.stroke-width", namespaceContainerKey), nil, ptr.To("0"))

		//<!-- Workload Elements -->
		for _, workload := range workloads[namespace] {
			workloadKey := fmt.Sprintf("%s.%s", namespaceContainerKey, stringFormatter(workload.Name))
			graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.class", workloadKey), nil, ptr.To(workload.Type))
		}

		//<!-- Namespace Footer -->
		// use `block` class, so the namespace spans multiple grid rows.
		namespaceFooterKey := fmt.Sprintf("table.%s.%s", namespace, namespace)
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.label", namespaceFooterKey), nil, insertNewline(namespace))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.font-size", namespaceFooterKey), nil, ptr.To("40"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.font-color", namespaceFooterKey), nil, ptr.To("black"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.stroke-width", namespaceFooterKey), nil, ptr.To("1"))
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.stroke-dash", namespaceFooterKey), nil, ptr.To("5"))
		if namespace == "kube-system" {
			graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.fill", namespaceFooterKey), nil, ptr.To("linear-gradient(#ffe5e5, #ff9999)"))
		} else {
			graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.fill", namespaceFooterKey), nil, ptr.To("linear-gradient(#ebf3fa, #cfe2f3)"))
		}
		graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.class", namespaceFooterKey), nil, ptr.To("<[base; block]>"))
	}

	//<!-- Parent Container (Spacer) -->
	// adds a spacer to create additional visual space between the periodic table and the key
	spacerKey := "spacer"
	graph, _, _ = d2oracle.Create(graph, nil, "spacer")
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.label", spacerKey), nil, &empty)
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.stroke-width", spacerKey), nil, ptr.To("0"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.width", spacerKey), nil, ptr.To("50"))

	//<!-- Parent Container (Key) -->
	key := "key"
	graph, _, _ = d2oracle.Create(graph, nil, key)
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.grid-gap", key), nil, ptr.To("10"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.grid-columns", key), nil, ptr.To("1"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.style.border-radius", key), nil, ptr.To("8"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.Deployment.style.border-radius", key), nil, ptr.To("8"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.Deployment.style.fill", key), nil, ptr.To(theme["Deployment"]))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.StatefulSet.style.border-radius", key), nil, ptr.To("8"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.StatefulSet.style.fill", key), nil, ptr.To(theme["StatefulSet"]))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.DaemonSet.style.border-radius", key), nil, ptr.To("8"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.DaemonSet.style.fill", key), nil, ptr.To(theme["DaemonSet"]))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.Job.style.border-radius", key), nil, ptr.To("8"))
	graph, _ = d2oracle.Set(graph, nil, fmt.Sprintf("%s.Job.style.fill", key), nil, ptr.To(theme["Job"]))

	out := d2format.Format(graph.AST)
	// `Set` takes a *string as such it tries to quote `"[base; block]"`. This is problematic, by quoting the first
	// class is ignored hence, we do some ugly hack and remove the quotes, leaving us with `[base; block]`.
	out = strings.Replace(out, `"<`, "", -1)
	out = strings.Replace(out, `>"`, "", -1)

	f, err := os.OpenFile("output.d2", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = f.WriteString(out)

}

func newGraph() *d2graph.Graph {
	ruler, _ := textmeasure.NewRuler()
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	compileOpts := &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}
	ctx := tlog.WithDefault(context.Background())
	_, graph, _ := d2lib.Compile(ctx, "", compileOpts, nil)
	return graph
}
