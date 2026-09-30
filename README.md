# Learn Agent the Hard Way — 我的逐章手写练习

[![Verify exercises](https://github.com/WonderfulClaire/agent-hard-way/actions/workflows/verify.yml/badge.svg)](https://github.com/WonderfulClaire/agent-hard-way/actions/workflows/verify.yml)

> 跟着 [Leihb/learn-agent-the-hard-way](https://github.com/Leihb/learn-agent-the-hard-way)（中文电子书《从 60 行 Go 代码开始，亲手写出一个 agent》），**不复制粘贴、每一章的代码都自己敲**——从最小 API 调用开始，逐步理解 agent harness 的核心机制。

`agent = LLM + tool use`。这个仓库不是一个包装好的 agent 产品，而是一条可追踪的学习路径：从裸 HTTP 请求、工具循环，逐步走到权限、上下文压缩和跨会话记忆。当前已完成 `ex01`–`ex20`；每章都是独立 Go module，并由 CI 逐章执行 `go test`、`go vet` 和 `go build`。后五章把 harness 继续扩展到 skill 与 subagent：先做最小加载器，再做按需触发、skill 写保护、任务委派与有限并发 fan-out。

## 你能在这里看到什么

- **递进式实现**：每个目录只引入一组新机制，方便对比 agent 能力是怎样长出来的。
- **安全边界练习**：包括 read-before-write、bash 超时/输出截断、deny/ask/allow 权限和覆盖前备份。
- **上下文工程**：会话持久化、token 预算、历史压缩、分层规则与跨会话记忆。
- **可重复验证**：本地一条命令与 GitHub Actions 使用同一套检查。
- **从 Harness 接到后训练**：新增 [Agentic Post-Training Bridge](docs/agentic-post-training-bridge.md)，把 provider、tool schema、session、context、memory 映射到 trajectory、verifier、reward 与 held-out harness evaluation。


## 独立项目：Verifier-Guided Agentic Training

仓库中新增一个可独立运行的 Python 项目：**[Verifier-Guided Agentic Training for Terminal Agents](projects/verifier-guided-agentic-training/)**。

它把 Harness 进一步接到后训练数据闭环：**Verifier Audit → Hard Trajectory Mining → Replay/Repair → Raw / Filtered / Repaired Data → Paired Held-out Evaluation**。该目录只包含可公开的最小实现与实验协议，不包含私有模型、服务器路径、运行产物或凭证。

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
| Part 3 · 记住事情 | ex11 | 会话持久化——history 写盘成 JSONL，增量追加 + 崩溃后丢半行恢复 | ✅ 已敲 |
| | ex12 | 上下文预算——窗口对照 + 估算/真实双值 + 75% 门槛告警 | ✅ 已敲 |
| | ex13 | 压缩——让模型总结旧对话，safeSplitIndex 落 user 边界 + 整重写 | ✅ 已敲 |
| | ex14 | 规则文件——.harnessrules 分层拼进 system prompt | ✅ 已敲 |
| | ex15 | 跨会话记忆——MEMORY.md 模型自写自读，记错=改文件 | ✅ 已敲 |
| Part 4 · 长出知识 | ex16 | 最小 skill 加载器：读取 `SKILL.md` 元数据与正文 | ✅ 已敲 |
| | ex17 | 按需触发：基于 trigger / description 的最小技能路由 | ✅ 已敲 |
| | ex18 | Skill 安全边界：技能目录默认只读，防止 agent 静默改写自己的执行策略 | ✅ 已敲 |
| Part 5 · 长出分身 | ex19 | 第一个 subagent：显式 task scope + result identity | ✅ 已敲 |
| | ex20 | 并行 fan-out：并发上限 + context 取消 + 结果归并 | ✅ 已敲 |
| Bridge · 从执行到学习 | 文档 | Harness → Trajectory → Verifier → SFT / GRPO | ✅ |

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

> 注意：CI 只做不访问模型的静态验证。**真正跑通需要你自己的模型 key**——仓库不包含任何 key。

## 本地验证

在仓库根目录执行：

```bash
./scripts/verify.sh
```

该脚本会发现所有 `ex*/go.mod`，并对每个章节依次执行 `go test ./...`、`go vet ./...` 和 `go build ./...`。GitHub Actions 会在 Go 1.22 和当前稳定版上执行同一脚本。

## 仓库结构

```
agent-hard-way/
├── README.md
├── scripts/verify.sh # 逐章 test + vet + build
├── .github/workflows/verify.yml
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
├── ex11/main.go   # 会话持久化（JSONL 增量写盘 + 崩溃恢复）
├── ex12/main.go   # 上下文预算（窗口对照 + 估算/真实双值）
├── ex13/main.go   # 压缩（模型自总结旧对话 + 安全分割 + 整重写）
├── ex14/main.go   # 规则文件（.harnessrules 分层拼装）
├── ex15/main.go   # 跨会话记忆（MEMORY.md 模型自写自读）
├── ex16/main.go   # skill loader
├── ex17/main.go   # skill trigger / routing
├── ex18/main.go   # skill write-protection boundary
├── ex19/main.go   # first scoped subagent
├── ex20/main.go   # bounded parallel fan-out
└── LICENSE        # MIT（练习代码沿用原书 exercises/ 的 MIT 许可）
```

## 许可

练习实现沿用原书 [`exercises/` 的 MIT 许可](https://github.com/Leihb/learn-agent-the-hard-way/blob/main/LICENSE)；正文概念归原作者 Leihb 所有（原书正文 CC BY-NC-SA 4.0）。本仓库仅为个人学习用途的逐章手写练习。
