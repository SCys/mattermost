# Mattermost Rich Slash Commands Specification (Discord/Slack Level)

> **Document Version:** 1.0.0  
> **Target Branch:** `feat/rich-slash-commands`  
> **Status:** Draft / Active Implementation  
> **Author:** Antigravity Engineering  

---

## 1. 概述与设计背景 (Overview & Motivation)

Mattermost 原生 Slash Command 设计继承自早期的 IRC 与经典 Slack 模式：
- 依靠单行纯文本输入（如 `/jira create "Fix bug" --priority high`）；
- 依靠服务端或第三方自行用正则或空格切分字符串；
- 命令参数缺乏强类型校验、枚举选择和可视化槽位支持；
- 执行响应生命周期单一（只能等待同步返回，或通过异步 Incoming Webhook 发新帖子），缺乏 Discord 风格的“思考中 (Thinking...)”延迟响应和直接弹窗（Command-to-Modal）闭环。

本规范定义了**方案2：自研轻量扩展 Core Slash Command**的完整技术实现方案。通过在 Mattermost 核心模型与前后端交互层中注入结构化参数契约与交互生命周期，使 Mattermost Slash Command 具备媲美 Discord Application Commands 和现代 Slack Block Kit 的交互体验。

---

## 2. 核心特性对比矩阵 (Feature Comparison Matrix)

| 体验维度 | Mattermost 现状 (Legacy) | Discord Application Commands | 本方案改造后 (Rich Slash Commands) |
| :--- | :--- | :--- | :--- |
| **参数类型系统** | 仅支持简单 Text/StaticList/DynamicList | 强类型: String, Integer, Number, Boolean, User, Channel, Role, Attachment | 强类型: `TextInput`, `Integer`, `Number`, `Boolean`, `User`, `Channel`, `Choice` |
| **参数输入交互** | 纯文本输入框，选完建议后退化为字符串 | 胶囊化参数槽 (Pills/Chips)，Tab 键在必填/选填参数间跳转 | 参数专属语义提示（名称、类型、必选状态徽章）与自动补全强化 |
| **参数传递协议** | 传递纯文本 `text: "arg1 arg2"` | 传递结构化 JSON Options 数组与 Resolved 实体映射 | 既保留原始 `text` 字段向下兼容，又额外注入结构化 `options` (JSON Array) |
| **执行生命周期** | 同步 HTTP 请求（超时即报红），不支持异步延迟响应 | 支持 Deferred ("Bot is thinking...")，后续通过 Webhook PATCH 更新 | 支持 `deferred` 响应类型，由系统自动生成短暂的思考状态并支持后续更新 |
| **交互式唤起** | 执行后若要弹窗，需由后端异步调用 Dialog API | 响应直接返回 `type: 9 (MODAL)` 即刻调起表单 | 响应支持 `type: "modal"` 与 TriggerId 配合无缝打开交互式弹窗 |
| **向下兼容性** | 基准 | 无 | **100% 兼容**：现有所有内置命令与老旧外部 Webhook 完全无感知无损运行 |

---

## 3. 架构设计与交互生命周期 (Architecture & Lifecycle)

### 3.1 生命周期流程图 (Sequence Diagram)

```text
[用户 User]                 [Webapp 前端]              [Mattermost Server]             [第三方 Bot / Webhook]
     |                            |                            |                                |
     |-- 1. 输入 /deploy --------->|                            |                                |
     |                            |-- 2. 匹配 Autocomplete --->|                                |
     |                            |<- 3. 返回包含 Options 参数 -|                                |
     |<-- 4. 渲染可视化参数提示 ---|                            |                                |
     |   (env: Choice, dry_run: Bool)                          |                                |
     |                            |                            |                                |
     |-- 5. 回车提交执行 ---------->|                            |                                |
     |                            |-- 6. POST /commands/exec ->|                                |
     |                            |   (含 CommandArgs.Options) |                                |
     |                            |                            |-- 7. 转发 Webhook ------------->|
     |                            |                            |   (携带 text + options json)    |
     |                            |                            |                                |
     |                            |                            |<- 8. 返回 deferred 响应 --------|
     |                            |                            |   {"response_type": "deferred"}|
     |                            |<- 9. 显示 "Thinking..." ---|                                |
     |<-- 10. 展示思考中气泡 ------|   (Ephemeral 临时消息)     |                                |
     |                            |                            |                                |
     |                            |                            |<- 11. 异步 POST response_url --|
     |                            |                            |   (更新为最终执行成功卡片)        |
     |<-- 12. 原地刷新为结果卡片 --|<-- WebSocket 更新推送 -----|                                |
```

---

## 4. 数据模型与协议定义 (Data Models & Protocols)

### 4.1 参数类型定义 (`model.AutocompleteArgType`)

在 `server/public/model/command_autocomplete.go` 中扩展：

```go
const (
    // 现有原生类型
    AutocompleteArgTypeText        AutocompleteArgType = "TextInput"
    AutocompleteArgTypeStaticList  AutocompleteArgType = "StaticList"
    AutocompleteArgTypeDynamicList AutocompleteArgType = "DynamicList"

    // 扩展强类型 (Discord 级别)
    AutocompleteArgTypeInteger  AutocompleteArgType = "Integer"
    AutocompleteArgTypeNumber   AutocompleteArgType = "Number"
    AutocompleteArgTypeBoolean  AutocompleteArgType = "Boolean"
    AutocompleteArgTypeUser     AutocompleteArgType = "User"
    AutocompleteArgTypeChannel  AutocompleteArgType = "Channel"
    AutocompleteArgTypeChoice   AutocompleteArgType = "Choice"
)
```

### 4.2 选项与参数结构 (`AutocompleteChoice` & `CommandOptionArg`)

