"""Synthetic checks for statistics and missing observations, not Windows evidence."""

from contextlib import redirect_stdout
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest


SPEC = importlib.util.spec_from_file_location("asd_resource_summary", Path(__file__).with_name("summarize_resources.py"))
resource = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(resource)


def process(pid, private=10, *, owner="backend_server", role=None, created="133700000000000001", cpu=0, **changes):
    row = {
        "process_id": pid, "creation_filetime": created, "executable_path": rf"C:\isolated\{owner}.exe",
        "role": role or owner, "owner_role": owner, "is_descendant": role == "console_aux",
        "cpu_total_ms": cpu, "private_bytes": private, "working_set_bytes": 2 * private if private is not None else None,
        "handle_count": 6, "thread_count": 2,
    }
    row.update(changes)
    return row


def sample(index, rows, *, scenario="managed-1", phase="idle", count=1, gaps=None, elapsed=None):
    elapsed = index * 1000 if elapsed is None else elapsed
    rows = [dict(row) for row in rows] if rows is not None else None
    for row in rows or []:
        row.setdefault("observed_elapsed_ms", elapsed + 100)
    return {
        "scenario": scenario, "phase": phase, "session_count": count,
        "snapshot": {
            "sample_index": index, "sample_elapsed_ms": elapsed, "sample_duration_ms": 150,
            "cim_query_duration_ms": 80, "processes": rows, "gaps": gaps or [],
        },
    }


def summary(count, *, first=0, elapsed=None):
    return {
        "commit": "synthetic-test-only", "result": "pass", "logical_processors": 4,
        "phases": [{"scenario": "managed-1", "phase": "idle", "session_count": 1,
                    "first_sample": first, "sample_count": count, "elapsed_ms": elapsed if elapsed is not None else count * 1000}],
    }


def group(analysis, name):
    return next(row for row in analysis["phases"][0]["groups"] if row["id"] == name)


