package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	pulokv1 "github.com/PulokSaha0706/my-controller/pkg/apis/pulok.dev/v1alpha1"
	clientset "github.com/PulokSaha0706/my-controller/pkg/generated/clientset/versioned"
	crdinformers "github.com/PulokSaha0706/my-controller/pkg/generated/informers/externalversions"
)

func int32Ptr(i int32) *int32 { return &i }

func intstrPtr(i int) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: int32(i)}
}

func recreateDeployment(kubeClient *kubernetes.Clientset, kluster *pulokv1.Kluster) {
	replicas := kluster.Spec.Replicas
	image := kluster.Spec.Image
	port := kluster.Spec.Port
	if port == 0 {
		port = 9090
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bookapi-deployment",
			Namespace: "default",
			Labels:    map[string]string{"app": "bookapi"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "bookapi"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "bookapi"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "bookapi-container",
							Image:   image,
							Command: []string{"./BookApi", "start", "-p", fmt.Sprint(port)},
							Ports: []corev1.ContainerPort{
								{ContainerPort: port},
							},
						},
					},
				},
			},
		},
	}

	_, err := kubeClient.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		fmt.Println("❌ Error creating Deployment:", err)
	} else {
		fmt.Println("✅ Deployment recreated")
	}
}

func recreateService(kubeClient *kubernetes.Clientset, kluster *pulokv1.Kluster) {
	port := kluster.Spec.Port
	if port == 0 {
		port = 9090
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bookapi-service",
			Namespace: "default",
			Labels:    map[string]string{"app": "bookapi"},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": "bookapi",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstrPtr(int(port)),
				},
			},
		},
	}

	_, err := kubeClient.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		fmt.Println("❌ Error creating Service:", err)
	} else {
		fmt.Println("✅ Service recreated")
	}
}

func main() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "Path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	customClient, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Custom Kluster Informer
	crdFactory := crdinformers.NewSharedInformerFactory(customClient, 0)
	klusterInformer := crdFactory.Pulokdev().V1alpha1().Klusters().Informer()

	// Core Kubernetes Informers
	coreFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	deploymentInformer := coreFactory.Apps().V1().Deployments().Informer()
	serviceInformer := coreFactory.Core().V1().Services().Informer()

	klusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			kluster := obj.(*pulokv1.Kluster)
			fmt.Println("🟢 Kluster Created:", kluster.Name)
			recreateDeployment(kubeClient, kluster)
			recreateService(kubeClient, kluster)
		},
		DeleteFunc: func(obj interface{}) {
			kluster := obj.(*pulokv1.Kluster)
			fmt.Println("🔴 Kluster Deleted:", kluster.Name)
			// Optional: Delete associated resources
		},
	})

	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			deploy := obj.(*appsv1.Deployment)
			if deploy.Labels["app"] == "bookapi" {
				fmt.Println("🔁 Deployment deleted — self-healing triggered")
				klusters, err := customClient.PulokdevV1alpha1().Klusters("default").List(context.TODO(), metav1.ListOptions{})
				if err == nil && len(klusters.Items) > 0 {
					recreateDeployment(kubeClient, &klusters.Items[0])
				}
			}
		},
	})

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			svc := obj.(*corev1.Service)
			if svc.Labels["app"] == "bookapi" {
				fmt.Println("🔁 Service deleted — self-healing triggered")
				klusters, err := customClient.PulokdevV1alpha1().Klusters("default").List(context.TODO(), metav1.ListOptions{})
				if err == nil && len(klusters.Items) > 0 {
					recreateService(kubeClient, &klusters.Items[0])
				}
			}
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	crdFactory.Start(stopCh)
	coreFactory.Start(stopCh)

	<-stopCh
}
