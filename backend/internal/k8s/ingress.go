package k8s

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sIngress "gkube/pkg/k8s/ingress"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ingress struct {
}

var Ingress = new(ingress)

// GetIngressList
//
//	@Description: 获取ingress
//	@receiver i
//	@param c
func (i *ingress) GetIngressList(c *gin.Context) {
	var query IngressQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingressList, err := k8sIngress.GetIngressList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressList)
}

// GetIngressByName
//
//	@Description: 获取ingress根据名称
//	@receiver i
//	@param c
func (i *ingress) GetIngressByName(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingress, err := k8sIngress.GetIngressByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingress)
}

// GetIngressYaml
//
//	@Description: 获取ingress的yaml
//	@receiver i
//	@param c
func (i *ingress) GetIngressYaml(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ingressYaml, err := k8sIngress.GetIngressYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressYaml)
}

// CreateIngress
//
//	@Description: 创建ingress
//	@receiver i
//	@param c
func (i *ingress) CreateIngress(c *gin.Context) {
	var body IngressCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.CreateIngress(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateIngress
//
//	@Description: 更新ingress
//	@receiver i
//	@param c
func (i *ingress) UpdateIngress(c *gin.Context) {
	var body IngressUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.UpdateIngress(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteIngressByName
//
//	@Description: 删除ingress根据名称
//	@receiver i
//	@param c
func (i *ingress) DeleteIngressByName(c *gin.Context) {
	var body IngressDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sIngress.DeleteIngressByName(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetIngressEvents
//
//	@Description: 获取ingress事件
//	@receiver i
//	@param c
func (i *ingress) GetIngressEvents(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Ingress", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress事件失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
		})
	}
	response.Success(c, "执行成功", result)
}

// CheckIngressTLSCertStatus
//
//	@Description: 检查 Ingress TLS 证书状态
//	@receiver i
//	@param c
func (i *ingress) CheckIngressTLSCertStatus(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	ing, err := k8sIngress.GetIngressByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}

	var results []map[string]any
	secretCache := make(map[string]map[string][]byte)

	for _, tls := range ing.Spec.TLS {
		entry := map[string]any{
			"hosts":      tls.Hosts,
			"secretName": tls.SecretName,
		}
		if tls.SecretName == "" {
			entry["status"] = "unknown"
			entry["message"] = "未配置 Secret"
			results = append(results, entry)
			continue
		}

		// Use cached secret data if already fetched (wildcard cert shared across TLS entries)
		secretData, cached := secretCache[tls.SecretName]
		if !cached {
			secret, err := client.CoreV1().Secrets(query.Namespace).Get(context.TODO(), tls.SecretName, metav1.GetOptions{})
			if err != nil {
				entry["status"] = "error"
				entry["message"] = fmt.Sprintf("Secret %s 不存在或无法访问", tls.SecretName)
				results = append(results, entry)
				continue
			}
			secretData = secret.Data
			secretCache[tls.SecretName] = secretData
		}

		certPEM, ok := secretData["tls.crt"]
		if !ok {
			entry["status"] = "error"
			entry["message"] = "Secret 中缺少 tls.crt"
			results = append(results, entry)
			continue
		}

		// pem.Decode returns the first PEM block; for a cert chain we only need the leaf cert
		block, _ := pem.Decode(certPEM)
		if block == nil {
			entry["status"] = "error"
			entry["message"] = "无法解析 PEM 证书"
			results = append(results, entry)
			continue
		}

		certs, err := x509.ParseCertificates(block.Bytes)
		if err != nil {
			entry["status"] = "error"
			entry["message"] = fmt.Sprintf("解析证书失败: %v", err)
			results = append(results, entry)
			continue
		}

		if len(certs) == 0 {
			entry["status"] = "error"
			entry["message"] = "证书链为空"
			results = append(results, entry)
			continue
		}

		entry["notBefore"] = certs[0].NotBefore.Format(time.RFC3339)
		entry["notAfter"] = certs[0].NotAfter.Format(time.RFC3339)
		entry["issuer"] = certs[0].Issuer.CommonName
		entry["subject"] = certs[0].Subject.CommonName

		now := time.Now()
		if now.After(certs[0].NotAfter) {
			entry["status"] = "expired"
			entry["message"] = fmt.Sprintf("证书已于 %s 过期", certs[0].NotAfter.Format("2006-01-02 15:04:05"))
		} else if now.Add(30 * 24 * time.Hour).After(certs[0].NotAfter) {
			daysLeft := int(certs[0].NotAfter.Sub(now).Hours() / 24)
			entry["status"] = "expiring"
			entry["message"] = fmt.Sprintf("证书将在 %d 天后过期", daysLeft)
		} else {
			daysLeft := int(certs[0].NotAfter.Sub(now).Hours() / 24)
			entry["status"] = "valid"
			entry["message"] = fmt.Sprintf("证书有效，剩余 %d 天", daysLeft)
		}
		results = append(results, entry)
	}
	response.Success(c, "执行成功", results)
}

type IngressQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type IngressUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type IngressDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
