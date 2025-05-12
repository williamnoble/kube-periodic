package main

import (
	"context"
	"flag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"sync"
)

func newClient() *kubernetes.Clientset {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	return clientset
}

func meta(
	clientset *kubernetes.Clientset,
	wg *sync.WaitGroup,
	ctx context.Context,
	workloadsChan chan Workload,
) {

	// Deployments
	wg.Add(1)
	go func() {
		defer wg.Done()
		deployments, err := clientset.AppsV1().Deployments(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing deployments: %s", err.Error())
			return
		}
		for _, d := range deployments.Items {
			workloadMeta := Workload{
				Name:      d.Name,
				Type:      DeploymentType,
				Namespace: d.Namespace,
			}
			workloadsChan <- workloadMeta
		}
	}()

	// StatefulSets
	wg.Add(1)
	go func() {
		defer wg.Done()
		statefulSets, err := clientset.AppsV1().StatefulSets(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing deployments: %s", err.Error())
			return
		}
		for _, s := range statefulSets.Items {
			workloadMeta := Workload{
				Name:      s.Name,
				Type:      StatefulSetType,
				Namespace: s.Namespace,
			}
			workloadsChan <- workloadMeta
		}
	}()

	// DaemonSets
	wg.Add(1)
	go func() {
		defer wg.Done()
		daemonSets, err := clientset.AppsV1().DaemonSets(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error fetching daemonsets: %s", err.Error())
			return
		}
		for _, d := range daemonSets.Items {
			workloadMeta := Workload{
				Name:      d.Name,
				Type:      DaemonSetType,
				Namespace: d.Namespace,
			}
			workloadsChan <- workloadMeta
		}
	}()

	// Jobs
	wg.Add(1)
	go func() {
		defer wg.Done()
		jobs, err := clientset.BatchV1().Jobs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error fetching jobs: %s", err.Error())
			return
		}
		for _, j := range jobs.Items {
			workloadMeta := Workload{
				Name:      j.Name,
				Type:      JobSetType,
				Namespace: j.Namespace,
			}
			workloadsChan <- workloadMeta
		}
	}()
}
