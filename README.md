# mdgen

本项目提供一个在 macOS 本地运行的 Markdown 生成服务，支持命令行前台模式与通过 launchd 的后台常驻模式。服务在本机 `127.0.0.1:27123` 监听 HTTP，接收页面参数生成 `.md` 文件写入指定目录。

## 架构与设计
- **HTTP 框架**：使用标准库 `net/http`，路由由 `http.ServeMux` 提供，便于在受限环境下零依赖运行。
- **端口与安全**：默认仅监听 `127.0.0.1:27123`；可选 `X-API-Key` 头鉴权，未配置时关闭。
- **配置优先级**：`CLI flags` > 环境变量（前缀 `MDGEN_`）> 配置文件（`~/.config/mdgen/config.yaml` 或 `config.json`）。
- **Markdown 模板**：默认 front-matter（title、created_at）+ 一级标题 + 原文内容。
- **文件写入策略**：先写入 `.tmp` 再 `rename`，并带唯一命名（存在冲突自动追加 `-1` 递增）。
- **日志**：使用 `slog` JSON 日志，记录 request_id、路径、耗时与状态码。

## 运行方式
### 前台运行（调试）
```bash
# 拉取依赖（无外部依赖）
go run ./cmd/mdgen
```

可用 flags 覆盖配置：
```bash
go run ./cmd/mdgen --listen 127.0.0.1:28000 --output ~/Documents/notes --api-key secret
```

### 配置文件示例
参见 [examples/config.yaml](examples/config.yaml)。环境变量前缀为 `MDGEN_`，如 `MDGEN_OUTPUT_DIR`。

### HTTP API
- `GET /healthz` → 健康检查
- `GET /version` → 版本信息
- `POST /api/note`，JSON 体：
```json
{
  "title": "页面标题",
  "content": "正文",
  "folder": "inbox"   // 可选，写入子目录
}
```
响应示例：
```json
{
  "ok": true,
  "filename": "20250101_page-title.md",
  "path": "/Users/me/Documents/notes/inbox/20250101_page-title.md",
  "status": "saved"
}
```

### launchd 安装
确保先构建可执行文件：
```bash
go build -o mdgen ./cmd/mdgen
```

运行安装脚本（会复制二进制到 `~/bin/mdgen`，配置到 `~/.config/mdgen/config.yaml`）：
```bash
bash scripts/install.sh
```
卸载：
```bash
bash scripts/uninstall.sh
```

### curl 示例
```bash
curl -X POST http://127.0.0.1:27123/api/note \
  -H 'Content-Type: application/json' \
  -d '{"title":"Test Page","content":"Hello from curl"}'
```
如开启 API Key：添加 `-H 'X-API-Key: <key>'`。

## 测试
```bash
go test ./...
```

## 产物清单
- `cmd/mdgen`：主程序入口
- `internal/`：配置、日志、服务器、笔记处理逻辑
- `examples/config.yaml`：配置示例
- `deploy/com.example.mdgen.plist`：launchd 模板
- `scripts/install.sh` 与 `scripts/uninstall.sh`
