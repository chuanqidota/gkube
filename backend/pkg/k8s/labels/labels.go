package labels

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"golang.org/x/sync/singleflight"
)

// LabelFilter represents a single label filter condition.
type LabelFilter struct {
	Key      string   `json:"key" binding:"required"`
	Operator string   `json:"operator" binding:"required,oneof== != in notin"`
	Values   []string `json:"values" binding:"required,min=1"`
}

// BuildLabelSelector converts a slice of LabelFilter into a Kubernetes label selector string.
// Returns an empty string if filters is nil or empty.
func BuildLabelSelector(filters []LabelFilter) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		switch f.Operator {
		case "=":
			if len(f.Values) != 1 {
				return "", fmt.Errorf("= 需要恰好1个值")
			}
			parts = append(parts, fmt.Sprintf("%s=%s", f.Key, f.Values[0]))
		case "!=":
			if len(f.Values) != 1 {
				return "", fmt.Errorf("!= 需要恰好1个值")
			}
			parts = append(parts, fmt.Sprintf("%s!=%s", f.Key, f.Values[0]))
		case "in":
			if len(f.Values) == 0 {
				return "", fmt.Errorf("in 至少需要1个值")
			}
			quoted := quoteSelectorValues(f.Values)
			parts = append(parts, fmt.Sprintf("%s in (%s)", f.Key, strings.Join(quoted, ",")))
		case "notin":
			if len(f.Values) == 0 {
				return "", fmt.Errorf("notin 至少需要1个值")
			}
			quoted := quoteSelectorValues(f.Values)
			parts = append(parts, fmt.Sprintf("%s notin (%s)", f.Key, strings.Join(quoted, ",")))
		default:
			return "", fmt.Errorf("不支持的操作符: %s", f.Operator)
		}
	}

	selector := strings.Join(parts, ",")
	// Validate with K8s parser to catch any illegal characters.
	if _, err := labels.Parse(selector); err != nil {
		return "", fmt.Errorf("非法的 label selector: %w", err)
	}
	return selector, nil
}

// quoteSelectorValues quotes values that contain K8s label-selector special characters.
func quoteSelectorValues(values []string) []string {
	quoted := make([]string, len(values))
	for i, v := range values {
		if strings.ContainsAny(v, ",() \"\\") {
			quoted[i] = fmt.Sprintf("%q", v)
		} else {
			quoted[i] = v
		}
	}
	return quoted
}

// LabelData holds the available label keys and per-key values for a resource type.
type LabelData struct {
	Keys   []string            `json:"keys"`
	Values map[string][]string `json:"values"`
}

type labelCache struct {
	data     LabelData
	expireAt time.Time
}

var cache sync.Map

var fetchGroup singleflight.Group

const cacheTTL = 30 * time.Second

// maxPages is the maximum number of list pages to fetch (200 items each, up to 1000 total).
const maxPages = 5

// maxValuesPerKey caps the number of values returned per label key.
const maxValuesPerKey = 50

// buildCacheKey constructs a deterministic cache key including clusterName.
func buildCacheKey(clusterName, namespace, resourceType string) string {
	ns := namespace
	if ns == "" {
		ns = "__cluster__"
	}
	return fmt.Sprintf("%s/%s/%s", clusterName, ns, resourceType)
}

// GetAvailableLabels discovers label keys and values from existing resources of the given type.
// Results are cached for 30 seconds keyed by clusterName/namespace/resourceType.
// Uses paginated listing (200 items per page, up to 5 pages = 1000 resources).
// Concurrent requests for the same key are coalesced via singleflight.
func GetAvailableLabels(ctx context.Context, client *kubernetes.Clientset, dynamicClient dynamic.Interface, aeClient *apiextensionsclientset.Clientset, clusterName, namespace, resourceType string) (LabelData, error) {
	cacheKey := buildCacheKey(clusterName, namespace, resourceType)

	if v, ok := cache.Load(cacheKey); ok {
		c := v.(labelCache)
		if time.Now().Before(c.expireAt) {
			return c.data, nil
		}
		cache.Delete(cacheKey)
	}

	// Coalesce concurrent fetches for the same cache key.
	result, err, _ := fetchGroup.Do(cacheKey, func() (any, error) {
		var data LabelData
		var fetchErr error

		if !isBuiltinResource(resourceType) {
			data, fetchErr = fetchCRDLabels(ctx, dynamicClient, namespace, resourceType)
		} else {
			data, fetchErr = fetchBuiltinLabels(ctx, client, dynamicClient, aeClient, namespace, resourceType)
		}
		if fetchErr != nil {
			return LabelData{}, fetchErr
		}

		cache.Store(cacheKey, labelCache{data: data, expireAt: time.Now().Add(cacheTTL)})
		return data, nil
	})
	if err != nil {
		return LabelData{}, err
	}
	return result.(LabelData), nil
}

