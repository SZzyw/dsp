# 贡献指南

语言 / Language: [中文](CONTRIBUTING.md) | [English](CONTRIBUTING.en.md)

感谢你对 DS2API 的贡献。

## 开发环境设置

### 后端（Go）

```bash
# 1. 克隆仓库
git clone https://github.com/CJackHwang/ds2api.git
cd ds2api

# 2. 配置
cp config.example.json config.json
# 编辑 config.json

# 3. 启动后端
go run ./cmd/ds2api
```

### 前端（WebUI）

```bash
cd webui
npm install
npm run dev
```

WebUI 语言包位于 `webui/src/locales/`。

## 代码规范

- **Go**: 提交前运行 `gofmt`，并确保 `go test ./...` 通过
- **JavaScript/React**: 保持现有代码风格（函数组件）
- **提交信息**: 使用语义化前缀（`feat:`, `fix:`, `docs:`）

## 提交 PR

1. Fork 仓库
2. 创建分支（如 `feature/xxx`）
3. 提交更改
4. 推送分支
5. 发起 Pull Request

## WebUI 构建

```bash
./scripts/build-webui.sh
```

## 项目结构

```text
ds2api/
├── cmd/ds2api/         # 本地/容器启动入口
├── api/index.go        # Vercel Serverless 入口
├── internal/           # Go 后端核心实现
├── webui/              # React WebUI 源码
├── static/admin/       # WebUI 构建产物
├── Dockerfile
├── docker-compose.yml
└── vercel.json
```

## 问题反馈

请使用 [GitHub Issues](https://github.com/CJackHwang/ds2api/issues) 并附复现步骤与日志。
