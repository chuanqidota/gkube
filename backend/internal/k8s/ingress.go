package k8s

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sIngress "gkube/pkg/k8s/ingress"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetIngressByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sIngress.GetIngressByName(client, namespace, name)
	},
	"执行成功", "查询ingress失败",
)

var GetIngressYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sIngress.GetIngressYaml(client, namespace, name)
	},
	"执行成功", "获取ingress失败",
)

var CreateIngress = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sIngress.CreateIngress(client, namespace, yaml)
	},
	"执行成功", "创建ingress失败",
)

var DeleteIngressByName = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sIngress.DeleteIngressByName(client, namespace, name)
	},
	"执行成功", "删除ingress失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetIngressList 列表 —— 非分页
func GetIngressList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	ingressList, err := k8sIngress.GetIngressList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressList)
}

// UpdateIngress 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateIngress(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.UpdateIngress(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetIngressEvents 事件 —— 内联 K8s 调用
func GetIngressEvents(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(p.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("involvedObject.name", p.Name),
			fields.OneTermEqualSelector("involvedObject.kind", "Ingress"),
		).String(),
	})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取ingress事件失败:%s", err.Error()))
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

// CheckIngressTLSCertStatus 检查 TLS 证书状态 —— 复杂内联逻辑
func CheckIngressTLSCertStatus(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	ing, err := k8sIngress.GetIngressByName(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("查询ingress失败:%v", err.Error()))
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
		secretData, cached := secretCache[tls.SecretName]
		if !cached {
			secret, err := client.CoreV1().Secrets(p.Namespace).Get(context.TODO(), tls.SecretName, metav1.GetOptions{})
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

// GetIngressClassList 返回集群中所有 IngressClass 名称
func GetIngressClassList(c *gin.Context) {
	var query struct {
		ClusterName string `form:"clusterName" binding:"required"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	names, err := k8sIngress.ListIngressClasses(client)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, "执行成功", names)
}
