import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    titleKey?: string
    title?: string
    parent?: string
    requireAdmin?: boolean
    public?: boolean
    icon?: string
  }
}
