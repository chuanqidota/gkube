package event

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type KubeEvent struct {
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	InvolvedObject ObjectRef `json:"involvedObject"`
	FirstTimestamp string    `json:"firstTimestamp,omitempty"`
	LastTimestamp  string    `json:"lastTimestamp,omitempty"`
	Count          int32     `json:"count"`
	Source         Source    `json:"source"`
}

type ObjectRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	UID       string `json:"uid,omitempty"`
}

type Source struct {
	Component string `json:"component"`
	Host      string `json:"host,omitempty"`
}

func ListEvents(client *kubernetes.Clientset, namespace, fieldSelector string, limit int64, continueToken string) ([]KubeEvent, string, string, error) {
	opts := metav1.ListOptions{FieldSelector: fieldSelector}
	if limit > 0 {
		opts.Limit = limit
	}
	if continueToken != "" {
		opts.Continue = continueToken
	}
	list, err := client.CoreV1().Events(namespace).List(context.TODO(), opts)
	if err != nil {
		return nil, "", "", err
	}
	events := make([]KubeEvent, 0, len(list.Items))
	for _, e := range list.Items {
		events = append(events, toKubeEvent(e))
	}
	return events, list.Continue, list.ResourceVersion, nil
}

func WatchEvents(client *kubernetes.Clientset, namespace, fieldSelector string) (watch.Interface, error) {
	opts := metav1.ListOptions{FieldSelector: fieldSelector}
	return client.CoreV1().Events(namespace).Watch(context.TODO(), opts)
}

func toKubeEvent(e corev1.Event) KubeEvent {
	ft := ""
	if !e.FirstTimestamp.IsZero() {
		ft = e.FirstTimestamp.Time.Format("2006-01-02 15:04:05")
	}
	lt := ""
	if !e.LastTimestamp.IsZero() {
		lt = e.LastTimestamp.Time.Format("2006-01-02 15:04:05")
	}
	return KubeEvent{
		Type:    e.Type,
		Reason:  e.Reason,
		Message: e.Message,
		InvolvedObject: ObjectRef{
			Kind:      e.InvolvedObject.Kind,
			Name:      e.InvolvedObject.Name,
			Namespace: e.InvolvedObject.Namespace,
			UID:       string(e.InvolvedObject.UID),
		},
		FirstTimestamp: ft,
		LastTimestamp:  lt,
		Count:          e.Count,
		Source: Source{
			Component: e.Source.Component,
			Host:      e.Source.Host,
		},
	}
}