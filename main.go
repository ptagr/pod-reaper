package main

import (
	"time"
	"flag"
	"path/filepath"
	"os"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"github.com/cloudflare/cfssl/log"
	"strconv"
	"k8s.io/client-go/rest"
)

const (
	LifetimeAnnotation string = "pod.kubernetes.io/lifetime"
)

func main(){

	log.Level = log.LevelInfo

	log.Infof("Hello from pod reaper! Hide all the pods!\n")



	var config *rest.Config = nil
	var err error = nil
	if(remoteExec() != "") {
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
		pods, err := clientset.CoreV1().Pods(namespace()).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		log.Infof("Checking %d pods in namespace %s\n", len(pods.Items), namespace())
		killedPods := 0
		for _,v := range pods.Items {
			if val, ok := v.Annotations[LifetimeAnnotation]; ok {
				log.Debugf("pod %s : Found annotation %s with value %s\n", v.Name, LifetimeAnnotation, val)
				lifetime, _ := time.ParseDuration(val)
				if lifetime == 0 {
					log.Debugf("pod %s : provided value %s is incorrect\n", v.Name, val)
				} else {
					log.Debugf("pod %s : %s\n", v.Name, v.CreationTimestamp)
					currentLifetime := time.Now().Sub(v.CreationTimestamp.Time)
					if currentLifetime > lifetime {
						log.Infof("pod %s : pod is past its lifetime and will be killed.\n", v.Name)
						err := clientset.CoreV1().Pods(v.Namespace).Delete(v.Name, &metav1.DeleteOptions{})
						if err != nil {
							panic(err.Error())
						}
						log.Infof("pod %s : pod killed.\n", v.Name)
						killedPods++
					}
				}
			}
		}

		log.Infof("Killed %d Old Pods. Now sleeping for %d seconds", killedPods, int(sleepDuration().Seconds()))
		time.Sleep(sleepDuration())
	}

}

func remoteExec() string {
	return os.Getenv("REMOTE_EXEC")
}

func sleepDuration() time.Duration {
	if h := os.Getenv("REAPER_INTERVAL_IN_SEC"); h != "" {
		s,_ := strconv.Atoi(h)
		return time.Duration(s) * time.Second
	}
	return 60 * time.Second
}

func namespace() string {
	if h := os.Getenv("REAPER_NAMESPACE"); h != "" {
		return h
	}
	return metav1.NamespaceAll
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}