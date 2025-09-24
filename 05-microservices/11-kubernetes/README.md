# ☸️ Kubernetes微服务部署指南

本模块展示了Go微服务在Kubernetes上的完整部署方案，包含生产级配置和云原生最佳实践。

## 🎯 核心特性

### 🚀 生产就绪部署
- ✅ 多副本高可用部署
- ✅ 滚动更新和回滚
- ✅ 健康检查和就绪探针
- ✅ 资源限制和配额管理
- ✅ Pod反亲和性和容忍度

### 🔐 安全最佳实践
- ✅ 非特权用户运行
- ✅ 只读根文件系统
- ✅ RBAC权限管理
- ✅ 网络策略隔离
- ✅ Pod安全策略

### 📊 可观测性
- ✅ Prometheus指标收集
- ✅ Grafana可视化面板
- ✅ 自定义告警规则
- ✅ 分布式追踪(Jaeger)
- ✅ 结构化日志收集

### 🔀 流量管理
- ✅ Ingress流量路由
- ✅ Service Mesh支持
- ✅ 负载均衡策略
- ✅ TLS证书管理
- ✅ 熔断和重试

### 📈 自动扩缩容
- ✅ 水平Pod自动扩缩容(HPA)
- ✅ 基于CPU和内存的扩缩容
- ✅ Pod中断预算(PDB)
- ✅ 集群自动扩缩容支持

## 📁 文件结构

```
05-microservices/11-kubernetes/
├── deployment.yaml     # 核心部署配置
├── ingress.yaml       # 流量路由配置
├── monitoring.yaml    # 监控告警配置
├── kustomization.yaml # Kustomize配置
├── README.md          # 使用说明
└── scripts/           # 部署脚本
    ├── deploy.sh      # 一键部署
    ├── cleanup.sh     # 清理脚本
    └── update.sh      # 滚动更新
```

## 🚀 快速开始

### 1. 前置条件

```bash
# 检查Kubernetes集群
kubectl cluster-info

# 检查必要的CRD
kubectl get crd | grep -E "(servicemonitor|prometheusrule)"

# 创建命名空间
kubectl create namespace microservices
kubectl create namespace monitoring
```

### 2. 一键部署

```bash
# 应用所有配置
kubectl apply -f deployment.yaml
kubectl apply -f ingress.yaml
kubectl apply -f monitoring.yaml

# 等待部署完成
kubectl rollout status deployment/go-microservice -n microservices
```

### 3. 验证部署

```bash
# 检查Pod状态
kubectl get pods -n microservices -l app=go-microservice

# 检查服务
kubectl get svc -n microservices

# 检查Ingress
kubectl get ingress -n microservices
```

## 🔧 配置管理

### ConfigMap配置
```bash
# 查看配置
kubectl get configmap microservice-config -n microservices -o yaml

# 更新配置
kubectl patch configmap microservice-config -n microservices --patch '
data:
  log_level: "debug"
'

# 重启Pod应用新配置
kubectl rollout restart deployment/go-microservice -n microservices
```

### Secret管理
```bash
# 创建Secret
kubectl create secret generic microservice-secrets \
  --from-literal=postgres_url="postgres://user:pass@host:5432/db" \
  --from-literal=redis_url="redis://redis:6379/0" \
  -n microservices

# 更新Secret
kubectl patch secret microservice-secrets -n microservices --patch '
data:
  api_key: '$(echo -n "new_api_key" | base64)
```

## 📊 监控告警

### Prometheus指标
```bash
# 检查ServiceMonitor
kubectl get servicemonitor -n microservices

# 查看指标端点
kubectl port-forward svc/go-microservice-svc -n microservices 8080:8080
curl http://localhost:8080/api/metrics
```

### Grafana面板
```bash
# 导入Dashboard
kubectl apply -f monitoring.yaml

# 访问Grafana
kubectl port-forward svc/grafana -n monitoring 3000:80
# 访问: http://localhost:3000
```

### 告警规则
```bash
# 检查PrometheusRule
kubectl get prometheusrule -n microservices

# 查看告警状态
kubectl port-forward svc/prometheus -n monitoring 9090:9090
# 访问: http://localhost:9090/alerts
```

## 🌐 流量路由

### Ingress配置
```bash
# 检查Ingress状态
kubectl get ingress go-microservice-ingress -n microservices

# 获取外部IP
kubectl get ingress go-microservice-ingress -n microservices \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# 测试访问
curl https://api.example.com/v1/microservice/health
```

