# 从 Agent Harness 到 Agentic Post-Training

这个仓库的 ex01–ex15 主要回答一个问题：

> 一个能用工具、记住上下文、遵守权限边界的 Agent Harness 是怎么长出来的？

当这些机制进入后训练以后，每一个工程选择都会变成训练分布的一部分。

## 1. 现有练习和后训练变量的对应关系

| Harness 机制 | 当前练习 | 进入后训练后变成什么 |
| --- | --- | --- |
| provider protocol | ex04 | model / provider distribution |
| tool registry | ex06 | tool schema / action space |
| bash permission | ex07–ex09 | action constraint / safety boundary |
| base prompt | ex08 | system-prompt distribution |
| session JSONL | ex11 | trajectory logging / replay |
| context budget | ex12 | context policy |
| compaction | ex13 | history/state abstraction |
| project rules | ex14 | harness-specific instruction |
| cross-session memory | ex15 | persistent external state |

所以 Agentic Training 不能只写成：

~~~text
LLM + GRPO
~~~

更完整的是：

~~~text
Model
  +
Harness
  +
Environment
  +
Trajectory
  +
Verifier / Reward
  +
Optimizer
~~~

## 2. 为什么同一个模型换 Harness 可能表现不同

假设训练始终使用：

~~~text
read_file
write_file
run_tests
~~~

测试突然换成语义完全相同的：

~~~text
inspect_file
update_file
check_tests
~~~

如果成功率大幅下降，模型学到的可能包含工具名称和 schema pattern，而不只是工具语义。

这就是 harness generalization。

一个干净实验应保持：

- task 不变
- verifier 不变
- max steps 不变
- sampling 参数不变

只替换 tool schema。

## 3. Session JSONL 为什么天然适合做 Trajectory

ex11 已经把多轮交互写成 JSONL。

进入后训练后，一条轨迹最少还要带：

~~~text
task_id
model
harness_version
assistant action
tool arguments
tool observation
final answer
verifier result
reward components
token usage
~~~

这样才能回答：

- reward 为什么高？
- 是不是重复调用工具刷分？
- 哪一步开始失败？
- 换 verifier 后旧轨迹能不能重算？
- successful trajectory 能不能筛出来做 SFT？

## 4. Reward 不应该由 Agent 自己宣布

Agent 说“任务完成”不等于环境真的完成。

coding agent 典型错误：

~~~text
修改 tests
→ public tests PASS
→ naive reward = 1
→ 真实实现仍然错误
~~~

更可靠的结构：

~~~text
agent-visible environment
        ↓
trajectory
        ↓
final state
        ↓
independent verifier
  ├── original public tests
  ├── hidden tests
  └── integrity checks
        ↓
secure reward
~~~

如果 reward 有漏洞，PPO / GRPO 只会更高效地学习利用漏洞。

## 5. Context Policy 也需要 Ablation

ex12–ex15 已经说明 history 不是一个简单字符串。

后训练实验可以固定同一批 task，比较：

~~~text
A: full history
B: remove reasoning-like fields
C: compact older history
D: memory disabled
~~~

同时记录：

- task success
- invalid tool call
- turns
- tokens
- compaction count

不要破坏 tool_call_id 等协议字段，否则测到的是 API 格式错误，不是 history dependence。

## 6. 从 Rollout 到 SFT，再到 RL

完整闭环：

~~~text
task
  ↓
agent harness rollout
  ↓
secure verifier
  ↓
successful trajectories
  ↓
SFT
  ↓
new rollouts
  ↓
verifiable reward
  ↓
GRPO / PPO
  ↓
held-out harness evaluation
~~~

这里 ex01–ex15 负责的是上半段的 Harness。

优化算法和 reward 机制可继续看：

- [RL From Scratch · LLM Post-Training](https://github.com/WonderfulClaire/rl-from-scratch/tree/main/10_rlhf_dpo_grpo)
- [RL From Scratch · Agentic Post-Training](https://github.com/WonderfulClaire/rl-from-scratch/tree/main/12_agentic_post_training)
- [K3 Agentic Post-Training Lab](https://github.com/WonderfulClaire/kimi-k3-deep-dive)

## 7. 面试时怎么解释这三个仓库的关系

可以把它们看成三层：

~~~text
agent-hard-way
    ↓
理解 Harness 到底是什么

rl-from-scratch
    ↓
理解 PPO / GRPO / reward 的数学机制

kimi-k3-deep-dive / K3Lab
    ↓
把 Harness + Verifier + SFT + GRPO 变成可运行实验
~~~

这三层分别回答：

1. Agent 怎么执行？
2. Policy 怎么更新？
3. 怎么证明训练真的改善了 Agent，而不是改善了 benchmark exploit？
