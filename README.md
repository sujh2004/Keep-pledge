# KeepPledge 守约挑战

基于 `documents/Development.md` 搭建的课程项目 MVP。项目采用前后端分离结构：

- `frontend/`: React 18 + TypeScript + Vite + Ant Design + Zustand
- `backend/`: Go + Gin + GORM + JWT，按 Handler -> Service -> Repository 分层
- `deploy/`: Docker Compose、Nginx 与前后端镜像配置

## 本地开发

环境要求：

- Go 1.22+
- Node.js 18+
- MySQL 8.0
- Redis 7

前端：

```bash
cd frontend
npm install
npm run dev
```

后端：

```bash
cd backend
go mod tidy
go run ./cmd/server
```

后端默认读取 `backend/config.yaml`，MySQL 与 Redis 可用 `deploy/docker-compose.yml` 启动。

## 测试与验证

后端单元测试：

```bash
cd backend
go test ./...
```

前端构建检查：

```bash
cd frontend
npm run build
```

Docker 配置检查：

```bash
cd deploy
docker compose config
```

## Docker 一键启动

```bash
cd deploy
docker compose up --build
```

启动后访问：

- 前端：`http://localhost:3000`
- 后端 API：`http://localhost:8080/api/v1`

## 课程展示流程

1. 注册两个用户，分别记录个人主页中的用户 ID。
2. 用户 A 创建公开挑战，填写誓约和失败后果。
3. 用户 B 在挑战广场加入挑战。
4. 用户 A 发起打卡，查看连击、XP、热力图变化。
5. 用户 A 向用户 B 发送好友申请，用户 B 接受。
6. 将挑战状态改为 completed，查看证书记录和证书图片。
7. 查看排行榜、通知中心、成就墙和个人中心统计。
