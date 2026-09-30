from pprint import pprint
from agentic_training_lab.data_utility import build_data_arms
from agentic_training_lab.trajectory_mining import rank_hard_trajectories
from agentic_training_lab.verifier_audit import audit_verifier, exact_semantic_verifier


def main():
    trajectories = [
        {"trial_id": "trial-hard", "task_id": "task-a", "status": "completed", "passed": False,
         "budget": {"max_steps": 4},
         "events": [{"command": "cat input.txt"}, {"command": "cat input.txt"},
                    {"protocol_error": True}, {"command": "python fix.py"}]},
        {"trial_id": "trial-success", "task_id": "task-b", "status": "completed", "passed": True,
         "events": [{"command": "pytest -q"}]},
    ]
    print("Hard trajectories")
    pprint(rank_hard_trajectories(trajectories))

    print("\nData arms")
    pprint(build_data_arms(
        [{"status": "completed", "passed": True}, {"status": "completed", "passed": False}],
        [{"status": "completed", "passed": True, "replay_matched": True}],
    ))

    expected = [{"id": 1, "score": 7}]
    probes = [
        {"name": "valid-format-variant", "should_accept": True,
         "payload": {"text": '[{"score":7, "id":1}]', "expected": expected,
                     "protected": {"input": "hash"}, "expected_protected": {"input": "hash"}}},
        {"name": "wrong-value", "should_accept": False,
         "payload": {"text": '[{"score":8, "id":1}]', "expected": expected,
                     "protected": {"input": "hash"}, "expected_protected": {"input": "hash"}}},
    ]
    print("\nVerifier audit")
    pprint(audit_verifier(exact_semantic_verifier, probes))


if __name__ == "__main__":
    main()
