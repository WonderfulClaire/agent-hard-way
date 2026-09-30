"""Verifier-audit primitives for synthetic terminal tasks."""
from __future__ import annotations
import json
from typing import Any, Callable


def _strict_json(text: str) -> Any:
    def unique_object(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("duplicate JSON key")
            result[key] = value
        return result
    return json.loads(
        text,
        object_pairs_hook=unique_object,
        parse_constant=lambda value: (_ for _ in ()).throw(ValueError("non-finite JSON")),
    )


def _same_type_and_value(left: Any, right: Any) -> bool:
    if type(left) is not type(right):
        return False
    if isinstance(left, list):
        return len(left) == len(right) and all(_same_type_and_value(a, b) for a, b in zip(left, right))
    if isinstance(left, dict):
        return left.keys() == right.keys() and all(_same_type_and_value(left[k], right[k]) for k in left)
    return left == right


def exact_semantic_verifier(payload: dict[str, Any]) -> bool:
    try:
        actual = _strict_json(payload["text"])
        expected = payload["expected"]
        return (
            _same_type_and_value(actual, expected)
            and payload.get("protected") == payload.get("expected_protected")
        )
    except (KeyError, TypeError, ValueError, json.JSONDecodeError):
        return False


def audit_verifier(verifier: Callable[[dict[str, Any]], bool], probes: list[dict[str, Any]]) -> dict[str, Any]:
    rows = []
    for probe in probes:
        expected = bool(probe["should_accept"])
        actual = bool(verifier(probe["payload"]))
        rows.append({"name": probe["name"], "expected": expected, "actual": actual, "correct": actual is expected})
    return {
        "passed": all(r["correct"] for r in rows),
        "probes": len(rows),
        "false_accepts": sum((not r["expected"]) and r["actual"] for r in rows),
        "false_rejects": sum(r["expected"] and (not r["actual"]) for r in rows),
        "rows": rows,
    }
