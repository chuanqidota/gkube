#### 地址
```
https://github.com/kubernetes/client-go
```

#### 查看资源分组
```
kubectl api-resources
```

#### 安装client-go
```
go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest
go mod tidy
```

#### 对照如何操作资源
- k8s.io/client-go/kubernetes/clientset.go中 NewForConfig作为入口
- kubectl api-resources 中对应的资源分组就是Clientset中的属性