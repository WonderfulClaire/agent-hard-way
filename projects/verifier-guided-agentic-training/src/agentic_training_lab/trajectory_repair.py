"""Replay validation before continuing a failed agent trajectory."""
from __future__ import annotations
from typing import Any, Callable


def validate_replay(events: list[dict[str, Any]], execute: Callable[[str], Any]) -> dict[str, Any]:
    receipt = []
    for index, event in enumerate(events):
        if event.get("protocol_error"):
            receipt.append({"event": index, "skipped": "protocol_error"})
            continue
        command = event.get("command")
        if not isinstance(command, str) or not command.strip():
            continue
        if "result" not in event:
            raise ValueError("command event is missing its recorded result")
        actual = execute(command)
        matches = actual == event["result"]
        receipt.append({"event": index, "command": command, "matches": matches})
        if not matches:
            return {"replay_matched": False, "first_divergence": index, "receipt": receipt}
    return {"replay_matched": True, "first_divergence": None, "receipt": receipt}
