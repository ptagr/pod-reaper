package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudflare/cfssl/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	lifetimeAnnotation string = "pod.kubernetes.io/lifetime"
)

func main() {

	log.Level = log.LevelDebug

	log.Infof("Hello from pod reaper! Hide all the pods!\n")

	var config *rest.Config
	var err error
	var maxReaperCount = maxReaperCountPerRun()
	var (
		reapEvicted  = reapEvictedPods()
		runAsCronJob = cronJob()
	)

	if !reapEvicted {
		log.Debugf("REAP_EVICTED_PODS not set. Not reaping evicted pods.")
	}

	if remoteExec() {
		log.Debug("Loading kubeconfig from in cluster config")
		config, err = rest.InClusterConfig()
	} else {
		var kubeconfig *string
		if home := homeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}

		flag.Parse()
		log.Infof("Loading kubeconfig from %s\n", *kubeconfig)

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		reaperNamespaces := namespaces()
		if len(reaperNamespaces) == 0 {
			panic("No namespace specified. Exiting.")
		}
		for _, ns := range reaperNamespaces {
			pods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}

			log.Infof("Checking %d pods in namespace %s\n", len(pods.Items), ns)
			killedPods := 0
			for _, v := range pods.Items {
				if val, ok := v.Annotations[lifetimeAnnotation]; ok {
					log.Debugf("pod %s : Found annotation %s with value %s\n", v.Name, lifetimeAnnotation, val)
					lifetime, _ := time.ParseDuration(val)
					if lifetime == 0 {
						log.Debugf("pod %s : provided value %s is incorrect\n", v.Name, val)
					} else if killedPods < maxReaperCount {
						log.Debugf("pod %s : %s\n", v.Name, v.CreationTimestamp)
						currentLifetime := time.Now().Sub(v.CreationTimestamp.Time)
						if currentLifetime > lifetime {
							log.Infof("pod %s : pod is past its lifetime and will be killed.\n", v.Name)
							err := clientset.CoreV1().Pods(v.Namespace).Delete(context.TODO(), v.Name, metav1.DeleteOptions{})
							if err != nil {
								panic(err.Error())
							}
							log.Infof("pod %s : pod killed.\n", v.Name)
							killedPods++
						}
					} else {
						log.Debugf("pod %s : max %d pods killed\n", v.Name, maxReaperCount)
					}
				}

				if reapEvicted && strings.Contains(v.Status.Reason, "Evicted") {
					log.Debugf("pod %s : pod is evicted and needs to be deleted", v.Name)
					err := clientset.CoreV1().Pods(v.Namespace).Delete(context.TODO(), v.Name, metav1.DeleteOptions{})
					if err != nil {
						panic(err.Error())
					}
					log.Infof("pod %s : pod killed.\n", v.Name)
					killedPods++
				}

			}

			log.Infof("Killed %d Old/Evicted Pods.", killedPods)
		}
		if !runAsCronJob {
			log.Infof("Now sleeping for %d seconds", int(sleepDuration().Seconds()))
			time.Sleep(sleepDuration())
		} else {
			break
		}
	}

}

func remoteExec() bool {
	if val, ok := os.LookupEnv("REMOTE_EXEC"); ok {
		boolVal, err := strconv.ParseBool(val)
		if err == nil {
			return boolVal
		} else {
			panic("REMOTE_EXEC var incorrectly set")
		}
	}
	panic("REMOTE_EXEC var not set")
}

func maxReaperCountPerRun() int {
	i, err := strconv.Atoi(os.Getenv("MAX_REAPER_COUNT_PER_RUN"))
	if err != nil {
		i = 30
	}
	return i
}

func reapEvictedPods() bool {
	if val, ok := os.LookupEnv("REAP_EVICTED_PODS"); ok {
		boolVal, err := strconv.ParseBool(val)
		if err == nil {
			return boolVal
		}
	}
	return false
}

func cronJob() bool {
	if val, ok := os.LookupEnv("CRON_JOB"); ok {
		boolVal, err := strconv.ParseBool(val)
		if err == nil {
			return boolVal
		}
	}
	return false
}

func sleepDuration() time.Duration {
	if h := os.Getenv("REAPER_INTERVAL_IN_SEC"); h != "" {
		s, _ := strconv.Atoi(h)
		return time.Duration(s) * time.Second
	}
	return 60 * time.Second
}

func namespaces() []string {
	if h := os.Getenv("REAPER_NAMESPACES"); h != "" {
		namespaces := strings.Split(h, ",")
		if len(namespaces) == 1 && strings.ToLower(namespaces[0]) == "all" {
			return []string{metav1.NamespaceAll}
		}
		return namespaces
	}
	return []string{}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
