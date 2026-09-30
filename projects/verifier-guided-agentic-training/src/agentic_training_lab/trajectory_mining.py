"""Hard-trajectory mining using visible execution evidence only."""
from __future__ import annotations
from collections import Counter
from typing import Any


def trajectory_features(row: dict[str, Any]) -> dict[str, Any]:
    events = list(row.get("events") or [])
    commands = [e["command"].strip() for e in events if isinstance(e.get("command"), str) and e["command"].strip()]
    counts = Counter(commands)
    repeated = sum(max(0, n - 1) for n in counts.values())
    protocol_errors = sum(bool(e.get("protocol_error")) for e in events)
    budget = row.get("budget") or {}
    max_steps = int(budget.get("max_steps") or max(1, len(events)))
    step_ratio = min(1.0, len(events) / max(1, max_steps))

    status = str(row.get("status", "unknown"))
    passed = row.get("passed") is True
    completed_failure = status == "completed" and row.get("passed") is False
    infrastructure_failure = status != "completed"
    exhausted = completed_failure and step_ratio >= 0.95

    reasons = []
    if infrastructure_failure: reasons.append("infrastructure")
    if protocol_errors: reasons.append("protocol")
    if repeated: reasons.append("repeated_command")
    if exhausted: reasons.append("budget_exhausted")
    if completed_failure: reasons.append("verified_failure")

    if infrastructure_failure:
        score = -1.0
    else:
        score = 1.0 if completed_failure else -0.5
        score += min(0.35, 0.10 * protocol_errors)
        score += min(0.25, 0.05 * repeated)
        if completed_failure:
            score += 0.25 * step_ratio

    return {
        "score": round(score, 6),
        "status": status,
        "passed": passed,
        "events": len(events),
        "step_ratio": step_ratio,
        "protocol_errors": protocol_errors,
        "repeated_commands": repeated,
        "reasons": reasons,
    }


def rank_hard_trajectories(rows: list[dict[str, Any]], top_k: int | None = None) -> list[dict[str, Any]]:
    ranked, seen = [], set()
    for row in rows:
        trial_id = row.get("trial_id")
        if not isinstance(trial_id, str) or not trial_id:
            raise ValueError("each trajectory needs a non-empty trial_id")
        if trial_id in seen:
            raise ValueError("duplicate trial_id")
        seen.add(trial_id)
        features = trajectory_features(row)
        if features["status"] != "completed" or features["passed"]:
            continue
        ranked.append({"trial_id": trial_id, "task_id": row.get("task_id"), **features})

    ranked.sort(key=lambda item: (-item["score"], str(item.get("task_id")), item["trial_id"]))
    if top_k is not None:
        if top_k < 1:
            raise ValueError("top_k must be positive")
        ranked = ranked[:top_k]
    return ranked
