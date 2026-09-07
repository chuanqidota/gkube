// Shared form type definitions for WorkloadForm, CronJobForm, JobForm

export interface Label { key: string; value: string }

export interface Port { name: string; containerPort: number | null; protocol: string }

export interface EnvVar {
  name: string; value: string; type: 'plain' | 'configMapKeyRef' | 'secretKeyRef' | 'fieldRef'
  configMapName: string; configMapKey: string
  secretName: string; secretKey: string
  fieldPath: string
}

export interface Resources { requests: { cpu: string; memory: string }; limits: { cpu: string; memory: string } }

export interface VolumeMount { name: string; mountPath: string; subPath: string; readOnly: boolean }

export interface Probe { type: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null; execCommand: string; initialDelaySeconds: number; periodSeconds: number; timeoutSeconds: number; failureThreshold: number }

export interface LifecycleHandler { type: 'exec' | 'httpGet' | 'tcpSocket'; execCommand: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null }

export interface Volume { name: string; type: string; hostPath: string; hostPathType: string; configMapName: string; secretName: string; pvcName: string }

export interface Tolerance { key: string; operator: string; value: string; effect: string; tolerationSeconds: number | null }

export interface Annotation { key: string; value: string }

export interface VolumeClaimTemplate {
  name: string; storageSize: string; storageClassName: string; accessModes: string[]
}

export interface AffinityRule { weight: number; topologyKey: string; namespaces: string; labelKey: string; labelValue: string }

export interface TopologySpreadConstraint { maxSkew: number; topologyKey: string; whenUnsatisfiable: string; labelKey: string; labelValue: string }

export interface Container {
  name: string; image: string; imagePullPolicy: string
  ports: Port[]; env: EnvVar[]; resources: Resources
  volumeMounts: VolumeMount[]; livenessProbe: Probe | null; readinessProbe: Probe | null; startupProbe: Probe | null
  command: string; args: string
  lifecycle: { preStop: LifecycleHandler | null; postStart: LifecycleHandler | null }
  securityContext: { runAsUser: number | null; runAsNonRoot: boolean; readOnlyRootFilesystem: boolean; privileged: boolean; capabilitiesAdd: string[]; capabilitiesDrop: string[] }
}

// ============ Base form data (common across all workload forms) ============

export interface BaseFormData {
  name: string
  namespace: string
  labels: Label[]
  containers: Container[]
  initContainers: Container[]
  volumes: Volume[]
  nodeSelector: Label[]
  tolerations: Tolerance[]
  annotations: Annotation[]
  serviceAccountName: string
  terminationGracePeriodSeconds: number | null
  imagePullSecrets: string[]
  podAffinityRules: AffinityRule[]
  podAntiAffinityRules: AffinityRule[]
  topologySpreadConstraints: TopologySpreadConstraint[]
}

// ============ Workload-specific form data ============

export interface WorkloadFormData extends BaseFormData {
  replicas: number
  strategyType: string
  maxSurge: string
  maxUnavailable: string
  serviceName: string
  updateStrategy: string
  dsUpdateStrategy: string
  podManagementPolicy: string
  volumeClaimTemplates: VolumeClaimTemplate[]
  dnsPolicy: string
  hostNetwork: boolean
  priorityClassName: string
}

export interface CronJobFormData extends BaseFormData {
  schedule: string
  concurrencyPolicy: string
  suspend: boolean
  startingDeadlineSeconds: number | null
  timeZone: string
  successfulJobsHistoryLimit: number | null
  failedJobsHistoryLimit: number | null
  completions: number | null
  parallelism: number | null
  backoffLimit: number | null
  restartPolicy: string
}

export interface JobFormData extends BaseFormData {
  completions: number | null
  parallelism: number | null
  backoffLimit: number | null
  activeDeadlineSeconds: number | null
  ttlSecondsAfterFinished: number | null
  completionMode: string
  restartPolicy: string
}

// ============ Factory functions ============

export function createEmptyLifecycleHandler(): LifecycleHandler {
  return { type: 'exec', execCommand: '', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null }
}

export function createEmptyEnv(): EnvVar {
  return { name: '', value: '', type: 'plain', configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: '' }
}

export function createEmptyProbe(): Probe {
  return { type: 'httpGet', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null, execCommand: '', initialDelaySeconds: 15, periodSeconds: 10, timeoutSeconds: 5, failureThreshold: 3 }
}

export function createEmptyContainer(): Container {
  return {
    name: '', image: '', imagePullPolicy: 'IfNotPresent',
    ports: [], env: [],
    resources: { requests: { cpu: '', memory: '' }, limits: { cpu: '', memory: '' } },
    volumeMounts: [], livenessProbe: null, readinessProbe: null, startupProbe: null,
    command: '', args: '',
    lifecycle: { preStop: null, postStart: null },
    securityContext: { runAsUser: null, runAsNonRoot: false, readOnlyRootFilesystem: false, privileged: false, capabilitiesAdd: [], capabilitiesDrop: [] },
  }
}
