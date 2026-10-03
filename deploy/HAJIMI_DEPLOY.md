# hajimi 自动部署

适用于 `kule-re/sub2api` 的 `dev` 分支、x86_64 服务器和已运行的 `/opt/sub2api` Docker Compose 安装。

## 流程

推送 dev → 后端单元测试 / WorkBuddy 助手测试 / 前端检查与关键测试 / 部署脚本测试 → 根 Dockerfile 构建完整 linux/amd64 镜像 → 发布到 `ghcr.io/kule-re/hajimi:sha-<完整提交号>` → SSH 更新服务器 → 数据库与应用数据备份 → 只更新 sub2api → 健康检查。

实际部署使用构建输出的 `ghcr.io/kule-re/hajimi@sha256:<摘要>`，不会因同名标签变化而使用另一版镜像。若 dev 已有更新提交，旧运行只保留镜像，跳过部署。GitHub 流程串行，服务器另有文件锁防止并发更新。

代码：`.github/workflows/hajimi-deploy.yml`、`deploy/hajimi-deploy.sh`。

## 一次性设置

### 1. 准备部署专用 SSH 密钥

在 Windows PowerShell 执行；如果同名密钥已存在，不覆盖它：

```powershell
ssh-keygen -t ed25519 -C hajimi-actions -f "$env:USERPROFILE\.ssh\hajimi_actions"
Get-Content "$env:USERPROFILE\.ssh\hajimi_actions.pub"
```

这把密钥用于无人值守，生成时不设置口令。把 `.pub` 公钥追加到服务器部署用户的 `~/.ssh/authorized_keys`；不要替换原有公钥。确认 `.ssh` 权限为 700、authorized_keys 为 600。当前流程默认以 root 连接，用户需要能操作 Docker 和读取部署目录。

在本机验证：

```powershell
ssh -i "$env:USERPROFILE\.ssh\hajimi_actions" -p 35169 root@45.205.6.70 "docker compose version"
```

### 2. 固定服务器身份

通过已确认身份的 SSH 会话或服务器控制台查看服务器指纹：

```bash
ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub
```

本机获取 host key 并查看指纹：

```powershell
ssh-keyscan -t ed25519 -p 35169 45.205.6.70 | Set-Content -Encoding ascii -LiteralPath "$env:TEMP\hajimi-known-hosts"
ssh-keygen -lf "$env:TEMP\hajimi-known-hosts"
```

两处指纹一致后，使用 hajimi-known-hosts 文件内容作为下面的 known_hosts Secret。非 22 端口的主机名格式是 `[45.205.6.70]:35169`。流程启用严格身份检查，不会自动信任扫描结果。

### 3. GitHub 设置

仓库 Settings → Environments 创建 `hajimi`。在这个环境添加 Secrets：

| Secret | 内容 |
| --- | --- |
| `HAJIMI_SSH_HOST` | `45.205.6.70` |
| `HAJIMI_SSH_PRIVATE_KEY` | 部署专用私钥 hajimi_actions 的完整内容，含头尾和换行 |
| `HAJIMI_SSH_KNOWN_HOSTS` | 上一步核对过的 known_hosts 文件内容 |

私钥只填 GitHub Secret，不发到聊天、仓库或日志。

在 Settings → Secrets and variables → Actions → Variables 设置 **Repository variables**：

| Variable | 内容 |
| --- | --- |
| `HAJIMI_DEPLOY_ENABLED` | `true` 才允许连接服务器；不设置时只测试、构建和发布镜像 |
| `HAJIMI_SSH_PORT` | `35169`；实际端口不同则修改 |
| `HAJIMI_SSH_USER` | `root`；实际部署用户不同则修改 |

仓库须允许 GitHub Actions 运行和写入 Packages。首次 GHCR 镜像由本仓库 workflow 创建，以源仓库 label 关联；如果同名包已经由别处创建，需要在包的 Manage Actions access 中授权本仓库。流程用 GitHub 自动签发的 GITHUB_TOKEN 推送和拉取，无需另存长期 GHCR token。

