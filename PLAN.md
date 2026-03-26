# 多集群管理服务
集群注册/纳管（支持kubeconfig和token上传、自动探测）。
集群状态监控（心跳、版本、节点数）。
多集群资源聚合查询。

# K8s工作负载管理服务
负载列表、详情、YAML编辑。
扩缩容、重启、镜像更新、回滚。
事件查看、容器日志（实时流）。
支持Deployment、StatefulSet、DaemonSet、Job/CronJob。

# 核心资源管理服务
Namespace、Node、PV/PVC管理。
Service、Ingress管理。
ConfigMap、Secret管理（加密存储敏感数据）。

# 用户管理服务
用户、团队、角色CRUD。
RBAC权限模型设计（平台级、集群级、命名空间级）。
对接LDAP/OIDC。

# 审计日志服务
通过API网关拦截或AOP记录操作日志。
审计日志存储到PostgreSQL或Elasticsearch。
审计日志查询界面。

# 告警中心
集成Prometheus Alertmanager API。
告警规则管理（创建、编辑、启用/停用）。
告警静默、抑制配置。
通知渠道管理（邮件、钉钉、Slack等）。
告警历史查询、图表展示。

# 日志管理
集成Elasticsearch或Loki。
日志检索界面（支持时间范围、关键词、标签过滤）。

# 监控配置
集成Prometheus，consul 直接使用promsql 进行读取时序数据库的数据
自定义监控指标配置。

# Harbor镜像仓库管理
集成Harbor API，管理项目、成员、镜像、复制规则。
镜像漏洞扫描结果展示。
镜像同步/复制配置。

# CI/CD流水线
集成Jenkins。
流水线模板定义、可视化编排。
流水线触发（手动、Git事件）、状态查看、日志查看。
对接Git仓库（GitHub、GitLab）。

# 业务应用管理
应用模板管理（Helm Chart或自定义模板）。
应用部署（支持多集群、多环境）。
应用版本管理、回滚。
应用拓扑图展示。