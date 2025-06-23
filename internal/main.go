package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Load config
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("Failed to decode config: %v", err)
	}

	// Get k8s client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	labelSelector := fmt.Sprintf("%s=%s", config.LabelKey, config.PrimaryLabelValue)

	for {
		primaryPod, err := getPodByLabel(clientset, config.Namespace, labelSelector)
		if err != nil {
			log.Printf("Error getting primary pod: %v", err)
			err = sendWecomAlert(config.WecomWebhook,
				fmt.Sprintf("Failed to get primary pod in namespace %s: %v", config.Namespace, err))
			if err != nil {
				log.Printf("Failed to send wecom alert: %v", err)
			}
			continue
		}

		// 获取当前服务指向的实例
		svc, err := clientset.CoreV1().Services(config.Namespace).Get(context.TODO(), config.ServiceName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error getting service: %v", err)
			continue
		}

		currentTarget := svc.Spec.Selector[config.LabelKey]

		// 只有当前指向primary且primary不健康时才切换到standby
		if currentTarget == config.PrimaryLabelValue && !isPodHealthy(primaryPod) {
			log.Printf("Primary pod is not healthy (Phase: %v, Ready: %v), switching to standby",
				primaryPod.Status.Phase, isPodHealthy(primaryPod))
			err = updateServiceSelector(clientset, config.Namespace, config.ServiceName,
				map[string]string{config.LabelKey: config.StandbyLabelValue})
			if err != nil {
				log.Printf("Error updating service selector: %v", err)
				err = sendWecomAlert(config.WecomWebhook,
					fmt.Sprintf("Failed to switch to standby in namespace %s: %v", config.Namespace, err))
				if err != nil {
					log.Printf("Failed to send wecom alert: %v", err)
				}
			} else {
				err = sendWecomAlert(config.WecomWebhook,
					fmt.Sprintf("Successfully switched to standby in namespace %s,svc name is %s, current target is %s", config.Namespace, config.ServiceName, config.StandbyLabelValue))
				if err != nil {
					log.Printf("Failed to send wecom alert: %v", err)
				}
			}
		} else {
			if currentTarget == config.StandbyLabelValue {
				log.Printf("Service is pointing to standby, keeping it as is (Primary Phase: %v, Ready: %v)",
					primaryPod.Status.Phase, isPodHealthy(primaryPod))
			} else {
				log.Printf("Service is pointing to %s (Primary Phase: %v, Ready: %v)",
					currentTarget, primaryPod.Status.Phase, isPodHealthy(primaryPod))
			}
		}

		time.Sleep(60 * time.Second)
	}
}

func updateServiceSelector(clientset *kubernetes.Clientset, namespace string, serviceName string, selector map[string]string) error {
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	service.Spec.Selector = selector

	_, err = clientset.CoreV1().Services(namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	return err
}

func getPodByLabel(clientset *kubernetes.Clientset, namespace string, labelSelector string) (*corev1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found with label %s", labelSelector)
	}
	return &pods.Items[0], nil
}

// 添加一个新的辅助函数来检查pod状态
func isPodHealthy(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	// 检查pod的ready状态
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
