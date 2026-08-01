# gkube

gkube 是一个 Kubernetes 多集群管理 Web 平台。

它的目标是把日常的集群运维操作从命令行搬到浏览器里——用一个统一的界面管理多套 K8s 集群,并对集群内各类资源做可视化的增删改查,降低使用门槛。

## 它想解决什么

- **多集群统一纳管**:把分散在不同环境的多套 K8s 集群接入同一个平台,通过界面切换,不必为每个集群单独配置 kubectl 和 kubeconfig。
- **资源的图形化管理**:对工作负载(Pod、Deployment、StatefulSet、DaemonSet、Job/CronJob)、网络(Service、Ingress、NetworkPolicy)、存储(PV/PVC、StorageClass)、配置(ConfigMap、Secret、ResourceQuota)、节点、命名空间等主要资源类型提供列表、详情、YAML 编辑、创建和删除等操作。
- **运维常用工具**:内置 Web 终端(直接在浏览器里执行 kubectl 式命令)和实时日志查看器,覆盖排障场景;同时提供集群健康状态总览与多维度仪表盘。

## 使用方式

平台采用账号登录(纯 Token 认证),登录后即可在已接入的集群间切换并管理资源。适合个人开发者自建,也适合小团队作为轻量级的 K8s 管理入口。