### Service Mesh (Istio)
```bash
# 启用Istio注入
kubectl label namespace microservices istio-injection=enabled

# 重启Pod应用Sidecar
kubectl rollout restart deployment/go-microservice -n microservices

# 检查Sidecar状态
kubectl get pods -n microservices -o jsonpath='{.items[*].spec.containers[*].name}'
```

## 📈 扩缩容管理

### 手动扩缩容
```bash
# 扩容到5个副本
kubectl scale deployment go-microservice -n microservices --replicas=5

# 查看扩容状态
kubectl get deployment go-microservice -n microservices -w
```

### 自动扩缩容
```bash
# 检查HPA状态
kubectl get hpa go-microservice-hpa -n microservices

# 查看扩缩容事件
kubectl describe hpa go-microservice-hpa -n microservices

# 触发负载测试
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- \
  /bin/sh -c "while sleep 0.01; do wget -q -O- http://go-microservice-svc.microservices.svc.cluster.local/; done"
```

## 🔄 滚动更新

### 更新镜像版本
```bash
# 更新镜像
kubectl set image deployment/go-microservice go-microservice=microservice:v1.1.0 -n microservices

# 查看滚动更新状态
kubectl rollout status deployment/go-microservice -n microservices

# 查看更新历史
kubectl rollout history deployment/go-microservice -n microservices
```

### 回滚部署
```bash
# 回滚到上一个版本
kubectl rollout undo deployment/go-microservice -n microservices

# 回滚到特定版本
kubectl rollout undo deployment/go-microservice -n microservices --to-revision=2
```

## 🐛 故障排除

### 常见问题诊断

**1. Pod启动失败**
```bash
# 查看Pod状态
kubectl get pods -n microservices -l app=go-microservice

# 查看Pod详细信息
kubectl describe pod <pod-name> -n microservices

# 查看Pod日志
kubectl logs <pod-name> -n microservices
```

**2. 服务不可访问**
```bash
# 检查Service端点
kubectl get endpoints go-microservice-svc -n microservices

# 测试Pod内部连接
kubectl exec -it <pod-name> -n microservices -- wget -qO- http://localhost:8080/health

# 检查网络策略
kubectl get networkpolicy -n microservices
```

**3. 健康检查失败**
```bash
# 查看探针配置
kubectl get deployment go-microservice -n microservices -o yaml | grep -A 10 livenessProbe

# 手动测试健康检查
kubectl port-forward <pod-name> -n microservices 8080:8080
curl http://localhost:8080/health
```

### 日志收集
```bash
# 查看所有相关日志
kubectl logs -l app=go-microservice -n microservices --tail=100

# 实时监控日志
kubectl logs -f deployment/go-microservice -n microservices

# 查看事件
kubectl get events -n microservices --sort-by='.lastTimestamp'
```

### 性能调试
```bash
# 检查资源使用
kubectl top pods -n microservices -l app=go-microservice

# 查看HPA指标
kubectl get hpa go-microservice-hpa -n microservices -o yaml

# 进入Pod调试
kubectl exec -it <pod-name> -n microservices -- sh
```

## 🔒 安全配置

### RBAC权限
```bash
# 检查ServiceAccount
kubectl get serviceaccount microservice-sa -n microservices

# 检查Role绑定
kubectl describe rolebinding microservice-rolebinding -n microservices

# 测试权限
kubectl auth can-i get pods --as=system:serviceaccount:microservices:microservice-sa -n microservices
```

### 网络策略
```bash
# 检查网络策略
kubectl get networkpolicy -n microservices

# 测试网络连接
kubectl run netshoot --rm -i --tty --image nicolaka/netshoot -- /bin/bash
```

### Pod安全
```bash
# 检查安全上下文
kubectl get deployment go-microservice -n microservices -o yaml | grep -A 20 securityContext

# 验证非root用户
kubectl exec <pod-name> -n microservices -- id
```

## 🎯 最佳实践

### ✅ 推荐做法
- 使用命名空间隔离不同环境
- 配置适当的资源限制和请求
- 实现完整的健康检查
- 使用滚动更新策略
- 配置Pod反亲和性
- 实现监控和告警
- 使用非特权用户运行

### ❌ 避免做法
- 在容器中运行多个进程
- 使用latest标签
- 忽略资源限制
- 禁用健康检查
- 硬编码配置信息
- 忽略安全配置
- 单副本部署生产服务

## 🔗 相关资源

- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [Prometheus Operator](https://prometheus-operator.dev/)
- [Istio Service Mesh](https://istio.io/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager](https://cert-manager.io/)

---

**🎉 恭喜！** 您已掌握Kubernetes微服务部署的完整方案！