"""Construct data arms and summarize paired held-out utility."""
from __future__ import annotations
from collections import defaultdict
from typing import Any


def build_data_arms(originals: list[dict[str, Any]], repairs: list[dict[str, Any]]) -> dict[str, list[dict[str, Any]]]:
    raw = [r for r in originals if r.get("status") == "completed"]
    filtered = [r for r in raw if r.get("passed") is True and r.get("protocol_valid", True) is True]
    repaired = list(filtered)
    repaired += [
        r for r in repairs
        if r.get("status") == "completed"
        and r.get("passed") is True
        and r.get("replay_matched") is True
        and r.get("protocol_valid", True) is True
    ]
    return {"raw": raw, "filtered": filtered, "repaired": repaired}


def paired_delta(control: list[dict[str, Any]], candidate: list[dict[str, Any]]) -> dict[str, Any]:
    left = {r["slot_id"]: bool(r["passed"]) for r in control}
    right = {r["slot_id"]: bool(r["passed"]) for r in candidate}
    if set(left) != set(right):
        raise ValueError("control and candidate must contain identical paired slots")

    wins = losses = ties = 0
    per_task = defaultdict(list)
    for slot in sorted(left):
        if not left[slot] and right[slot]: wins += 1
        elif left[slot] and not right[slot]: losses += 1
        else: ties += 1
        per_task[slot.split("::", 1)[0]].append((left[slot], right[slot]))

    n = len(left)
    control_sr = sum(left.values()) / max(1, n)
    candidate_sr = sum(right.values()) / max(1, n)
    regressions = [
        task_id for task_id, pairs in sorted(per_task.items())
        if sum(y for _, y in pairs) / len(pairs) < sum(x for x, _ in pairs) / len(pairs)
    ]
    return {
        "paired_trials": n,
        "control_success_rate": control_sr,
        "candidate_success_rate": candidate_sr,
        "delta_percentage_points": 100.0 * (candidate_sr - control_sr),
        "wins": wins,
        "losses": losses,
        "ties": ties,
        "regressed_tasks": regressions,
    }