如果 hajimi 环境有 required reviewers，则部署仍会等待环境审批。设置 HAJIMI_DEPLOY_ENABLED 不会绕过审批规则。保持已有环境保护，按希望的发布方式配置。

### 4. 推送后查看结果

把本次自动化文件提交到 `dev` 并推送后，在仓库 Actions → Hajimi Deploy 查看运行。镜像构建和服务器部署均未在本地验证为成功；第一次 GitHub 运行是实际联调。

`workflow_dispatch` 的界面入口需要 workflow 文件存在于仓库默认分支；仅存在 dev 时用 push dev 触发，不需要合并 main 才能使用 push 自动化。

## 服务器上的更新行为

- 脚本要求现有 sub2api 容器、`.env` 和 Compose 标签存在，拒绝初始化新安装或切换到另一个部署目录。
- 自动读取现有 Compose 项目名和原配置文件，包括自定义文件名和其他覆盖文件。使用原 `.env`、项目目录、网络和数据路径。
- 先拉取镜像、验证候选配置，再在 `backups/hajimi-<时间>.<随机值>/` 保存数据库 dump、应用 `/app/data`、`.env`、Compose 配置和旧镜像标识。备份目录和文件使用私有权限。
- PostgreSQL 备份由正在运行的 postgres 容器执行，使用其真实用户、数据库和客户端；验证命令成功、非空及 archive 目录可读取。
- 应用数据从运行中的容器读取，兼容绑定目录和命名卷。数据库和应用数据是分别备份的，不声称是跨组件原子快照。
- 仅生成 `/opt/sub2api/compose.hajimi.yml` 的应用 image/pull_policy 配置。用 `up --no-deps --no-build --pull never --wait sub2api` 更新应用，等待最多 180 秒，检查运行镜像 ID。
- GHCR 凭据从 stdin 输入，只保存在一次性 Docker 配置目录，退出时移除；SSH 临时私钥和远端脚本也在结束时清理。
- 不执行 down、数据库/Redis 镜像更新、数据删除或后台日志上传。容器替换期间会有短暂服务中断，健康检查不证明上游模型或实际用户任务成功。
- 备份自动保留，不自动删除；需按实际磁盘空间制定备份保留策略。

之后手工 Compose 操作也应包含 `compose.hajimi.yml`，避免启动回原配置中的官方镜像。用 `docker inspect sub2api --format '{{index .Config.Labels "com.docker.compose.project.config_files"}}'` 查看该容器的完整配置文件列表。

## 失败与恢复

拉镜像、配置检查或备份失败会停止，不重建应用。健康检查失败或运行镜像不匹配时报告失败，保留新配置、备份和旧镜像，等待人工检查；不盲目恢复数据库或切回旧镜像，因为应用启动可能已经执行迁移。

先在服务器查看 `docker logs --tail=100 sub2api` 和 Compose 状态。备份中 old-image-tag.txt 指向保留的旧镜像，数据库恢复或应用回退应根据迁移兼容性单独决定。关闭 Repository variable HAJIMI_DEPLOY_ENABLED 即可阻止下一次服务器部署。

## 本地验证

```bash
bash -n deploy/hajimi-deploy.sh
python3 -B -m unittest discover -s deploy/tests -p 'test_hajimi_deploy.py'
```

故障注入测试只使用临时目录和模拟 Docker，不连接生产服务器、GHCR 或真实数据库。在 Windows Git Bash 中模拟 flock；Ubuntu 使用系统 flock。真实 Docker 构建、SSH、GHCR 权限和迁移仍需首次运行验证。

官方参考：[GitHub 发布容器镜像](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images)、[GHCR](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)、[Docker Compose up](https://docs.docker.com/reference/cli/docker/compose/up/)。
