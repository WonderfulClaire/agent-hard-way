# Resume wording

**Verifier-Guided Agentic Training for Terminal Agents｜个人项目｜GitHub**

- **Agentic Data Engine：** 构建“任务合成—隔离执行—自动验证—轨迹采集—训练—评测”闭环，围绕长程终端任务生成可执行训练样本，并将 Verifier 可靠性纳入数据流水线。
- **Verifier Audit：** 设计 Development Probes + Frozen Held-out Probes 双阶段审计，修复阶段仅暴露开发反例，冻结后再用未见探针验证误接受/误拒风险。
- **Trajectory Repair & Hard Mining：** 对真实 Agent 失败轨迹进行隔离重放，确认失败可复现后继续修复；按 Verified Failure、协议错误、重复命令与预算耗尽挖掘困难轨迹，同时排除环境/API 故障样本。
- **Post-training Data Utility：** 构建 Raw / Filtered / Repaired 三类训练数据，在相同 supervised action-token budget 下进行后训练，并使用固定 Harness 与独立 Held-out Tasks 做成对评测，统计任务级回归与训练 Seed 方差。
