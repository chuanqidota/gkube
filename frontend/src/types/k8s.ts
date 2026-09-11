// K8s resource type definitions
// Based on actual field accesses across Detail pages and API transform functions

// ============ 基础类型 ============

export interface K8sObjectMeta {
  name: string
  namespace?: string
  uid?: string
  resourceVersion?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  creationTimestamp: string
  deletionTimestamp?: string
}

export interface K8sLabelSelector {
  matchLabels?: Record<string, string>
  matchExpressions?: { key: string; operator: string; values?: string[] }[]
}

export interface K8sCondition {
  type: string
  status: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

export interface K8sOwnerReference {
  apiVersion: string
  kind: string
  name: string
  uid: string
  controller?: boolean
  blockOwnerDeletion?: boolean
}

// ============ 容器相关 ============

export interface K8sContainerPort {
  name?: string
  containerPort: number
  protocol?: string
  hostIP?: string
  hostPort?: number
}

export interface K8sEnvVar {
  name: string
  value?: string
  valueFrom?: Record<string, any>
}

export interface K8sResourceRequirements {
  requests?: { cpu?: string; memory?: string }
  limits?: { cpu?: string; memory?: string }
}

export interface K8sVolumeMount {
  name: string
  mountPath: string
  subPath?: string
  readOnly?: boolean
}

export interface K8sProbe {
  httpGet?: { path?: string; port: number | string; scheme?: string; host?: string }
  tcpSocket?: { port: number | string; host?: string }
  exec?: { command?: string[] }
  initialDelaySeconds?: number
  periodSeconds?: number
  timeoutSeconds?: number
  successThreshold?: number
  failureThreshold?: number
}

export interface K8sLifecycleHandler {
  exec?: { command?: string[] }
  httpGet?: { path?: string; port: number | string; scheme?: string; host?: string }
  tcpSocket?: { port: number | string; host?: string }
}

export interface K8sLifecycle {
  postStart?: K8sLifecycleHandler
  preStop?: K8sLifecycleHandler
}

export interface K8sSecurityContext {
  runAsUser?: number
  runAsGroup?: number
  runAsNonRoot?: boolean
  privileged?: boolean
  readOnlyRootFilesystem?: boolean
  allowPrivilegeEscalation?: boolean
  capabilities?: K8sCapabilities
}

export interface K8sCapabilities {
  add?: string[]
  drop?: string[]
}

export interface K8sContainer {
  name: string
  image: string
  command?: string[]
  args?: string[]
  ports?: K8sContainerPort[]
  env?: K8sEnvVar[]
  resources?: K8sResourceRequirements
  volumeMounts?: K8sVolumeMount[]
  livenessProbe?: K8sProbe
  readinessProbe?: K8sProbe
  startupProbe?: K8sProbe
  lifecycle?: K8sLifecycle
  securityContext?: K8sSecurityContext
  imagePullPolicy?: string
}

// ============ Pod 相关 ============

export interface K8sToleration {
  key?: string
  operator?: string
  value?: string
  effect?: string
  tolerationSeconds?: number
}

export interface K8sAffinity {
  nodeAffinity?: Record<string, any>
  podAffinity?: Record<string, any>
  podAntiAffinity?: Record<string, any>
}

export interface K8sTopologySpreadConstraint {
  maxSkew: number
  topologyKey: string
  whenUnsatisfiable: string
  labelSelector?: K8sLabelSelector
  matchLabelKeys?: string[]
  minDomains?: number
}

export interface K8sPodSpec {
  containers: K8sContainer[]
  initContainers?: K8sContainer[]
  volumes?: K8sVolume[]
  nodeSelector?: Record<string, string>
  tolerations?: K8sToleration[]
  affinity?: K8sAffinity
  topologySpreadConstraints?: K8sTopologySpreadConstraint[]
  serviceAccountName?: string
  hostNetwork?: boolean
  dnsPolicy?: string
  restartPolicy?: string
  terminationGracePeriodSeconds?: number
  imagePullSecrets?: { name: string }[]
  priorityClassName?: string
  securityContext?: K8sSecurityContext
}

export interface K8sPodTemplateSpec {
  metadata?: K8sObjectMeta
  spec: K8sPodSpec
}

export interface K8sPodCondition {
  type: string
  status: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

export interface K8sContainerStatus {
  name: string
  ready: boolean
  restartCount: number
  image: string
  state?: Record<string, any>
  lastState?: Record<string, any>
}

export interface K8sPodStatus {
  phase?: string
  conditions?: K8sPodCondition[]
  hostIP?: string
  podIP?: string
  containerStatuses?: K8sContainerStatus[]
  initContainerStatuses?: K8sContainerStatus[]
}

export interface K8sPod {
  apiVersion: string
  kind: 'Pod'
  metadata: K8sObjectMeta
  spec: K8sPodSpec
  status?: K8sPodStatus
}

// ============ 存储相关 ============

export interface K8sVolume {
  name: string
  configMap?: { name: string; defaultMode?: number; items?: { key: string; path: string }[] }
  secret?: { secretName: string; defaultMode?: number; items?: { key: string; path: string }[] }
  emptyDir?: { medium?: string; sizeLimit?: string }
  hostPath?: { path: string; type?: string }
  persistentVolumeClaim?: { claimName: string; readOnly?: boolean }
  nfs?: { server: string; path: string; readOnly?: boolean }
  csi?: { driver: string; volumeAttributes?: Record<string, string> }
  projected?: { sources: Record<string, any>[]; defaultMode?: number }
  [key: string]: any
}

export interface K8sVolumeClaimTemplate {
  metadata?: Partial<K8sObjectMeta>
  spec: {
    accessModes?: string[]
    resources: { requests: { storage: string } }
    storageClassName?: string
    volumeMode?: string
  }
}

// ============ 工作负载 ============

export interface K8sDeploymentSpec {
  replicas?: number
  selector: K8sLabelSelector
  template: K8sPodTemplateSpec
  strategy?: {
    type?: string
    rollingUpdate?: { maxSurge?: number | string; maxUnavailable?: number | string }
  }
}

export interface K8sDeploymentStatus {
  replicas?: number
  readyReplicas?: number
  availableReplicas?: number
  updatedReplicas?: number
  conditions?: K8sCondition[]
}

export interface K8sDeployment {
  apiVersion: 'apps/v1'
  kind: 'Deployment'
  metadata: K8sObjectMeta
  spec: K8sDeploymentSpec
  status?: K8sDeploymentStatus
}

export interface K8sStatefulSetSpec {
  replicas?: number
  selector: K8sLabelSelector
  template: K8sPodTemplateSpec
  serviceName?: string
  podManagementPolicy?: string
  updateStrategy?: { type?: string; rollingUpdate?: { partition?: number } }
  volumeClaimTemplates?: K8sVolumeClaimTemplate[]
}

export interface K8sStatefulSetStatus {
  replicas?: number
  readyReplicas?: number
  updatedReplicas?: number
  currentReplicas?: number
  conditions?: K8sCondition[]
}

export interface K8sStatefulSet {
  apiVersion: 'apps/v1'
  kind: 'StatefulSet'
  metadata: K8sObjectMeta
  spec: K8sStatefulSetSpec
  status?: K8sStatefulSetStatus
}

export interface K8sDaemonSetSpec {
  selector: K8sLabelSelector
  template: K8sPodTemplateSpec
  updateStrategy?: {
    type?: string
    rollingUpdate?: { maxUnavailable?: number | string; maxSurge?: number | string }
  }
}

export interface K8sDaemonSetStatus {
  desiredNumberScheduled?: number
  currentNumberScheduled?: number
  numberReady?: number
  updatedNumberScheduled?: number
  numberAvailable?: number
  numberMisscheduled?: number
  conditions?: K8sCondition[]
}

export interface K8sDaemonSet {
  apiVersion: 'apps/v1'
  kind: 'DaemonSet'
  metadata: K8sObjectMeta
  spec: K8sDaemonSetSpec
  status?: K8sDaemonSetStatus
}

export interface K8sReplicaSetSpec {
  replicas?: number
  selector: K8sLabelSelector
  template: K8sPodTemplateSpec
  minReadySeconds?: number
}

export interface K8sReplicaSetStatus {
  replicas?: number
  readyReplicas?: number
  availableReplicas?: number
  fullyLabeledReplicas?: number
  conditions?: K8sCondition[]
}

export interface K8sReplicaSet {
  apiVersion: 'apps/v1'
  kind: 'ReplicaSet'
  metadata: K8sObjectMeta
  spec: K8sReplicaSetSpec
  status?: K8sReplicaSetStatus
}

export interface K8sJobSpec {
  completions?: number
  parallelism?: number
  backoffLimit?: number
  activeDeadlineSeconds?: number
  ttlSecondsAfterFinished?: number
  completionMode?: string
  template: K8sPodTemplateSpec
  selector?: K8sLabelSelector
}

export interface K8sJobStatus {
  active?: number
  succeeded?: number
  failed?: number
  conditions?: K8sCondition[]
  startTime?: string
  completionTime?: string
}

export interface K8sJob {
  apiVersion: 'batch/v1'
  kind: 'Job'
  metadata: K8sObjectMeta
  spec: K8sJobSpec
  status?: K8sJobStatus
}

export interface K8sCronJobSpec {
  schedule: string
  concurrencyPolicy?: string
  suspend?: boolean
  startingDeadlineSeconds?: number
  timeZone?: string
  successfulJobsHistoryLimit?: number
  failedJobsHistoryLimit?: number
  jobTemplate: { spec: K8sJobSpec }
}

export interface K8sCronJobStatus {
  active?: { name: string; namespace: string }[]
  lastScheduleTime?: string
  lastSuccessfulTime?: string
}

export interface K8sCronJob {
  apiVersion: 'batch/v1'
  kind: 'CronJob'
  metadata: K8sObjectMeta
  spec: K8sCronJobSpec
  status?: K8sCronJobStatus
}

// ============ 网络 ============

export interface K8sServicePort {
  name?: string
  protocol?: string
  port: number
  targetPort?: number | string
  nodePort?: number
  appProtocol?: string
}

export interface K8sServiceSpec {
  type?: string
  selector?: Record<string, string>
  ports: K8sServicePort[]
  clusterIP?: string
  clusterIPs?: string[]
  externalIPs?: string[]
  loadBalancerIP?: string
  sessionAffinity?: string
  externalTrafficPolicy?: string
}

export interface K8sServiceStatus {
  loadBalancer?: { ingress?: { ip?: string; hostname?: string }[] }
}

export interface K8sService {
  apiVersion: 'v1'
  kind: 'Service'
  metadata: K8sObjectMeta
  spec: K8sServiceSpec
  status?: K8sServiceStatus
}

export interface K8sIngressRule {
  host?: string
  http?: {
    paths: {
      path: string
      pathType: string
      backend: {
        service?: { name: string; port: { number?: number; name?: string } }
        resource?: Record<string, any>
      }
    }[]
  }
}

export interface K8sIngressTLS {
  hosts?: string[]
  secretName?: string
}

export interface K8sIngressSpec {
  ingressClassName?: string
  rules?: K8sIngressRule[]
  tls?: K8sIngressTLS[]
  defaultBackend?: Record<string, any>
}

export interface K8sIngress {
  apiVersion: 'networking.k8s.io/v1'
  kind: 'Ingress'
  metadata: K8sObjectMeta
  spec: K8sIngressSpec
  status?: { loadBalancer?: { ingress?: { ip?: string; hostname?: string }[] } }
}

// ============ 存储 ============

export interface K8sPersistentVolumeSpec {
  capacity?: { storage: string }
  accessModes?: string[]
  persistentVolumeReclaimPolicy?: string
  storageClassName?: string
  claimRef?: { namespace: string; name: string }
  nfs?: { server: string; path: string; readOnly?: boolean }
  hostPath?: { path: string; type?: string }
  csi?: { driver: string; volumeHandle: string; volumeAttributes?: Record<string, string> }
}

export interface K8sPersistentVolume {
  apiVersion: 'v1'
  kind: 'PersistentVolume'
  metadata: K8sObjectMeta
  spec: K8sPersistentVolumeSpec
  status?: { phase?: string; reason?: string }
}

export interface K8sPersistentVolumeClaimSpec {
  accessModes?: string[]
  resources: { requests: { storage: string } }
  storageClassName?: string
  volumeName?: string
}

export interface K8sPersistentVolumeClaim {
  apiVersion: 'v1'
  kind: 'PersistentVolumeClaim'
  metadata: K8sObjectMeta
  spec: K8sPersistentVolumeClaimSpec
  status?: { phase?: string; capacity?: { storage: string } }
}

export interface K8sStorageClass {
  apiVersion: 'storage.k8s.io/v1'
  kind: 'StorageClass'
  metadata: K8sObjectMeta
  provisioner: string
  parameters?: Record<string, string>
  reclaimPolicy?: string
  volumeBindingMode?: string
  allowVolumeExpansion?: boolean
}

// ============ 配置 ============

export interface K8sConfigMap {
  apiVersion: 'v1'
  kind: 'ConfigMap'
  metadata: K8sObjectMeta
  data?: Record<string, string>
  binaryData?: Record<string, string>
}

export interface K8sSecret {
  apiVersion: 'v1'
  kind: 'Secret'
  metadata: K8sObjectMeta
  data?: Record<string, string>
  type?: string
  stringData?: Record<string, string>
}

export interface K8sResourceQuota {
  apiVersion: 'v1'
  kind: 'ResourceQuota'
  metadata: K8sObjectMeta
  spec: { hard?: Record<string, string>; scopes?: string[] }
  status?: { hard?: Record<string, string>; used?: Record<string, string> }
}

export interface K8sLimitRange {
  apiVersion: 'v1'
  kind: 'LimitRange'
  metadata: K8sObjectMeta
  spec: {
    limits: {
      type: string
      max?: Record<string, string>
      min?: Record<string, string>
      default?: Record<string, string>
      defaultRequest?: Record<string, string>
      maxLimitRequestRatio?: Record<string, string>
    }[]
  }
}

// ============ 其他 ============

export interface K8sHPA {
  apiVersion: 'autoscaling/v2'
  kind: 'HorizontalPodAutoscaler'
  metadata: K8sObjectMeta
  spec: {
    scaleTargetRef: { apiVersion: string; kind: string; name: string }
    minReplicas?: number
    maxReplicas: number
    metrics?: any[]
    behavior?: Record<string, any>
  }
  status?: {
    currentReplicas: number
    desiredReplicas: number
    conditions?: K8sCondition[]
    lastScaleTime?: string
  }
}

export interface K8sNamespace {
  apiVersion: 'v1'
  kind: 'Namespace'
  metadata: K8sObjectMeta
  spec?: { finalizers?: string[] }
  status?: { phase?: string }
}

export interface K8sNodeCondition {
  type: string
  status: string
  lastHeartbeatTime?: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

export interface K8sNodeAddress {
  type: string
  address: string
}

export interface K8sNodeInfo {
  kubeletVersion: string
  osImage: string
  kernelVersion: string
  containerRuntimeVersion: string
  architecture?: string
}

export interface K8sNode {
  apiVersion: 'v1'
  kind: 'Node'
  metadata: K8sObjectMeta
  spec: {
    podCIDR?: string
    taints?: K8sToleration[]
    unschedulable?: boolean
  }
  status: {
    conditions?: K8sNodeCondition[]
    addresses?: K8sNodeAddress[]
    capacity?: Record<string, string>
    allocatable?: Record<string, string>
    nodeInfo?: K8sNodeInfo
  }
}

export interface K8sEvent {
  apiVersion: 'v1'
  kind: 'Event'
  metadata: K8sObjectMeta
  involvedObject: { kind: string; namespace?: string; name: string; uid?: string }
  reason?: string
  message?: string
  type?: string
  count?: number
  firstTimestamp?: string
  lastTimestamp?: string
  lastSeen?: string
  source?: { component?: string; host?: string }
}

// ============ 列表响应 ============

export interface K8sListResponse<T> {
  apiVersion: string
  kind: string
  metadata: { resourceVersion?: string; continue?: string }
  items: T[]
}