```go
// AutocompleteChoice 代表固定枚举选项（对应 Discord Choices）
type AutocompleteChoice struct {
    Name  string `json:"name"`  // 面向用户展示的标签，如 "生产环境 (Prod)"
    Value any    `json:"value"` // 传给后端的真实值，如 "production"
}

// CommandOptionArg 代表执行时被解析出来的结构化实参
type CommandOptionArg struct {
    Name     string `json:"name"`               // 参数名
    Type     AutocompleteArgType `json:"type"`  // 参数类型
    Value    any    `json:"value"`              // 解析后的类型值 (string, int, bool 等)
    RawValue string `json:"raw_value"`          // 原始字符串输入
}
```

### 4.3 执行参数扩展 (`model.CommandArgs`)

在 `server/public/model/command_args.go` 中引入：

```go
type CommandArgs struct {
    UserId          string             `json:"user_id"`
    ChannelId       string             `json:"channel_id"`
    TeamId          string             `json:"team_id"`
    RootId          string             `json:"root_id"`
    ParentId        string             `json:"parent_id"`
    TriggerId       string             `json:"trigger_id,omitempty"`
    ConnectionId    string             `json:"connection_id,omitempty"`
    Command         string             `json:"command"`
    SiteURL         string             `json:"-"`
    T               i18n.TranslateFunc `json:"-"`
    UserMentions    UserMentionMap     `json:"-"`
    ChannelMentions ChannelMentionMap  `json:"-"`

    // 新增：结构化参数列表
    Options         []CommandOptionArg `json:"options,omitempty"`
    // 新增：方便快捷读取的参数字典映射
    Parameters      map[string]any     `json:"parameters,omitempty"`
}
```

### 4.4 出站 Webhook 载荷 (Outgoing Webhook Payload)

当向第三方 Webhook 发出 POST 请求时，Mattermost 将自动携带解析后的 `options` 字段：

```http
POST /your-bot-webhook HTTP/1.1
Content-Type: application/x-www-form-urlencoded

token=xxxxxxxxxxxxxxxxxxxx
team_id=team_id_123
team_domain=myteam
channel_id=channel_id_123
channel_name=town-square
user_id=user_id_123
user_name=john
command=%2Fdeploy
text=production+--dry-run+true
trigger_id=trigger_id_123
response_url=https%3A%2F%2Fmattermost.example.com%2Fhooks%2Fcommands%2Fhook_123
options=%5B%7B%22name%22%3A%22env%22%2C%22type%22%3A%22Choice%22%2C%22value%22%3A%22production%22%7D%2C%7B%22name%22%3A%22dry-run%22%2C%22type%22%3A%22Boolean%22%2C%22value%22%3Atrue%7D%5D
```

---

## 5. 响应类型扩展 (`CommandResponse`)

第三方可返回全新的响应指令：

### 5.1 延迟响应 (Deferred / Thinking...)
```json
{
  "response_type": "deferred",
  "text": "机器人正在处理您的部署任务，请稍候..."
}
```
**行为**：Mattermost 会立即为该用户展示一条 Ephemeral 思考中卡片。第三方在任务完成后，调用 `response_url` 发送 POST/PUT 将该卡片替换为完整产物。

### 5.2 直接唤起弹窗 (Command-to-Modal)
```json
{
  "response_type": "modal",
  "props": {
    "dialog": {
      "title": "新建工单",
      "elements": [
        {"display_name": "标题", "name": "title", "type": "text"},
        {"display_name": "严重程度", "name": "severity", "type": "select", "options": [{"text": "P0", "value": "p0"}]}
      ]
    }
  }
}
```

---

## 6. 第三方 Bot 开发实战示例 (Developer Examples)

以 Python Flask 为例，编写一个同时支持纯文本与 Discord 级结构化参数的 Mattermost Bot：

```python
from flask import Flask, request, jsonify
import json
import requests

app = Flask(__name__)

@app.route('/slash/deploy', methods=['POST'])
def handle_deploy():
    token = request.form.get('token')
    user_name = request.form.get('user_name')
    response_url = request.form.get('response_url')
    
    # 尝试读取新版结构化 options
    options_json = request.form.get('options')
    options = json.loads(options_json) if options_json else []
    
    # 转换为字典方便按名访问
    params = {opt['name']: opt['value'] for opt in options}
    env = params.get('env', 'staging')
    dry_run = params.get('dry-run', False)
    
    # 演示：耗时任务，先返回 deferred 延迟响应
    # 如果处理极快，也可以直接返回 in_channel / ephemeral
    return jsonify({
        "response_type": "deferred",
        "text": f"正在向 `{env}` 执行部署 (dry_run={dry_run})，正在拉取容器镜像..."
    })

if __name__ == '__main__':
    app.run(port=5000)
```

---

## 7. 实施计划与审查清单 (Implementation Checklist)

- [x] **架构设计与文档规范编制** (`docs/specs/rich_slash_commands_specification.md`)
- [ ] **公共模型层扩展** (`server/public/model/`):
  - 增加 `AutocompleteArgType` 强类型常量
  - 增加 `AutocompleteChoice` 与 `CommandOptionArg` 结构
  - 扩展 `CommandArgs` 与 `CommandResponse`
- [ ] **服务端解析与调度引擎改造** (`server/channels/app/command.go`):
  - 实现命令参数自动结构化解析器（Named `--key value` 与 Positional 解析）
  - 出站 Webhook 载荷注入 `options` JSON
  - 支持 `deferred` 响应类型的内部处理
- [ ] **Webapp 客户端体验强化** (`webapp/channels/src/components/suggestion/command_provider/`):
  - 在自动补全列表中显示参数的类型徽章与必填指示
- [ ] **单元测试与回归校验**
