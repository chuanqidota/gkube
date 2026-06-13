package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type tenancyHandler struct{}

var Tenancy = new(tenancyHandler)

type Tenant struct {
	Name        string            `json:"name"`
	Namespaces  []string          `json:"namespaces"`
	Users       []string          `json:"users"`
	Quotas      *TenantQuotas     `json:"quotas"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type TenantQuotas struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
	Pods    int    `json:"pods"`
}

type CreateTenantRequest struct {
	ClusterName string            `json:"clusterName" form:"cluster"`
	Name        string            `json:"name"`
	Users       []string          `json:"users"`
	Quotas      *TenantQuotas     `json:"quotas"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ListTenants lists all tenants (namespaces with tenant label)
func (h *tenancyHandler) ListTenants(c *gin.Context) {
	clusterName := c.Query("cluster")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "gkube.io/tenant",
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取租户列表失败: %s", err.Error()))
		return
	}

	tenants := make(map[string]*Tenant)
	for _, ns := range namespaces.Items {
		tenantName := ns.Labels["gkube.io/tenant"]
		if _, exists := tenants[tenantName]; !exists {
			tenants[tenantName] = &Tenant{
				Name:        tenantName,
				Namespaces:  []string{},
				Users:       []string{},
				Labels:      ns.Labels,
				Annotations: ns.Annotations,
			}
		}
		tenants[tenantName].Namespaces = append(tenants[tenantName].Namespaces, ns.Name)
	}

	var result []*Tenant
	for _, t := range tenants {
		result = append(result, t)
	}

	response.Success(c, "获取成功", result)
}

// GetTenant gets details of a specific tenant
func (h *tenancyHandler) GetTenant(c *gin.Context) {
	tenantName := c.Query("name")
	clusterName := c.Query("cluster")

	if tenantName == "" {
		response.Fail(c, "租户名称不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	// Get namespaces for this tenant
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("gkube.io/tenant=%s", tenantName),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取租户信息失败: %s", err.Error()))
		return
	}

	if len(namespaces.Items) == 0 {
		response.Fail(c, "租户不存在")
		return
	}

	tenant := &Tenant{
		Name:        tenantName,
		Namespaces:  []string{},
		Users:       []string{},
		Labels:      namespaces.Items[0].Labels,
		Annotations: namespaces.Items[0].Annotations,
	}

	for _, ns := range namespaces.Items {
		tenant.Namespaces = append(tenant.Namespaces, ns.Name)
	}

	// Get users from role bindings
	for _, ns := range namespaces.Items {
		roleBindings, _ := client.RbacV1().RoleBindings(ns.Name).List(context.TODO(), metav1.ListOptions{})
		for _, rb := range roleBindings.Items {
			for _, subject := range rb.Subjects {
				if subject.Kind == "User" {
					tenant.Users = append(tenant.Users, subject.Name)
				}
			}
		}
	}

	// Get quotas
	quotas, _ := client.CoreV1().ResourceQuotas(namespaces.Items[0].Name).List(context.TODO(), metav1.ListOptions{})
	if len(quotas.Items) > 0 {
		q := quotas.Items[0]
		tenant.Quotas = &TenantQuotas{
			CPU:     q.Spec.Hard.Cpu().String(),
			Memory:  q.Spec.Hard.Memory().String(),
			Storage: q.Spec.Hard.Storage().String(),
			Pods:    int(q.Spec.Hard.Pods().Value()),
		}
	}

	response.Success(c, "获取成功", tenant)
}

// CreateTenant creates a new tenant
func (h *tenancyHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if req.Name == "" {
		response.Fail(c, "租户名称不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	// Create namespace for tenant
	nsName := req.Name
	labels := map[string]string{
		"gkube.io/tenant": req.Name,
	}
	for k, v := range req.Labels {
		labels[k] = v
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nsName,
			Labels:      labels,
			Annotations: req.Annotations,
		},
	}

	_, err = client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建命名空间失败: %s", err.Error()))
		return
	}

	// Create resource quota if specified
	if req.Quotas != nil {
		quota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name + "-quota",
				Namespace: nsName,
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{},
			},
		}

		if req.Quotas.CPU != "" {
			quota.Spec.Hard[corev1.ResourceCPU] = resource.MustParse(req.Quotas.CPU)
		}
		if req.Quotas.Memory != "" {
			quota.Spec.Hard[corev1.ResourceMemory] = resource.MustParse(req.Quotas.Memory)
		}
		if req.Quotas.Storage != "" {
			quota.Spec.Hard[corev1.ResourceStorage] = resource.MustParse(req.Quotas.Storage)
		}
		if req.Quotas.Pods > 0 {
			quota.Spec.Hard[corev1.ResourcePods] = resource.MustParse(fmt.Sprintf("%d", req.Quotas.Pods))
		}

		_, err = client.CoreV1().ResourceQuotas(nsName).Create(context.TODO(), quota, metav1.CreateOptions{})
		if err != nil {
			response.Fail(c, fmt.Sprintf("创建资源配额失败: %s", err.Error()))
			return
		}
	}

	// Create role binding for users
	if len(req.Users) > 0 {
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name + "-admin",
				Namespace: nsName,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "admin",
			},
			Subjects: []rbacv1.Subject{},
		}

		for _, user := range req.Users {
			roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     user,
			})
		}

		_, err = client.RbacV1().RoleBindings(nsName).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
		if err != nil {
			response.Fail(c, fmt.Sprintf("创建角色绑定失败: %s", err.Error()))
			return
		}
	}

	response.Success(c, "租户创建成功", nil)
}

// DeleteTenant deletes a tenant and all its namespaces
func (h *tenancyHandler) DeleteTenant(c *gin.Context) {
	tenantName := c.Query("name")
	clusterName := c.Query("cluster")

	if tenantName == "" {
		response.Fail(c, "租户名称不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	// Get namespaces for this tenant
	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("gkube.io/tenant=%s", tenantName),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取租户信息失败: %s", err.Error()))
		return
	}

	// Delete all namespaces
	for _, ns := range namespaces.Items {
		err = client.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
		if err != nil {
			response.Fail(c, fmt.Sprintf("删除命名空间失败: %s", err.Error()))
			return
		}
	}

	response.Success(c, "租户删除成功", nil)
}

// AddNamespaceToTenant adds a namespace to a tenant
func (h *tenancyHandler) AddNamespaceToTenant(c *gin.Context) {
	tenantName := c.Query("tenant")
	namespace := c.Query("namespace")
	clusterName := c.Query("cluster")

	if tenantName == "" || namespace == "" {
		response.Fail(c, "租户名称和命名空间不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	// Update namespace labels
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取命名空间失败: %s", err.Error()))
		return
	}

	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["gkube.io/tenant"] = tenantName

	_, err = client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新命名空间失败: %s", err.Error()))
		return
	}

	response.Success(c, "命名空间已添加到租户", nil)
}

// RemoveNamespaceFromTenant removes a namespace from a tenant
func (h *tenancyHandler) RemoveNamespaceFromTenant(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("cluster")

	if namespace == "" {
		response.Fail(c, "命名空间不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群客户端失败: %s", err.Error()))
		return
	}

	// Update namespace labels
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取命名空间失败: %s", err.Error()))
		return
	}

	if ns.Labels != nil {
		delete(ns.Labels, "gkube.io/tenant")
	}

	_, err = client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新命名空间失败: %s", err.Error()))
		return
	}

	response.Success(c, "命名空间已从租户移除", nil)
}
