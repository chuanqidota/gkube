package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sCrd "gkube/pkg/k8s/crd"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type crd struct{}

var Crd = new(crd)

func (c *crd) GetCRDList(ginCtx *gin.Context) {
	clusterName := ginCtx.Query("clusterName")
	client, err := k8s.GetApiExtensionsClientByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	crdList, err := k8sCrd.GetCRDList(client)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取CRD列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, crd := range crdList {
		var versions []string
		for _, v := range crd.Spec.Versions {
			versions = append(versions, v.Name)
		}
		var scope string
		if crd.Spec.Scope == "Namespaced" {
			scope = "Namespaced"
		} else {
			scope = "Cluster"
		}
		var categories []string
		for _, v := range crd.Spec.Names.Categories {
			categories = append(categories, v)
		}
		result = append(result, map[string]any{
			"name":       crd.Name,
			"group":      crd.Spec.Group,
			"versions":   versions,
			"kind":       crd.Spec.Names.Kind,
			"plural":     crd.Spec.Names.Plural,
			"scope":      scope,
			"categories": categories,
			"age":        crd.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(ginCtx, "执行成功", result)
}

func (c *crd) GetCRDDetail(ginCtx *gin.Context) {
	name := ginCtx.Query("name")
	clusterName := ginCtx.Query("clusterName")
	if name == "" {
		response.Fail(ginCtx, "name参数不能为空")
		return
	}
	client, err := k8s.GetApiExtensionsClientByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	crd, err := k8sCrd.GetCRDDetail(client, name)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取CRD详情失败:%s", err.Error()))
		return
	}
	response.Success(ginCtx, "执行成功", crd)
}

func (c *crd) GetCRDYaml(ginCtx *gin.Context) {
	name := ginCtx.Query("name")
	clusterName := ginCtx.Query("clusterName")
	if name == "" {
		response.Fail(ginCtx, "name参数不能为空")
		return
	}
	client, err := k8s.GetApiExtensionsClientByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sCrd.GetCRDYaml(client, name)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取CRD YAML失败:%s", err.Error()))
		return
	}
	response.Success(ginCtx, "执行成功", map[string]string{"yaml": yamlContent})
}

func (c *crd) GetCustomResourceList(ginCtx *gin.Context) {
	group := ginCtx.Query("group")
	version := ginCtx.Query("version")
	resource := ginCtx.Query("resource")
	namespace := ginCtx.Query("namespace")
	clusterName := ginCtx.Query("clusterName")
	if group == "" || version == "" || resource == "" {
		response.Fail(ginCtx, "group, version, resource参数不能为空")
		return
	}
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s配置失败:%s", err.Error()))
		return
	}
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	items, err := k8sCrd.GetCustomResourceList(config, gvr, namespace)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取自定义资源列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"name":      item.GetName(),
			"namespace": item.GetNamespace(),
			"age":       item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":    item.GetLabels(),
		})
	}
	response.Success(ginCtx, "执行成功", result)
}

func (c *crd) GetCustomResourceYaml(ginCtx *gin.Context) {
	group := ginCtx.Query("group")
	version := ginCtx.Query("version")
	resource := ginCtx.Query("resource")
	namespace := ginCtx.Query("namespace")
	name := ginCtx.Query("name")
	clusterName := ginCtx.Query("clusterName")
	if group == "" || version == "" || resource == "" || name == "" {
		response.Fail(ginCtx, "参数不能为空")
		return
	}
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s配置失败:%s", err.Error()))
		return
	}
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	yamlContent, err := k8sCrd.GetCustomResourceYaml(config, gvr, namespace, name)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取自定义资源YAML失败:%s", err.Error()))
		return
	}
	response.Success(ginCtx, "执行成功", map[string]string{"yaml": yamlContent})
}

func (c *crd) DeleteCustomResource(ginCtx *gin.Context) {
	group := ginCtx.Query("group")
	version := ginCtx.Query("version")
	resource := ginCtx.Query("resource")
	namespace := ginCtx.Query("namespace")
	name := ginCtx.Query("name")
	clusterName := ginCtx.Query("clusterName")
	if group == "" || version == "" || resource == "" || name == "" {
		response.Fail(ginCtx, "参数不能为空")
		return
	}
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		response.Fail(ginCtx, fmt.Sprintf("获取k8s配置失败:%s", err.Error()))
		return
	}
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	if err := k8sCrd.DeleteCustomResource(config, gvr, namespace, name); err != nil {
		response.Fail(ginCtx, fmt.Sprintf("删除自定义资源失败:%s", err.Error()))
		return
	}
	response.Success(ginCtx, "删除自定义资源成功", nil)
}
