# Learn Agent the Hard Way — 我的逐章手写练习

> 跟着 [Leihb/learn-agent-the-hard-way](https://github.com/Leihb/learn-agent-the-hard-way)（中文电子书《从 60 行 Go 代码开始，亲手写出一个 agent》），**不复制粘贴、每一章的代码都自己敲**——把 agent = LLM + tool use 这套东西从地基写到生产级 harness。

`agent = LLM + tool use`。模型你改不了，循环只有几十行——一个 agent 和另一个 agent 的全部差别，都在**工具的设计**里。这本书带你从一次裸 HTTP 请求开始，把工具循环、权限、上下文、skill、subagent、MCP、浏览器一个个亲手写出来。仓库里每一章 `exNN/main.go` 都是我**照着书敲出来、并 `go build` 验证过能编译**的实现（非复制粘贴）。

## 进度

| Part | 练习 | 内容 | 状态 |
|------|------|------|------|
| Part 0 | ex00 | 什么是 harness——模型不是产品 | 📖 概念 |
| Part 1 · 最小对话 | ex01 | 一次 API 调用 | ✅ 已敲 |
| | ex02 | 流式输出（SSE） | ✅ 已敲 |
| | ex03 | 多轮对话——messages 数组 + for 循环 | ✅ 已敲 |
| | ex04 | provider 抽象——同一份代码接两种协议 | ✅ 已敲 |
| Part 2 · 长出手脚 | ex05 | 第一个工具——agent loop 完整闭环 | ✅ 已敲 |
| | ex06 | 工具注册表——声明/执行分离 + read-before-write | ✅ 已敲 |
| | ex07 | bash 特权工具——超时/截断/固定 cwd | ✅ 已敲 |
| | ex08 | base prompt——软约束 vs 硬约束 + prompt cache | ✅ 已敲 |
| | ex09 | 权限系统——deny/ask/allow 三档闸门 | ✅ 已敲 |
| | ex10 | 误删保护——覆盖前备份 + 一键恢复 | ✅ 已敲 |
| Part 3 · 记住事情 | ex11–ex15 | 会话持久化 / 上下文预算 / 压缩 / 规则文件 / 跨会话记忆 | ⬜ |
| Part 4 · 长出知识 | ex16–ex18 | 最小 skill 加载器 / 按需触发 / 为何不让 agent 自写 skill | ⬜ |
| Part 5 · 长出分身 | ex19–ex20 | 第一个 subagent / 并行扇出与上限 | ⬜ |

## 怎么跑

每个 `exNN` 是独立 Go module。需要 **Go 1.22+** 和一个会说 OpenAI 协议的模型（云端 key 或本机 Ollama 均可，全书每个练习同时支持）。

```bash
# 云端（示例用 DeepSeek）
export OPENAI_BASE_URL=https://api.deepseek.com/v1
export OPENAI_API_KEY=sk-你的key
export MODEL=deepseek-v4-flash

# 或本机 Ollama（免费，断网可跑）
# export OPENAI_BASE_URL=http://localhost:11434/v1
# export OPENAI_API_KEY=ollama
# export MODEL=qwen3:4b-instruct

cd ex03 && go run .        # 进 REPL，输入 exit 退出
# 或
cd ex01 && go build -o ex01 . && ./ex01 "用一句话说明什么是 harness"
```

ex04 额外支持 Anthropic 协议（DeepSeek 的 `/anthropic` 兼容端点或本机 Ollama，都无需新 key）：

```bash
export PROTOCOL=anthropic
export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_API_KEY=sk-你的DeepSeek-key
export MODEL=deepseek-v4-flash
cd ex04 && go run .
```

> 注意：本仓库代码只验证过 `go build` 编译。**真正跑通需要你自己的模型 key**——我没有把任何 key 写进代码或提交。

## 仓库结构

```
agent-hard-way/
├── README.md
├── ex01/main.go   # 一次 API 调用
├── ex02/main.go   # 流式输出
├── ex03/main.go   # 多轮对话
├── ex04/main.go   # provider 抽象（OpenAI + Anthropic）
├── ex05/main.go   # 第一个工具 + agent loop 闭环
├── ex06/main.go   # 工具注册表 + read-before-write
├── ex07/main.go   # bash 特权工具
├── ex08/main.go   # base prompt + prompt cache
├── ex09/main.go   # 权限系统（deny/ask/allow）
├── ex10/main.go   # 误删保护 + 恢复
├── ex11..ex20/    # 后续章节，逐步补齐
└── LICENSE        # MIT（练习代码沿用原书 exercises/ 的 MIT 许可）
```

## 许可

练习实现沿用原书 [`exercises/` 的 MIT 许可](https://github.com/Leihb/learn-agent-the-hard-way/blob/main/LICENSE)；正文概念归原作者 Leihb 所有（原书正文 CC BY-NC-SA 4.0）。本仓库仅为个人学习用途的逐章手写练习。