class ResourceStatisticsTests(unittest.TestCase):
    def test_percentiles_have_explicit_interpolation_and_keep_missing_values(self):
        stats = resource.distribution([0, 10, 20, 30, None])
        self.assertEqual((stats["p50"], stats["max"]), (15, 30))
        self.assertAlmostEqual(stats["p95"], 28.5)
        self.assertEqual((stats["samples"], stats["valid_samples"], stats["missing_samples"]), (5, 4, 1))
        self.assertEqual(resource.distribution([7])["p95"], 7)
        self.assertIsNone(resource.distribution([None, float("nan"), float("inf"), True, -1])["p50"])

    def test_sum_each_snapshot_before_taking_peaks(self):
        rows = [sample(0, [process(1, 100), process(2, 1)]), sample(1, [process(1, 1), process(2, 100)])]
        result = group(resource.summarize(summary(2), rows), "product_total")
        # The processes peak at different times: 100+100 would be fabricated.
        self.assertEqual(result["metrics"]["private_bytes"]["observed_sum"]["max"], 101)
        self.assertEqual(result["observed_process_count"]["max"], 2)

    def test_owner_groups_include_auxiliaries_and_exclude_driver_sampler(self):
        rows = [
            process(1, 10, owner="product_manager"), process(2, 20), process(3, 5, owner="backend_client"),
            process(4, 3, role="console_aux"), process(5, 2, owner="product_manager", role="console_aux"),
            process(6, 7, owner="managed_fixture"), process(7, 100, owner="direct_fixture"),
            process(8, 1000, owner="measurement_driver"), process(9, 2000, owner="sampler"),
        ]
        result = resource.summarize(summary(1), [sample(0, rows)])
        for name, expected in (("product_total", 40), ("managed_total", 47), ("backend_total", 28), ("measurement_driver", 1000), ("sampler", 2000)):
            self.assertEqual(group(result, name)["metrics"]["private_bytes"]["observed_sum"]["max"], expected, name)
        details = result["phases"][0]["role_owner_groups"]
        auxiliary = next(row for row in details if row["role"] == "console_aux" and row["owner_roles"] == ["backend_server"])
        self.assertEqual(auxiliary["metrics"]["private_bytes"]["observed_sum"]["max"], 3)
        legacy = resource.summarize(summary(1), [sample(0, [process(1, 9, owner="backend_process")])])
        self.assertEqual(group(legacy, "product_total")["metrics"]["private_bytes"]["observed_sum"]["max"], 9)

    def test_missing_metric_is_partial_not_a_zero_or_complete_total(self):
        result = resource.summarize(summary(1), [sample(0, [process(1, 10), process(2, None)])])
        backend = group(result, "backend_server")
        metric = backend["metrics"]["private_bytes"]
        self.assertEqual(metric["observed_sum"]["p50"], 10)
        self.assertIsNone(metric["complete_observation_sum"]["p50"])
        self.assertEqual(metric["missing_process_values"], 1)
        self.assertEqual(backend["metrics"]["handle_count"]["complete_observation_sum"]["p50"], 12)
        self.assertFalse(backend["population_complete"])
        all_missing = resource.summarize(summary(1), [sample(0, [process(1, None)])])
        self.assertIsNone(group(all_missing, "backend_server")["metrics"]["private_bytes"]["observed_sum"]["max"])

    def test_gap_scope_and_departure_observations_are_distinct(self):
        rows = [process(1, 10, owner="product_manager"), process(2, 20)]
        hard = {"reason": "process_read_failed", "owner_role": "backend_server"}
        result = resource.summarize(summary(1), [sample(0, rows, gaps=[hard])])
        self.assertEqual(group(result, "product_manager")["metrics"]["private_bytes"]["complete_observation_sum"]["max"], 10)
        backend = group(result, "backend_server")
        self.assertEqual(backend["metrics"]["private_bytes"]["observed_sum"]["max"], 20)
        self.assertIsNone(backend["metrics"]["private_bytes"]["complete_observation_sum"]["max"])
        unknown = resource.summarize(summary(1), [sample(0, [], gaps=[{"reason": "cim_snapshot_failed", "owner_role": None}])])
        self.assertIsNone(group(unknown, "sampler")["metrics"]["private_bytes"]["observed_sum"]["max"])
        soft = resource.summarize(summary(1), [sample(0, rows, gaps=[{"reason": "previous_owned_not_observed", "owner_role": "backend_server"}])])
        backend = group(soft, "backend_server")
        self.assertEqual(backend["hard_gap_observations"], 0)
        self.assertEqual(backend["departure_observations"], 1)
        self.assertEqual(backend["metrics"]["private_bytes"]["complete_observation_sum"]["max"], 20)

    def test_pid_reuse_or_changed_image_never_creates_a_cpu_delta(self):
        for changes in ({"created": "133700000000000002"}, {"executable_path": r"C:\different.exe"}):
            with self.subTest(changes=changes):
                result = resource.summarize(summary(2), [sample(0, [process(7, cpu=10)]), sample(1, [process(7, cpu=999, **changes)])])
                cpu = group(result, "backend_server")["cpu"]
                self.assertIsNone(cpu["observed_cpu_ms_lower_bound"])
                self.assertEqual(cpu["matched_identity_pairs"], 0)
                self.assertEqual(cpu["unpaired_current_rows"], 1)

    def test_cpu_uses_per_process_actual_time_and_separate_phase_wall_time(self):
        rows = [
            sample(0, [process(1, cpu=100, observed_elapsed_ms=100, executable_path=r"C:\ASD\Worker.exe")]),
            sample(1, [process(1, cpu=150, observed_elapsed_ms=2100, executable_path="c:/asd/worker.exe")]),
        ]
        cpu = group(resource.summarize(summary(2, elapsed=3000), rows), "backend_server")["cpu"]
        self.assertEqual(cpu["observed_cpu_ms_lower_bound"], 50)
        self.assertEqual(cpu["observation_span_ms"], 2000)
        self.assertEqual(cpu["interval_one_core_percent_observed"]["p50"], 2.5)
        self.assertAlmostEqual(cpu["one_core_percent_lower_bound_over_phase"], 50 / 3000 * 100)
        self.assertEqual(cpu["intervals_with_complete_pairing"], 1)

    def test_counter_reset_missing_counter_and_missing_sample_are_not_zero_cpu(self):
        for changes, reason in (({"cpu": 1}, "cpu_counter_decreased"), ({"cpu": None}, "cpu_counter_or_time_missing"), ({"observed_elapsed_ms": None}, "cpu_counter_or_time_missing")):
            with self.subTest(changes=changes):
                after = {"cpu": 20, **changes}
                cpu = group(resource.summarize(summary(2), [sample(0, [process(1, cpu=10)]), sample(1, [process(1, **after)])]), "backend_server")["cpu"]
                self.assertIsNone(cpu["observed_cpu_ms_lower_bound"])
                self.assertEqual(cpu["interval_records"][0]["invalid_pair_reasons"], {reason: 1})
        result = resource.summarize(summary(3), [sample(0, [process(1, 5, cpu=0)]), sample(2, [process(1, 7, cpu=200)])])
        self.assertEqual(result["phases"][0]["missing_sample_indices"], [1])
        backend = group(result, "backend_server")
        self.assertEqual(backend["metrics"]["private_bytes"]["observed_sum"]["p50"], 6)
        self.assertEqual(backend["metrics"]["private_bytes"]["observed_sum"]["missing_samples"], 1)
        self.assertIsNone(backend["cpu"]["observed_cpu_ms_lower_bound"])

    def test_cpu_partial_pairs_remain_visible_but_not_complete(self):
        rows = [sample(0, [process(1, cpu=100), process(2, cpu=0)]), sample(1, [process(1, cpu=200), process(2, cpu=None)])]
        cpu = group(resource.summarize(summary(2), rows), "product_total")["cpu"]
        self.assertEqual(cpu["observed_cpu_ms_lower_bound"], 100)
        self.assertEqual(cpu["interval_one_core_percent_observed"]["p50"], 10)
        self.assertEqual(cpu["interval_one_core_percent_complete_observation"]["valid_samples"], 0)

    def test_only_registered_phase_samples_are_analyzed(self):
        rows = [
            sample(0, [process(1, 999, cpu=0)], phase="ui-ready"),
            sample(1, [process(1, 10, cpu=100)]), sample(2, [process(1, 20, cpu=120)]),
            sample(3, [process(1, 999, cpu=999)], phase="cleanup"),
            sample(4, [process(1, 999, cpu=999)], phase="all-output"),
        ]
        metadata = summary(2, first=1)
        metadata["phases"].append({"scenario": "managed-1", "phase": "cleanup", "first_sample": 3, "sample_count": 1})
        result = resource.summarize(metadata, rows)
        self.assertEqual(result["completed_phase_count"], 1)
        self.assertEqual(result["diagnostic_or_uncompleted_snapshot_count"], 3)
        self.assertEqual(group(result, "product_total")["metrics"]["private_bytes"]["observed_sum"]["max"], 20)
        self.assertEqual(group(result, "product_total")["cpu"]["observed_cpu_ms_lower_bound"], 20)

    def test_duplicate_identity_or_wrong_phase_fails_before_double_counting(self):
        with self.assertRaisesRegex(ValueError, "重复进程身份"):
            resource.summarize(summary(1), [sample(0, [process(1), process(1)])])
        with self.assertRaisesRegex(ValueError, "不符"):
            resource.summarize(summary(1), [sample(0, [process(1)], scenario="managed-10")])

    def test_cli_writes_only_derived_files_and_supports_empty_failed_run(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            raw_summary = json.dumps(summary(2)).encode("utf-8-sig")
            raw_samples = ("\n".join(json.dumps(row) for row in [sample(0, [process(1)]), sample(1, [process(1, cpu=1)])]) + "\n").encode("utf-8")
            (root / "summary.json").write_bytes(raw_summary)
            (root / "samples.jsonl").write_bytes(raw_samples)
            with redirect_stdout(io.StringIO()):
                self.assertEqual(resource.main([directory]), 0)
            self.assertEqual((root / "summary.json").read_bytes(), raw_summary)
            self.assertEqual((root / "samples.jsonl").read_bytes(), raw_samples)
            analysis = json.loads((root / "summary.analysis.json").read_text(encoding="utf-8"))
            self.assertEqual(analysis["completed_phase_count"], 1)
            self.assertIn("MiB", (root / "summary.table.md").read_text(encoding="utf-8"))
            self.assertIn("role / owner_role", (root / "summary.table.md").read_text(encoding="utf-8"))
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "summary.json").write_text(json.dumps({"result": "fail", "failure": "synthetic early stop", "phases": []}), encoding="utf-8")
            with redirect_stdout(io.StringIO()):
                self.assertEqual(resource.main([directory]), 0)
            analysis = json.loads((root / "summary.analysis.json").read_text(encoding="utf-8"))
            self.assertEqual(analysis["completed_phase_count"], 0)
            self.assertFalse(analysis["samples_file_present"])
            self.assertIsNone(analysis["source_sha256"]["samples.jsonl"])


if __name__ == "__main__":
    unittest.main()
