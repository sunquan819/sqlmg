# SQLMG

A lightweight, single-binary database management tool built with Go and Svelte 5. Inspired by Navicat, designed for simplicity.

**English** | [中文](#中文)

## Features

- **Multi-Database Support**: MySQL, PostgreSQL, SQLite
- **SQL Editor**: CodeMirror 6 with syntax highlighting and auto-completion
- **Table Designer**: Visual column, index, and foreign key management
- **DDL Preview/Generate**: Driver-aware SQL generation (e.g., `AUTO_INCREMENT` for MySQL, `AUTOINCREMENT` for SQLite)
- **Data Import/Export**: CSV, JSON, SQL formats with streaming support for large datasets
- **ER Diagram**: Interactive visualization with automatic layout using elkjs
- **Bilingual UI**: Chinese (default) and English, one-click switch
- **Single Binary**: Embedded frontend, zero external dependencies

## Screenshot

![SQLMG Screenshot](docs/screenshot.png)

## Installation

### Pre-built Binaries

Download from [Releases](https://github.com/sunquan819/sqlmg/releases).

### Build from Source

Requirements:
- Go 1.25+
- Node.js 18+

```bash
# Clone repository
git clone https://github.com/sunquan819/sqlmg.git
cd sqlmg

# Build frontend
cd web
npm install
npm run build
cd ..

# Build backend (embeds frontend)
go build -o sqlmg ./cmd/sqlmg
```

## Usage

```bash
# Start server (default: http://127.0.0.1:9180)
./sqlmg

# Development mode (serves from web/dist without embedding)
./sqlmg -dev

# Custom address and port
./sqlmg -addr 0.0.0.0 -port 8080
```

Open `http://127.0.0.1:9180` in your browser.

### Configuration

Data is stored in:
- **Connections & Settings**: `~/.sqlmg/sqlmg.db` (SQLite metastore)
- **Encrypted Passwords**: AES-GCM encryption

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + Enter` | Execute SQL |
| `Ctrl/Cmd + Shift + Enter` | Execute selected SQL |
| `Ctrl/Cmd + S` | Save table design |

## Development

### Project Structure

```
sqlmg/
├── cmd/sqlmg/           # Entry point
├── internal/
│   ├── app/             # Application bootstrap
│   ├── config/          # Configuration
│   ├── connection/      # Connection manager
│   ├── dbengine/        # Database driver interface
│   │   ├── mysql/       # MySQL driver
│   │   ├── postgres/    # PostgreSQL driver
│   │   └── sqlite/      # SQLite driver
│   ├── export/          # Export engine (CSV/JSON/SQL)
│   ├── handler/         # HTTP handlers
│   ├── import/          # Import engine (CSV/JSON/SQL)
│   ├── metastore/       # Embedded SQLite metastore
│   ├── security/        # AES-GCM encryption
│   └── server/          # Gin server setup
├── web/                 # Svelte 5 frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/  # UI components
│   │   │   ├── i18n/       # Translations
│   │   │   └── stores/     # State management
│   │   └── routes/         # SvelteKit routes
│   └── package.json
└── README.md
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/connections` | List connections |
| `POST` | `/api/connections` | Create connection |
| `PUT` | `/api/connections/:id` | Update connection |
| `DELETE` | `/api/connections/:id` | Delete connection |
| `POST` | `/api/connections/:id/test` | Test connection |
| `GET` | `/api/connections/:id/databases` | List databases |
| `GET` | `/api/connections/:id/schemas/:schema/tables` | List tables |
| `GET` | `/api/connections/:id/schemas/:schema/tables/:table/columns` | Get columns |
| `POST` | `/api/connections/:id/query` | Execute SQL |
| `GET` | `/api/connections/:id/schemas/:schema/ddl/:table` | Get DDL |
| `POST` | `/api/connections/:id/schemas/:schema/tables` | Create table |
| `PUT` | `/api/connections/:id/schemas/:schema/tables/:table` | Alter table |
| `DELETE` | `/api/connections/:id/schemas/:schema/tables/:table` | Drop table |
| `GET` | `/api/connections/:id/schemas/:schema/er` | Get ER diagram |
| `POST` | `/api/connections/:id/export` | Export data |
| `POST` | `/api/connections/:id/import` | Import data |
| `GET` | `/api/history` | Get query history |
| `DELETE` | `/api/history` | Clear query history |

### Tech Stack

**Backend:**
- Go 1.25
- Gin (HTTP server)
- go-sql-driver/mysql
- jackc/pgx/v5 (PostgreSQL)
- modernc.org/sqlite (Pure Go SQLite)
- filippo.io/edwards25519 (Encryption)

**Frontend:**
- Svelte 5 (Runes)
- SvelteKit
- CodeMirror 6
- AG Grid Community
- elkjs (ER diagram layout)

## License

MIT License. See [LICENSE](LICENSE).

---

## 中文

一个轻量级、单文件的数据库管理工具，使用 Go 和 Svelte 5 构建。灵感来自 Navicat，追求简洁。

### 功能特性

- **多数据库支持**: MySQL、PostgreSQL、SQLite
- **SQL 编辑器**: CodeMirror 6 语法高亮和自动补全
- **表设计器**: 可视化管理列、索引和外键
- **DDL 预览/生成**: 感知驱动的 SQL 生成（如 MySQL 用 `AUTO_INCREMENT`，SQLite 用 `AUTOINCREMENT`）
- **数据导入/导出**: CSV、JSON、SQL 格式，支持大数据流式导出
- **ER 图**: 交互式可视化，使用 elkjs 自动布局
- **双语界面**: 默认中文，一键切换英文
- **单文件部署**: 前端嵌入后端，零外部依赖

### 安装

从 [Releases](https://github.com/sunquan819/sqlmg/releases) 下载预编译二进制文件。

或从源码构建：

```bash
git clone https://github.com/sunquan819/sqlmg.git
cd sqlmg

cd web && npm install && npm run build && cd ..
go build -o sqlmg ./cmd/sqlmg
```

### 使用方法

```bash
# 启动服务器 (默认: http://127.0.0.1:9180)
./sqlmg

# 开发模式
./sqlmg -dev

# 自定义地址和端口
./sqlmg -addr 0.0.0.0 -port 8080
```

### 数据存储

- **连接和设置**: `~/.sqlmg/sqlmg.db` (SQLite 元数据库)
- **加密密码**: AES-GCM 加密存储

### 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl/Cmd + Enter` | 执行 SQL |
| `Ctrl/Cmd + Shift + Enter` | 执行选中 SQL |
| `Ctrl/Cmd + S` | 保存表设计 |

### 开发

项目结构和技术栈请参考英文部分。

### 许可证

MIT 许可证。详见 [LICENSE](LICENSE)。