// builtinResourceTypes is the set of known built-in K8s resource type strings.
var builtinResourceTypes = map[string]bool{
	"deployment": true, "statefulset": true, "daemonset": true,
	"pod": true, "service": true, "ingress": true,
	"configmap": true, "secret": true, "job": true, "cronjob": true,
	"replicaset": true, "horizontalpodautoscaler": true, "networkpolicy": true,
	"persistentvolumeclaim": true, "resourcequota": true, "limitrange": true,
	"namespace": true, "node": true, "persistentvolume": true,
	"storageclass": true, "volumesnapshot": true, "volumesnapshotclass": true,
	"customresourcedefinition": true,
}

// isBuiltinResource returns true if resourceType is a known built-in K8s resource.
func isBuiltinResource(resourceType string) bool {
	return builtinResourceTypes[resourceType]
}

// builtinGVRs maps resource type strings to their GroupVersionResource.
var builtinGVRs = map[string]schema.GroupVersionResource{
	"deployment":              {Group: "apps", Version: "v1", Resource: "deployments"},
	"statefulset":             {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"daemonset":               {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"pod":                     {Version: "v1", Resource: "pods"},
	"service":                 {Version: "v1", Resource: "services"},
	"ingress":                 {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"configmap":               {Version: "v1", Resource: "configmaps"},
	"secret":                  {Version: "v1", Resource: "secrets"},
	"job":                     {Group: "batch", Version: "v1", Resource: "jobs"},
	"cronjob":                 {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"replicaset":              {Group: "apps", Version: "v1", Resource: "replicasets"},
	"horizontalpodautoscaler": {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},
	"networkpolicy":           {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	"persistentvolumeclaim":   {Version: "v1", Resource: "persistentvolumeclaims"},
	"resourcequota":           {Version: "v1", Resource: "resourcequotas"},
	"limitrange":              {Version: "v1", Resource: "limitranges"},
	"namespace":               {Version: "v1", Resource: "namespaces"},
	"node":                    {Version: "v1", Resource: "nodes"},
	"persistentvolume":        {Version: "v1", Resource: "persistentvolumes"},
	"storageclass":            {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	"volumesnapshot":          {Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshots"},
	"volumesnapshotclass":     {Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshotclasses"},
	"customresourcedefinition": {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
}

// clusterScopedResources is the set of built-in resources that are cluster-scoped.
var clusterScopedResources = map[string]bool{
	"namespace": true, "node": true, "persistentvolume": true,
	"storageclass": true, "volumesnapshotclass": true, "customresourcedefinition": true,
}

// getGVR parses a resourceType string into a GroupVersionResource.
// Built-in resources are looked up from builtinGVRs.
// CRD instances use "group/version/resource" format.
func getGVR(resourceType string) (schema.GroupVersionResource, error) {
	if gvr, ok := builtinGVRs[resourceType]; ok {
		return gvr, nil
	}
	parts := strings.SplitN(resourceType, "/", 3)
	if len(parts) == 3 {
		return schema.GroupVersionResource{
			Group:    parts[0],
			Version:  parts[1],
			Resource: parts[2],
		}, nil
	}
	return schema.GroupVersionResource{}, fmt.Errorf("不支持的资源类型: %s", resourceType)
}

// fetchCRDLabels uses the dynamic client to discover labels from CRD instances.
// Attempts namespace-scoped first when namespace is provided, falls back to cluster-scoped.
// This handles both namespace-scoped and cluster-scoped CRDs correctly.
func fetchCRDLabels(ctx context.Context, dynamicClient dynamic.Interface, namespace, resourceType string) (LabelData, error) {
	gvr, err := getGVR(resourceType)
	if err != nil {
		return LabelData{}, err
	}

	// If namespace is provided, try namespace-scoped first (most common case).
	// If it fails or returns empty, fall back to cluster-scoped.
	if namespace != "" {
		ri := dynamicClient.Resource(gvr).Namespace(namespace)
		if data, err := collectLabelsFromDynamic(ctx, ri); err == nil && len(data.Keys) > 0 {
			return data, nil
		}
	}
	// Try cluster-scoped (either no namespace given, or namespace-scoped returned empty).
	ri := dynamicClient.Resource(gvr)
	return collectLabelsFromDynamic(ctx, ri)
}

// collectLabelsFromDynamic pages through a dynamic resource and collects all labels.
func collectLabelsFromDynamic(ctx context.Context, ri dynamic.ResourceInterface) (LabelData, error) {
	labelKeys := make(map[string]struct{})
	labelValues := make(map[string]map[string]struct{})

	collect := func(l map[string]string) {
		for k, v := range l {
			labelKeys[k] = struct{}{}
			if labelValues[k] == nil {
				labelValues[k] = make(map[string]struct{})
			}
			if len(labelValues[k]) < maxValuesPerKey {
				labelValues[k][v] = struct{}{}
			}
		}
	}

	continueToken := ""
	for page := 0; page < maxPages; page++ {
		listOpts := metav1.ListOptions{Limit: 200}
		if continueToken == "" {
			listOpts.ResourceVersion = "0"
		} else {
			listOpts.Continue = continueToken
		}

		list, err := ri.List(ctx, listOpts)
		if err != nil {
			return LabelData{}, err
		}

		for _, item := range list.Items {
			collect(item.GetLabels())
		}

		if list.GetContinue() == "" {
			break
		}
		continueToken = list.GetContinue()
	}

	return buildLabelData(labelKeys, labelValues), nil
}

// listLabelsFn is a function that returns labels from a paginated list.
// It receives a context for cancellation propagation.
// Returns (labels per item, continue token, error).
type listLabelsFn func(ctx context.Context, continueToken string) ([]map[string]string, string, error)

// makeListOpts creates ListOptions with pagination support.
func makeListOpts(continueToken string) metav1.ListOptions {
	opts := metav1.ListOptions{Limit: 200}
	if continueToken == "" {
		opts.ResourceVersion = "0"
	} else {
		opts.Continue = continueToken
	}
	return opts
}

// builtinListFnMap maps resource types to their list functions.
var builtinListFnMap = map[string]func(*kubernetes.Clientset, string) listLabelsFn{
	"deployment": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.AppsV1().Deployments(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"statefulset": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.AppsV1().StatefulSets(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"daemonset": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.AppsV1().DaemonSets(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"pod": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().Pods(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"service": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().Services(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"ingress": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.NetworkingV1().Ingresses(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"configmap": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().ConfigMaps(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"secret": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().Secrets(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"job": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.BatchV1().Jobs(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"cronjob": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.BatchV1().CronJobs(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"replicaset": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.AppsV1().ReplicaSets(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"horizontalpodautoscaler": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"networkpolicy": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.NetworkingV1().NetworkPolicies(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"persistentvolumeclaim": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().PersistentVolumeClaims(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"resourcequota": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().ResourceQuotas(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"limitrange": func(client *kubernetes.Clientset, ns string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().LimitRanges(ns).List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	// Cluster-scoped resources (ignore namespace)
	"namespace": func(client *kubernetes.Clientset, _ string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().Namespaces().List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"node": func(client *kubernetes.Clientset, _ string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().Nodes().List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"persistentvolume": func(client *kubernetes.Clientset, _ string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.CoreV1().PersistentVolumes().List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
	"storageclass": func(client *kubernetes.Clientset, _ string) listLabelsFn {
		return func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := client.StorageV1().StorageClasses().List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}
	},
}

// fetchBuiltinLabels uses the typed client to collect labels from built-in resources.
// Supports paginated listing (up to maxPages pages of 200).
func fetchBuiltinLabels(ctx context.Context, client *kubernetes.Clientset, dynamicClient dynamic.Interface, aeClient *apiextensionsclientset.Clientset, namespace, resourceType string) (LabelData, error) {
	labelKeys := make(map[string]struct{})
	labelValues := make(map[string]map[string]struct{})

	collect := func(l map[string]string) {
		for k, v := range l {
			labelKeys[k] = struct{}{}
			if labelValues[k] == nil {
				labelValues[k] = make(map[string]struct{})
			}
			if len(labelValues[k]) < maxValuesPerKey {
				labelValues[k][v] = struct{}{}
			}
		}
	}

	// Special cases for resources using dynamic client or apiextensions client
	switch resourceType {
	case "volumesnapshot":
		gvr := schema.GroupVersionResource{
			Group:    "snapshot.storage.k8s.io",
			Version:  "v1",
			Resource: "volumesnapshots",
		}
		if err := paginateDynamic(ctx, dynamicClient.Resource(gvr).Namespace(namespace), maxPages, collect); err != nil {
			return LabelData{}, err
		}
		return buildLabelData(labelKeys, labelValues), nil
	case "volumesnapshotclass":
		gvr := schema.GroupVersionResource{
			Group:    "snapshot.storage.k8s.io",
			Version:  "v1",
			Resource: "volumesnapshotclasses",
		}
		if err := paginateDynamic(ctx, dynamicClient.Resource(gvr), maxPages, collect); err != nil {
			return LabelData{}, err
		}
		return buildLabelData(labelKeys, labelValues), nil
	case "customresourcedefinition":
		if err := paginate(ctx, maxPages, func(ctx context.Context, ct string) ([]map[string]string, string, error) {
			list, err := aeClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, makeListOpts(ct))
			if err != nil {
				return nil, "", err
			}
			lbls := make([]map[string]string, len(list.Items))
			for i := range list.Items {
				lbls[i] = list.Items[i].Labels
			}
			return lbls, list.Continue, nil
		}, collect); err != nil {
			return LabelData{}, err
		}
		return buildLabelData(labelKeys, labelValues), nil
	}

	// Use function map for standard resources
	listFnFactory, ok := builtinListFnMap[resourceType]
	if !ok {
		return LabelData{}, fmt.Errorf("不支持的资源类型: %s", resourceType)
	}

	listFn := listFnFactory(client, namespace)
	if err := paginate(ctx, maxPages, listFn, collect); err != nil {
		return LabelData{}, err
	}

	return buildLabelData(labelKeys, labelValues), nil
}

// paginate is a generic pagination loop for typed client lists.
// listFn returns (labels per item, continue token, error).
// collectFn is called for each page's labels.
func paginate(ctx context.Context, maxPages int, listFn func(ctx context.Context, continueToken string) ([]map[string]string, string, error), collectFn func(map[string]string)) error {
	continueToken := ""
	for page := 0; page < maxPages; page++ {
		lbls, cont, err := listFn(ctx, continueToken)
		if err != nil {
			return err
		}
		for _, l := range lbls {
			collectFn(l)
		}
		if cont == "" {
			break
		}
		continueToken = cont
	}
	return nil
}

// paginateDynamic pages through a dynamic resource interface.
func paginateDynamic(ctx context.Context, ri dynamic.ResourceInterface, maxPages int, collectFn func(map[string]string)) error {
	continueToken := ""
	for page := 0; page < maxPages; page++ {
		listOpts := metav1.ListOptions{Limit: 200}
		if continueToken == "" {
			listOpts.ResourceVersion = "0"
		} else {
			listOpts.Continue = continueToken
		}

		list, err := ri.List(ctx, listOpts)
		if err != nil {
			return err
		}
		for _, item := range list.Items {
			collectFn(item.GetLabels())
		}
		if list.GetContinue() == "" {
			break
		}
		continueToken = list.GetContinue()
	}
	return nil
}

// buildLabelData converts raw label maps into the sorted LabelData response format.
func buildLabelData(labelKeys map[string]struct{}, labelValues map[string]map[string]struct{}) LabelData {
	data := LabelData{
		Keys:   make([]string, 0, len(labelKeys)),
		Values: make(map[string][]string, len(labelKeys)),
	}
	for k := range labelKeys {
		data.Keys = append(data.Keys, k)
	}
	sort.Strings(data.Keys)

	for k, vals := range labelValues {
		vs := make([]string, 0, len(vals))
		for v := range vals {
			vs = append(vs, v)
		}
		sort.Strings(vs)
		data.Values[k] = vs
	}
	return data
}
