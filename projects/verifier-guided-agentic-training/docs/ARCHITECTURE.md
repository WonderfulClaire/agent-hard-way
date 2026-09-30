# Architecture

## Task synthesis
Each generated task should have a visible instruction, initial workspace, reference behavior, and verifier contract. Generation and acceptance are separate stages.

## Isolated execution
Agent runs preserve model revision, prompt/config, actions, terminal observations, token usage when available, runtime status, and final verifier outcome. Infrastructure failures stay separate from model failures.

## Verifier audit
Use development probes for repair and frozen held-out probes for unseen validation. A verifier pass is not treated as truth unless the verifier itself has been tested.

## Failure attribution
Completed failures are decomposed into protocol errors, repeated actions, budget exhaustion, and other behavior-level causes. Infrastructure failures are excluded from training-mining priority.

## Replay and repair
Before continuing a failure, replay its visible command history in a clean environment. Only reproducible failures are eligible for continuation training.

## Post-training data arms
Raw = all completed traces; Filtered = verified successes; Repaired = filtered successes + verified repair continuations.

## Held-out evaluation
Use paired task/seed evaluation. Report regressions and uncertainty, not only one global average.
