// Shared add/remove array functions for workload forms
import { createEmptyContainer, createEmptyEnv, createEmptyLifecycleHandler } from './form-types'
import type { Container } from './form-types'

/**
 * Composable that provides add/remove functions for form arrays.
 * Takes a reactive form object with containers, labels, volumes, etc.
 */
export function useFormArrays(formData: {
  labels: { key: string; value: string }[]
  annotations: { key: string; value: string }[]
  containers: Container[]
  initContainers: Container[]
  volumes: { name: string; type: string; hostPath: string; hostPathType: string; configMapName: string; secretName: string; pvcName: string }[]
  imagePullSecrets: string[]
  volumeClaimTemplates?: { name: string; storageSize: string; storageClassName: string; accessModes: string[] }[]
}) {
  // Labels
  const addLabel = () => formData.labels.push({ key: '', value: '' })
  const removeLabel = (i: number) => formData.labels.splice(i, 1)

  // Annotations
  const addAnnotation = () => formData.annotations.push({ key: '', value: '' })
  const removeAnnotation = (i: number) => formData.annotations.splice(i, 1)

  // Containers
  const addContainer = () => formData.containers.push(createEmptyContainer())
  const removeContainer = (i: number) => { if (formData.containers.length > 1) formData.containers.splice(i, 1) }
  const addInitContainer = () => formData.initContainers.push(createEmptyContainer())
  const removeInitContainer = (i: number) => formData.initContainers.splice(i, 1)

  // Ports (per container)
  const addPort = (ci: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).ports.push({ name: '', containerPort: null, protocol: 'TCP' })
  }
  const removePort = (ci: number, pi: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).ports.splice(pi, 1)
  }

  // Env vars (per container)
  const addEnv = (ci: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).env.push(createEmptyEnv())
  }
  const removeEnv = (ci: number, ei: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).env.splice(ei, 1)
  }

  // Volumes
  const addVolume = () => formData.volumes.push({ name: '', type: 'emptyDir', hostPath: '', hostPathType: 'DirectoryOrCreate', configMapName: '', secretName: '', pvcName: '' })
  const removeVolume = (i: number) => formData.volumes.splice(i, 1)

  // Volume mounts (per container)
  const addVolumeMount = (ci: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).volumeMounts.push({ name: '', mountPath: '', subPath: '', readOnly: false })
  }
  const removeVolumeMount = (ci: number, mi: number, isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).volumeMounts.splice(mi, 1)
  }

  // Lifecycle (per container)
  const enableLifecycle = (ci: number, hookType: 'preStop' | 'postStart', isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).lifecycle[hookType] = createEmptyLifecycleHandler()
  }
  const disableLifecycle = (ci: number, hookType: 'preStop' | 'postStart', isInit?: boolean) => {
    (isInit ? formData.initContainers[ci] : formData.containers[ci]).lifecycle[hookType] = null
  }

  // Image pull secrets
  const addImagePullSecret = () => formData.imagePullSecrets.push('')
  const removeImagePullSecret = (i: number) => formData.imagePullSecrets.splice(i, 1)

  // Volume claim templates (StatefulSet only)
  const addVolumeClaimTemplate = () => formData.volumeClaimTemplates?.push({ name: '', storageSize: '1Gi', storageClassName: '', accessModes: ['ReadWriteOnce'] })
  const removeVolumeClaimTemplate = (i: number) => formData.volumeClaimTemplates?.splice(i, 1)

  // Security context capabilities (per container)
  const addCapability = (sc: Container['securityContext'], type: 'add' | 'drop') => {
    (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).push('')
  }
  const removeCapability = (sc: Container['securityContext'], type: 'add' | 'drop', i: number) => {
    (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).splice(i, 1)
  }

  return {
    addLabel, removeLabel,
    addAnnotation, removeAnnotation,
    addContainer, removeContainer,
    addInitContainer, removeInitContainer,
    addPort, removePort,
    addEnv, removeEnv,
    addVolume, removeVolume,
    addVolumeMount, removeVolumeMount,
    enableLifecycle, disableLifecycle,
    addImagePullSecret, removeImagePullSecret,
    addVolumeClaimTemplate, removeVolumeClaimTemplate,
    addCapability, removeCapability,
  }
}
