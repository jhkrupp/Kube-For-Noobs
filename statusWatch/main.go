package main

import (
	//"context"
	//"reflect"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
)

func onPodAdd(obj interface{}){
	pod := obj.(*coreV1.Pod)
	fmt.Printf("New pod added with image %s\n", pod.Status.ContainerStatuses[0].Image)
}

func onPodDelete(obj interface{}){
	pod := obj.(*coreV1.Pod)
	fmt.Printf("Pod with image %s was deleted", pod.Status.ContainerStatuses)
}
/*
func onPodUpdate(old, new interface{}) {
	oldPod = old.(*coreV1.Pod)
	newPod = new.(*coreV1.Pod)
	fmt.Printf("Pod originally had image %s, but was updated with image %s", oldPod.Status.ContainerStatuses, newPod.Status.ContainerStatuses)
}
*/
func newPodInformer(clientset *kubernetes.Clientset, ns string) {
	/*if ns == "" {
		factory := informers.NewSharedInformerFactory(clientset, 0)
	} else {
		factory := informers.NewSharedInformerFactory(clientset, 0)
	}*/
	factory := informers.NewSharedInformerFactory(clientset, 1000000000)
	informer := factory.Core().V1().Pods().Informer()
	fmt.Println("Controller has started")
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()
	defer fmt.Println("Controller has stopped")
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc : onPodAdd,
        DeleteFunc: onPodDelete,
        //UpdateFunc: onPodUpdate,
    })
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
        runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
        return
    }
    <-stopper
}

func main() {
	var ns string
	flag.StringVar(&ns, "namespace", "", "k8s cluster namespace")
	flag.Parse()
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	log.Println("using namespace: ", ns)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	newPodInformer(clientset, ns)
}