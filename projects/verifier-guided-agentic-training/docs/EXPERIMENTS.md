# Experiment protocol

## Main question
Does verifier-guided curation or repair improve terminal-agent performance relative to raw execution data under a matched supervision budget?

## Training arms

| Arm | Source |
|---|---|
| Base | no student post-training |
| Raw-SFT | all completed train trajectories |
| Filtered-SFT | verified successful train trajectories |
| Repaired-SFT | filtered successes + successful replay/repair trajectories |

Use at least two training seeds; three is preferred when compute allows.

## Controls
Keep fixed: base revision, tokenizer/chat template, LoRA/optimizer, supervised assistant-action token budget, split, evaluation tasks, sampling seeds, step/token limits, harness, and verifier version.

## Main table

| Student | Dev SR | Test SR | Delta vs Base | Wins | Losses | Regressed tasks |
|---|---:|---:|---:|---:|---:|---:|
| Base | | | | | | |
| Raw-SFT | | | | | | |
| Filtered-SFT | | | | | | |
| Repaired-SFT | | | | | | |

## Claim gate
Only frozen held-out test evaluation should supply resume improvement numbers. Do not substitute verifier probe accuracy, train reward, train loss, reference-solution success, a single development task, or rescore-only experiments.
