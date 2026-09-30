import unittest
from agentic_training_lab.data_utility import build_data_arms, paired_delta
from agentic_training_lab.trajectory_mining import rank_hard_trajectories
from agentic_training_lab.trajectory_repair import validate_replay
from agentic_training_lab.verifier_audit import audit_verifier, exact_semantic_verifier


class PublicPipelineTest(unittest.TestCase):
    def test_hard_mining_excludes_infrastructure(self):
        rows = [
            {"trial_id": "hard", "task_id": "t1", "status": "completed", "passed": False,
             "budget": {"max_steps": 2}, "events": [{"command": "ls"}, {"command": "ls", "protocol_error": True}]},
            {"trial_id": "infra", "task_id": "t2", "status": "api_or_harness_error", "passed": False,
             "events": [{"command": "ls"}]},
        ]
        self.assertEqual([r["trial_id"] for r in rank_hard_trajectories(rows)], ["hard"])

    def test_replay_requires_exact_visible_state(self):
        events = [{"command": "echo x", "result": {"stdout": "x\n", "returncode": 0}}]
        self.assertTrue(validate_replay(events, lambda _: {"stdout": "x\n", "returncode": 0})["replay_matched"])
        self.assertFalse(validate_replay(events, lambda _: {"stdout": "y\n", "returncode": 0})["replay_matched"])

    def test_verifier_audit(self):
        expected = [{"id": 1, "value": 3}]
        good = {"text": '[{"value": 3, "id": 1}]', "expected": expected,
                "protected": {"input": "abc"}, "expected_protected": {"input": "abc"}}
        wrong = dict(good, text='[{"value": 4, "id": 1}]')
        report = audit_verifier(exact_semantic_verifier, [
            {"name": "valid", "should_accept": True, "payload": good},
            {"name": "wrong", "should_accept": False, "payload": wrong},
        ])
        self.assertTrue(report["passed"])

    def test_data_arms_and_paired_delta(self):
        arms = build_data_arms(
            [{"status": "completed", "passed": True}, {"status": "completed", "passed": False}],
            [{"status": "completed", "passed": True, "replay_matched": True}],
        )
        self.assertEqual([len(arms[k]) for k in ("raw", "filtered", "repaired")], [2, 1, 2])
        report = paired_delta(
            [{"slot_id": "t1::0", "passed": False}, {"slot_id": "t2::0", "passed": True}],
            [{"slot_id": "t1::0", "passed": True}, {"slot_id": "t2::0", "passed": True}],
        )
        self.assertEqual((report["wins"], report["losses"]), (1, 0))


if __name__ == "__main__":
    unittest.main()
