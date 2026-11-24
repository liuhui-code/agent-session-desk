"""Evaluator regressions; replaying archives is not new Windows execution."""

import copy
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import unittest


ROOT = Path(__file__).resolve().parents[3]
EVIDENCE = ROOT / "docs/testing/results/2026-11-24-current-ui-followup-evidence"
SPEC = importlib.util.spec_from_file_location(
    "evaluate_win32", Path(__file__).with_name("evaluate_win32.py")
)
evaluator = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(evaluator)


def good_rows():
    return [
        dict(n=1, x=5, y=4, buttons=0, control=0, flags=1),
        dict(n=2, x=5, y=4, buttons=1, control=0, flags=0),
        dict(n=3, x=9, y=4, buttons=1, control=0, flags=1),
        dict(n=4, x=9, y=4, buttons=0, control=0, flags=0),
    ]


def review(rows):
    for index, row in enumerate(rows, 1):
        row["n"] = index
    intervals = [dict(begin_exclusive=0, end_inclusive=len(rows), mask=0, button=0)]
    return evaluator.evaluate(rows, intervals)


class StrictSequenceTests(unittest.TestCase):
    def test_archived_failure_is_zero_of_nine(self):
        folder = EVIDENCE / "win32-mouse"
        rows = [json.loads(line) for line in (folder / "mouse-packets.ndjson").read_text().splitlines()]
        intervals = json.loads((folder / "mouse-intervals.json").read_text())
        result = evaluator.evaluate(rows, intervals)
        self.assertEqual((result["passed"], result["total"]), (0, 9))

    def test_direct_terminal_control_is_three_of_three(self):
        data = json.loads((EVIDENCE / "win32-control/direct-win32-control.json").read_text())
        intervals = [dict(begin_exclusive=i*4, end_inclusive=(i+1)*4, mask=mask, button=0)
                     for i, mask in enumerate((0, 16, 8))]
        self.assertEqual(evaluator.evaluate(data["received"], intervals)["passed"], 3)

    def test_good_sequence(self):
        self.assertEqual(review(good_rows())["passed"], 1)

    def test_reject_duplicate_premature_reordered_wrong_target_or_modifier(self):
        rows = good_rows()
        mutations = {
            "duplicate down": rows[:2] + [copy.copy(rows[1])] + rows[2:],
            "duplicate up": rows + [copy.copy(rows[3])],
            "premature up": rows[:2] + [copy.copy(rows[3])] + rows[2:],
            "reordered": [rows[0], rows[2], rows[1], rows[3]],
            "missing drag": rows[:2] + rows[3:],
            "wrong target": rows[:2] + [dict(rows[2], x=10)] + rows[3:],
            "wrong modifier": rows[:2] + [dict(rows[2], control=8)] + rows[3:],
            "wrong button": rows[:2] + [dict(rows[2], buttons=4)] + rows[3:],
            "early hover release": rows[:2] + [dict(rows[2], buttons=0)] + rows[2:],
        }
        for name, altered in mutations.items():
            with self.subTest(name=name):
                self.assertEqual(review(copy.deepcopy(altered))["passed"], 0)

    def test_reject_missing_duplicate_or_unreviewed_records(self):
        interval = dict(begin_exclusive=0, end_inclusive=4, mask=0, button=0)
        for rows, intervals in [
            (good_rows(), []),
            (good_rows(), [dict(interval, end_inclusive=3)]),
            (good_rows(), [interval, interval]),
            ([dict(row, n=1) for row in good_rows()], [interval]),
            (good_rows(), [dict(interval, mask=4)]),
            ([None], [interval]),
            (good_rows(), [None]),
            (good_rows(), {"not": "intervals"}),
        ]:
            with self.subTest(rows=rows, intervals=intervals), self.assertRaises(ValueError):
                evaluator.evaluate(rows, intervals)

    def test_cli_failure_and_invalid_evidence_are_nonzero(self):
        folder = EVIDENCE / "win32-mouse"
        script = Path(__file__).with_name("evaluate_win32.py")
        command = [sys.executable, "-B", str(script), str(folder / "mouse-packets.ndjson")]
        failed = subprocess.run(command + [str(folder / "mouse-intervals.json")],
                                capture_output=True, text=True, check=False)
        self.assertEqual(failed.returncode, 1)
        self.assertEqual(json.loads(failed.stdout)["passed"], 0)
        invalid = subprocess.run(command + [str(folder / "absent.json")],
                                 capture_output=True, text=True, check=False)
        self.assertEqual(invalid.returncode, 2)
        self.assertEqual(invalid.stdout, "")


if __name__ == "__main__":
    unittest.main()
