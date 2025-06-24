# svc-controller

## 项目简介

`svc-controller` 是一个用于 Kubernetes 集群中服务主备切换的控制器。它定期检测主实例（primary pod）的健康状况，当主实例不可用时自动将 Service 指向备用实例（standby pod），并通过企业微信机器人发送告警通知。

![image](https://github.com/user-attachments/assets/b981464e-a7f4-4fc6-862f-29a687676c85)


## 功能特性

- 监控指定命名空间下的主实例 Pod 健康状态
- 自动切换 Service 的 selector 到备用实例
- 切换及异常时通过企业微信 webhook 发送通知
- 支持通过 ConfigMap 挂载配置

## 目录结构

```
svc-controller/
├── build/
│   ├── Dockerfile
│   ├── controller-configmap.yaml
│   └── controller-Deployment.yaml
├── config/
│   └── config.json
├── internal/
│   ├── config.go
│   ├── main.go
│   └── wecom.go
├── go.mod
├── go.sum
└── .gitignore
```

## 使用方法

### 1. 构建镜像

```sh
docker build -t your-repo/svc-controller:latest -f build/Dockerfile .
```

### 2. 部署到 Kubernetes

- 配置 ConfigMap（见 `build/controller-configmap.yaml`）
- 配置 Deployment 及 RBAC（见 `build/controller-Deployment.yaml`）

```sh
kubectl apply -f build/controller-configmap.yaml
kubectl apply -f build/controller-Deployment.yaml
```

### 3. 配置说明

`config/config.json` 示例：

```json
{
    "namespace": "us",
    "serviceName": "vnnox-rabbitmq",
    "primaryLabelValue": "us-rabbitmq",
    "standbyLabelValue": "us-rabbitmq-standby",
    "labelKey": "app.kubernetes.io/instance",
    "wecomWebhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxxx"
}
```

- `namespace`：监控的命名空间
- `serviceName`：需要切换的 Service 名称
- `primaryLabelValue`：主实例 Pod 的 label 值
- `standbyLabelValue`：备用实例 Pod 的 label 值
- `labelKey`：用于区分主备的 label key
- `wecomWebhook`：企业微信机器人 webhook 地址

## 依赖

- Go 1.17+
- Kubernetes client-go v0.23.0

## 贡献

欢迎提交 issue 和 PR！
