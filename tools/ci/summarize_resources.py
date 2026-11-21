#!/usr/bin/env python3
"""Summarize completed native Windows resource phases without changing raw data.

Usage: python tools/ci/summarize_resources.py DIRECTORY
Only Python's standard library is used. A summary is observational evidence, not
a performance gate, complete process accounting, or a Windows acceptance claim.
"""

from __future__ import annotations

import argparse
from collections import Counter
import hashlib
import json
import math
from pathlib import Path
import sys
import tempfile
from typing import Any


METRICS = ("private_bytes", "working_set_bytes", "handle_count", "thread_count")
DIAGNOSTIC_PHASES = {"cleanup", "ui-ready", "final"}
DEPARTURE_REASONS = {"previous_owned_not_observed"}
BACKEND_OWNERS = {"backend_server", "backend_client", "backend_process"}
PRODUCT_OWNERS = {"product_manager"} | BACKEND_OWNERS
GROUPS = (
    ("product_manager", "管理器及归属辅助进程", {"product_manager"}),
    ("backend_server", "后端服务端及归属辅助进程", {"backend_server"}),
    ("backend_client", "后端客户端及归属辅助进程", {"backend_client"}),
    ("backend_total", "后端合计", BACKEND_OWNERS),
    ("product_total", "产品合计（管理器＋后端）", PRODUCT_OWNERS),
    ("managed_fixture", "托管夹具及其后代", {"managed_fixture"}),
    ("managed_total", "托管总计（产品＋托管夹具）", PRODUCT_OWNERS | {"managed_fixture"}),
    ("direct_fixture", "直接运行夹具及其后代", {"direct_fixture"}),
    ("measurement_driver", "测量用终端驱动", {"measurement_driver"}),
    ("sampler", "采集器及其后代", {"sampler"}),
)


def numeric(value: Any) -> bool:
    """Counters are nonnegative finite numbers; bool is not a counter."""
    return isinstance(value, (int, float)) and not isinstance(value, bool) and math.isfinite(value) and value >= 0


def percentile(values: list[float], fraction: float) -> float | None:
    """Type-7 / inclusive linear interpolation, including the one-value case."""
    if not 0 <= fraction <= 1:
        raise ValueError("percentile fraction must be between zero and one")
    if not values:
        return None
    ordered = sorted(values)
    position = (len(ordered) - 1) * fraction
    lower, upper = math.floor(position), math.ceil(position)
    weight = position - lower
    return ordered[lower] * (1 - weight) + ordered[upper] * weight


def distribution(values: list[Any]) -> dict[str, Any]:
    known = [value for value in values if numeric(value)]
    return {
        "samples": len(values), "valid_samples": len(known), "missing_samples": len(values) - len(known),
        "p50": percentile(known, 0.50), "p95": percentile(known, 0.95),
        "max": max(known) if known else None,
    }


def identity(row: dict[str, Any]) -> tuple[int, str, str] | None:
    pid, created, path = (row.get(key) for key in ("process_id", "creation_filetime", "executable_path"))
    if not isinstance(pid, int) or isinstance(pid, bool) or pid <= 0:
        return None
    # FILETIME is intentionally a decimal string, never a rounded JSON float.
    if not isinstance(created, str) or not created.isascii() or not created.isdecimal():
        return None
    if not isinstance(path, str) or not path:
        return None
    return pid, created, path.replace("/", "\\").lower()


def owner(row: dict[str, Any]) -> str | None:
    value = row.get("owner_role")
    return value if isinstance(value, str) and value else None


def group_rows(sample: dict[str, Any] | None, owners: set[str | None], role: str | None = None) -> list[dict[str, Any]] | None:
    if sample is None or not isinstance(sample.get("snapshot", {}).get("processes"), list):
        return None
    return [row for row in sample["snapshot"]["processes"] if owner(row) in owners and (role is None or row.get("role") == role)]


def group_gaps(sample: dict[str, Any] | None, owners: set[str | None]) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    if sample is None:
        return [{"reason": "sample_missing", "owner_role": None}], []
    snapshot = sample.get("snapshot", {})
    gaps = list(snapshot.get("gaps") or [])
    if not isinstance(snapshot.get("processes"), list):
        gaps.append({"reason": "process_rows_missing", "owner_role": None})
    relevant = [gap for gap in gaps if gap.get("owner_role") is None or gap.get("owner_role") in owners]
    return (
        [gap for gap in relevant if gap.get("reason") not in DEPARTURE_REASONS],
        [gap for gap in relevant if gap.get("reason") in DEPARTURE_REASONS],
    )


def snapshot_totals(sample: dict[str, Any] | None, owners: set[str | None], role: str | None = None) -> dict[str, Any]:
    rows = group_rows(sample, owners, role)
    gaps, departures = group_gaps(sample, owners)
    totals: dict[str, Any] = {
        "sample_index": sample["snapshot"]["sample_index"] if sample is not None else None,
        "observed_process_count": len(rows) if rows is not None else None,
        "hard_gap_count": len(gaps), "departure_observation_count": len(departures),
        "population_complete": False, "metrics": {},
    }
    for metric in METRICS:
        known = [row[metric] for row in rows or [] if numeric(row.get(metric))]
        missing = len(rows) - len(known) if rows is not None else None
        # An empty, successfully observed group has an observed sum of zero.
        # An unreadable group or wholly absent counters remain null. A partial
        # known sum is explicitly separate from a complete-observation sum.
        observed = sum(known) if known or rows == [] and not gaps else None
        complete = observed if rows is not None and missing == 0 and not gaps else None
        totals["metrics"][metric] = {
            "observed_sum": observed, "complete_observation_sum": complete,
            "known_process_values": len(known), "missing_process_values": missing,
        }
    return totals


def cpu_interval(previous: dict[str, Any] | None, current: dict[str, Any] | None,
                 owners: set[str | None], role: str | None = None) -> dict[str, Any]:
    before, after = group_rows(previous, owners, role), group_rows(current, owners, role)
    previous_gaps, _ = group_gaps(previous, owners)
    current_gaps, _ = group_gaps(current, owners)
    result: dict[str, Any] = {
        "from_sample": previous["snapshot"]["sample_index"] if previous is not None else None,
        "to_sample": current["snapshot"]["sample_index"] if current is not None else None,
        "sample_wall_ms": None, "observed_cpu_ms": None, "one_core_percent_observed": None,
        "complete_for_observed_identities": False, "matched_identity_pairs": 0,
        "unpaired_previous_rows": len(before or []), "unpaired_current_rows": len(after or []),
        "invalid_pair_reasons": {}, "first_observed_elapsed_ms": None, "last_observed_elapsed_ms": None,
    }
    if previous is not None and current is not None:
        t0, t1 = (sample["snapshot"].get("sample_elapsed_ms") for sample in (previous, current))
        if numeric(t0) and numeric(t1) and t1 > t0:
            result["sample_wall_ms"] = t1 - t0
    if before is None or after is None:
        result["invalid_pair_reasons"] = {"sample_or_process_rows_missing": 1}
        return result
    before_by_id = {identity(row): row for row in before if identity(row) is not None}
    after_by_id = {identity(row): row for row in after if identity(row) is not None}
    shared = before_by_id.keys() & after_by_id.keys()
    invalid: Counter[str] = Counter()
    invalid_identities = len(before) + len(after) - len(before_by_id) - len(after_by_id)
    if invalid_identities:
        invalid["invalid_process_identity"] += invalid_identities
    deltas, rates, starts, ends = [], [], [], []
    for key in shared:
        old, new = before_by_id[key], after_by_id[key]
        c0, c1 = old.get("cpu_total_ms"), new.get("cpu_total_ms")
        t0, t1 = old.get("observed_elapsed_ms"), new.get("observed_elapsed_ms")
        if not all(numeric(value) for value in (c0, c1, t0, t1)):
            invalid["cpu_counter_or_time_missing"] += 1
        elif c1 < c0:
            invalid["cpu_counter_decreased"] += 1
        elif t1 <= t0:
            invalid["observation_time_not_increasing"] += 1
        else:
            deltas.append(c1 - c0)
            rates.append((c1 - c0) / (t1 - t0) * 100)
            starts.append(t0)
            ends.append(t1)
    result["matched_identity_pairs"] = len(deltas)
    result["unpaired_previous_rows"] = len(before) - len(shared)
    result["unpaired_current_rows"] = len(after) - len(shared)
    result["invalid_pair_reasons"] = dict(sorted(invalid.items()))
    if deltas:
        result["observed_cpu_ms"] = math.fsum(deltas)
        # Each process uses its own actual counter-observation interval. This
        # is an observed group rate, not an atomic whole-machine CPU percentage.
        result["one_core_percent_observed"] = math.fsum(rates)
        result["first_observed_elapsed_ms"] = min(starts)
        result["last_observed_elapsed_ms"] = max(ends)
    elif before == [] and after == [] and not previous_gaps and not current_gaps and result["sample_wall_ms"] is not None:
        result["observed_cpu_ms"] = 0
        result["one_core_percent_observed"] = 0
    result["complete_for_observed_identities"] = (
        not invalid and not result["unpaired_previous_rows"] and not result["unpaired_current_rows"]
        and not previous_gaps and not current_gaps and result["sample_wall_ms"] is not None
    )
    return result


def summarize_group(group_id: str, label: str, owners: set[str | None], samples: list[dict[str, Any] | None],
                    phase_elapsed_ms: Any, role: str | None = None) -> dict[str, Any]:
    totals = [snapshot_totals(sample, owners, role) for sample in samples]
    intervals = [cpu_interval(previous, current, owners, role) for previous, current in zip(samples, samples[1:])]
    cpu_values = [item["observed_cpu_ms"] for item in intervals if numeric(item["observed_cpu_ms"])]
    cpu_ms = math.fsum(cpu_values) if cpu_values else None
    observation_starts = [item["first_observed_elapsed_ms"] for item in intervals if numeric(item["first_observed_elapsed_ms"])]
    observation_ends = [item["last_observed_elapsed_ms"] for item in intervals if numeric(item["last_observed_elapsed_ms"])]
    span = max(observation_ends) - min(observation_starts) if observation_starts else None
    # Phase elapsed_ms is the actual measured window, not a nominal N*1000.
    # Reject an inconsistent denominator instead of silently reporting a rate.
    wall_valid = numeric(phase_elapsed_ms) and phase_elapsed_ms > 0 and (span is None or phase_elapsed_ms >= span)
    unique_ids = {identity(row) for sample in samples for row in group_rows(sample, owners, role) or []}
    unique_ids.discard(None)
    return {
        "id": group_id, "label": label, "owner_roles": sorted(owners, key=lambda item: item or ""), "role": role,
        "population_complete": False, "unique_observed_process_identities": len(unique_ids),
        "observed_process_count": distribution([item["observed_process_count"] for item in totals]),
        "hard_gap_observations": sum(item["hard_gap_count"] for item in totals),
        "samples_with_hard_gaps": sum(item["hard_gap_count"] > 0 for item in totals),
        "departure_observations": sum(item["departure_observation_count"] for item in totals),
        "metrics": {
            metric: {
                "observed_sum": distribution([item["metrics"][metric]["observed_sum"] for item in totals]),
                "complete_observation_sum": distribution([item["metrics"][metric]["complete_observation_sum"] for item in totals]),
                "missing_process_values": sum(item["metrics"][metric]["missing_process_values"] or 0 for item in totals),
                "samples_with_missing_values": sum(item["metrics"][metric]["missing_process_values"] is None or item["metrics"][metric]["missing_process_values"] > 0 for item in totals),
            } for metric in METRICS
        },
        "cpu": {
            "observed_cpu_ms_lower_bound": cpu_ms,
            "phase_elapsed_ms": phase_elapsed_ms if numeric(phase_elapsed_ms) else None,
            "observation_span_ms": span,
            "one_core_percent_lower_bound_over_phase": cpu_ms / phase_elapsed_ms * 100 if cpu_ms is not None and wall_valid else None,
            "phase_denominator_valid": wall_valid,
            "interval_one_core_percent_observed": distribution([item["one_core_percent_observed"] for item in intervals]),
            "interval_one_core_percent_complete_observation": distribution([item["one_core_percent_observed"] if item["complete_for_observed_identities"] else None for item in intervals]),
            "intervals": len(intervals), "intervals_with_observed_cpu": len(cpu_values),
            "intervals_with_complete_pairing": sum(item["complete_for_observed_identities"] for item in intervals),
            "matched_identity_pairs": sum(item["matched_identity_pairs"] for item in intervals),
            "unpaired_previous_rows": sum(item["unpaired_previous_rows"] for item in intervals),
            "unpaired_current_rows": sum(item["unpaired_current_rows"] for item in intervals),
            "interval_records": intervals,
        },
        "snapshot_totals": totals,
    }


def summarize(summary: dict[str, Any], records: list[dict[str, Any]]) -> dict[str, Any]:
    indexed: dict[int, dict[str, Any]] = {}
    for record in records:
        snapshot = record.get("snapshot")
        if not isinstance(snapshot, dict):
            raise ValueError("samples.jsonl 中的记录缺少 snapshot 对象")
        index = snapshot.get("sample_index")
        if not isinstance(index, int) or isinstance(index, bool) or index < 0 or index in indexed:
            raise ValueError("sample_index 必须是唯一的非负整数")
        rows = snapshot.get("processes")
        if rows is not None and (not isinstance(rows, list) or any(not isinstance(row, dict) for row in rows)):
            raise ValueError("snapshot.processes 必须是进程对象列表或 null")
        identities = [identity(row) for row in rows or [] if identity(row) is not None]
        if len(identities) != len(set(identities)):
            raise ValueError(f"sample {index} 含重复进程身份，拒绝重复求和")
        gaps = snapshot.get("gaps")
        if gaps is not None and (not isinstance(gaps, list) or any(not isinstance(gap, dict) for gap in gaps)):
            raise ValueError("snapshot.gaps 必须是对象列表或 null")
        indexed[index] = record
    phases = summary.get("phases", [])
    if not isinstance(phases, list):
        raise ValueError("summary.json.phases 必须是已完成阶段列表")
    analysis: dict[str, Any] = {
        "schema_version": 1, "source_commit": summary.get("commit"),
        "experiment_result": summary.get("result"), "experiment_failure": summary.get("failure"),
        "logical_processors": summary.get("logical_processors"), "population_complete": False,
        "percentile_method": "inclusive_linear_interpolation: position=(n-1)*q; q=0.50/0.95",
        "snapshot_count": len(records), "completed_phase_count": 0, "excluded_phases": [], "phases": [],
        "limitations": [
            "只统计 summary.json 已登记的采样阶段；cleanup、ui-ready、final 和未完成阶段的诊断快照不进入阶段统计。",
            "先对同一次非原子快照内已归属进程求和，再计算分位数；不同进程各自的峰值不能相加。",
            "observed_sum 只含可读数值；complete_observation_sum 排除指标缺测或相关读取缺口。两者都不代表完整进程总体。",
            "空的已观测组可记为零；计数缺失、整组读取失败和无法配对的 CPU 不填零。",
            "previous_owned_not_observed 单列为离场观察；它可能是短命客户端正常退出，不自动判定整个阶段缺测。",
            "CPU 只计算相邻阶段快照中 PID、字符串 FILETIME 和规范化路径一致的计数差；不跨缺失快照、身份重用、计数下降或未知时间插值。",
            "CPU 单核百分比按每进程实际 observed_elapsed_ms 差计算，可超过 100%；累计可见 CPU 是采样下界，漏掉短命客户端与进程尾段。",
            "Private Bytes 是私有提交，不是物理驻留；Working Set 含共享页，跨进程求和可能重复计数。",
            "采集器和测量终端驱动单列；共享 WMI provider 成本和测量干扰没有被完整捕获，也没有被扣除。",
            "直接夹具输出到排空管道，托管夹具输出经 ConPTY，两者不是等价终端路径，不能简单相减为产品净开销。",
            "短时基线不是内存预算、泄漏排除、长稳、真实 agent、桌面交互或发布验收。",
        ],
    }
    seen_phases, used_samples = set(), set()
    for phase in phases:
        if not isinstance(phase, dict):
            raise ValueError("summary.json.phases 的每项必须是对象")
        scenario, name = phase.get("scenario"), phase.get("phase")
        if not isinstance(scenario, str) or not isinstance(name, str):
            raise ValueError("已完成阶段缺少 scenario/phase 字符串")
        if name in DIAGNOSTIC_PHASES:
            analysis["excluded_phases"].append({"scenario": scenario, "phase": name, "reason": "diagnostic_phase"})
            continue
        if (scenario, name) in seen_phases:
            raise ValueError("同一 scenario/phase 重复登记")
        seen_phases.add((scenario, name))
        first, count = phase.get("first_sample"), phase.get("sample_count")
        if any(not isinstance(value, int) or isinstance(value, bool) for value in (first, count)) or first < 0 or count < 1:
            raise ValueError("已完成阶段需要非负 first_sample 和正整数 sample_count")
        selected, missing = [], []
        for index in range(first, first + count):
            if index in used_samples:
                raise ValueError("已完成阶段的 sample 范围重叠")
            used_samples.add(index)
            record = indexed.get(index)
            if record is not None and (record.get("scenario") != scenario or record.get("phase") != name or record.get("session_count") != phase.get("session_count")):
                raise ValueError(f"sample {index} 的 scenario/phase/session_count 与阶段登记不符")
            selected.append(record)
            if record is None:
                missing.append(index)
        gap_counts: Counter[tuple[str, str | None]] = Counter()
        for sample in selected:
            if sample is None:
                gap_counts[("sample_missing", None)] += 1
            else:
                for gap in sample["snapshot"].get("gaps") or []:
                    gap_counts[(str(gap.get("reason", "unknown_gap")), gap.get("owner_role"))] += 1
        definitions = [(group_id, label, owners, None) for group_id, label, owners in GROUPS]
        role_owners = sorted({(str(row.get("role", "unknown")), owner(row)) for sample in selected if sample is not None for row in sample["snapshot"].get("processes") or []}, key=lambda pair: (pair[0], pair[1] or ""))
        elapsed = phase.get("elapsed_ms")
        groups = [summarize_group(group_id, label, owners, selected, elapsed, role) for group_id, label, owners, role in definitions]
        details = [summarize_group(f"{role}/{owned or 'unknown'}", f"{role} / {owned or '未知归属'}", {owned}, selected, elapsed, role) for role, owned in role_owners]
        analysis["phases"].append({
            "scenario": scenario, "phase": name, "session_count": phase.get("session_count"),
            "first_sample": first, "expected_samples": count, "observed_samples": count - len(missing),
            "missing_sample_indices": missing, "elapsed_ms": elapsed, "warmup_seconds": phase.get("warmup_seconds"),
            "selected_session": phase.get("selected_session"), "expected_managers": phase.get("expected_managers"),
            "collection_duration_ms": distribution([sample["snapshot"].get("sample_duration_ms") if sample else None for sample in selected]),
            "cim_query_duration_ms": distribution([sample["snapshot"].get("cim_query_duration_ms") if sample else None for sample in selected]),
            "gap_counts": [{"reason": reason, "owner_role": owned, "count": number, "departure_observation": reason in DEPARTURE_REASONS} for (reason, owned), number in sorted(gap_counts.items(), key=lambda item: (item[0][0], item[0][1] or ""))],
            "groups": groups, "role_owner_groups": details,
        })
    analysis["completed_phase_count"] = len(analysis["phases"])
    analysis["diagnostic_or_uncompleted_snapshot_count"] = len(indexed.keys() - used_samples)
    return analysis


def display(value: Any, scale: float = 1) -> str:
    return f"{value / scale:.2f}" if numeric(value) else "—"


def triplet(stats: dict[str, Any], scale: float = 1) -> str:
    return " / ".join(display(stats[key], scale) for key in ("p50", "p95", "max"))


def cell(value: Any) -> str:
    return str(value if value is not None else "未知").replace("|", "\\|").replace("\r", " ").replace("\n", " ")


def render_markdown(analysis: dict[str, Any]) -> str:
    lines = [
        "# Windows 短时资源基线：已完成阶段的观测汇总", "",
        f"测试提交：`{cell(analysis['source_commit'])}`。原实验结果：**{cell(analysis['experiment_result'])}**。登记采样阶段 **{analysis['completed_phase_count']}** 个。", "",
        "本表不设通过阈值。原始文件为 `summary.json` 和 `samples.jsonl`；生成文件不修改原始证据。", "",
        "## 统计口径", "",
        "每个单元格依次为 **p50 / p95 / max**。分位数采用位置 `(n−1)×q` 的线性插值；先在同一采样内求组和，再统计各采样的组和。MiB = 1,048,576 bytes。进程数、句柄数和线程数同样先求每采样总量。", "",
        "表中数值是 **可见部分的求和**，不是完整进程总体；指标存在缺测时不补零。覆盖列为“该指标无缺测且无相关读取缺口的采样数 / 登记采样数”；原始 JSON 同时保留完整观测子集统计与每次采样的缺测数。`—` 表示没有可计算数值。", "",
        "CPU 百分比以一个逻辑核满载为 100%，可以大于 100%；按同一进程实际观测时间差计算，再把同一相邻采样区间内可配对进程相加。CPU 表不能代表完整进程生命周期。", "",
    ]
    if analysis.get("experiment_failure"):
        lines += [f"原实验失败信息：{cell(analysis['experiment_failure'])}", ""]
    if not analysis["phases"]:
        lines += ["没有已完成采样阶段可供汇总；诊断快照没有被计作资源阶段。", ""]
    for title, field in (("按用途分组", "groups"), ("按实际 role / owner_role 分组", "role_owner_groups")):
        lines += [f"## {title}", "", "管理器/后端组包含 owner_role 相同的 `console_aux` 等已归属后代；托管总计只包含产品组和托管夹具，驱动、采集器及直接运行夹具单列。不同组会有明确包含关系，不能再把这些行相加。", "",
                  "| 场景 / 阶段 | 组 | 可见进程数 | Private Bytes MiB | Working Set MiB | 句柄数 | 线程数 | CPU 单核 % | 覆盖 PB / WS / 句柄 / 线程 / CPU | 硬缺口 / 离场观察 |",
                  "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
        for phase in analysis["phases"]:
            for group in phase[field]:
                coverage = [f"{group['metrics'][metric]['complete_observation_sum']['valid_samples']}/{phase['expected_samples']}" for metric in METRICS]
                cpu = group["cpu"]
                coverage.append(f"{cpu['intervals_with_complete_pairing']}/{cpu['intervals']}")
                values = [f"{cell(phase['scenario'])} / {cell(phase['phase'])}", cell(group["label"]), triplet(group["observed_process_count"])]
                values += [triplet(group["metrics"][metric]["observed_sum"], 1024 * 1024 if metric.endswith("_bytes") else 1) for metric in METRICS]
                values += [triplet(cpu["interval_one_core_percent_observed"]), " / ".join(coverage), f"{group['hard_gap_observations']} / {group['departure_observations']}"]
                lines.append("| " + " | ".join(values) + " |")
        lines.append("")
    lines += ["## CPU 可见累计量与实际采样时长", "",
              "“阶段单核下界”= 可见累计 CPU 毫秒 / 原实验记录的实际阶段时长 × 100。它不补齐未成对进程的 CPU；分母无效时保留空值。“观测跨度”是有效进程配对覆盖的最早至最晚时间，可能短于阶段时长。", "",
              "| 场景 / 阶段 | 组 | 可见 CPU ms 下界 | 实际阶段 ms | 观测跨度 ms | 阶段单核 % 下界 | 有效身份配对 | 前 / 后端点未配对进程条目 |",
              "| --- | --- | --- | --- | --- | --- | --- | --- |"]
    for phase in analysis["phases"]:
        for group in phase["groups"]:
            cpu = group["cpu"]
            lines.append("| " + " | ".join([
                f"{cell(phase['scenario'])} / {cell(phase['phase'])}", cell(group["label"]),
                display(cpu["observed_cpu_ms_lower_bound"]), display(cpu["phase_elapsed_ms"]), display(cpu["observation_span_ms"]),
                display(cpu["one_core_percent_lower_bound_over_phase"]), str(cpu["matched_identity_pairs"]),
                f"{cpu['unpaired_previous_rows']} / {cpu['unpaired_current_rows']}",
            ]) + " |")
    lines += ["", "## 采样覆盖与缺口", "",
              f"共有 {analysis['snapshot_count']} 次原始快照，其中 {analysis['diagnostic_or_uncompleted_snapshot_count']} 次诊断或未完成阶段快照未纳入阶段统计。", "",
              "| 场景 / 阶段 | 实际 / 登记快照 | 单次采集 ms（p50 / p95 / max） | 缺口类型 / 归属 / 次数 |",
              "| --- | --- | --- | --- |"]
    for phase in analysis["phases"]:
        gaps = "；".join(f"{cell(gap['reason'])} / {cell(gap['owner_role'])} / {gap['count']}" + ("（离场观察）" if gap["departure_observation"] else "") for gap in phase["gap_counts"]) or "未记录"
        lines.append(f"| {cell(phase['scenario'])} / {cell(phase['phase'])} | {phase['observed_samples']} / {phase['expected_samples']} | {triplet(phase['collection_duration_ms'])} | {gaps} |")
    lines += ["", "## 限制", ""]
    lines += [f"- {item}" for item in analysis["limitations"]]
    return "\n".join(lines) + "\n"


def read_json(data: bytes, label: str) -> Any:
    def bad_constant(value: str) -> None:
        raise ValueError(f"{label} 含非法非有限数值 {value}")
    return json.loads(data.decode("utf-8-sig"), parse_constant=bad_constant)


def write_atomic(path: Path, text: str) -> None:
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", newline="\n", dir=path.parent, prefix=path.name + ".", suffix=".tmp", delete=False) as stream:
            temporary = Path(stream.name)
            stream.write(text)
        temporary.replace(path)
    finally:
        if temporary is not None:
            temporary.unlink(missing_ok=True)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="汇总已完成 Windows 资源采样阶段；保留缺测，不改原始证据。")
    parser.add_argument("directory", type=Path)
    args = parser.parse_args(argv)
    try:
        summary_path, samples_path = args.directory / "summary.json", args.directory / "samples.jsonl"
        summary_bytes = summary_path.read_bytes()
        summary = read_json(summary_bytes, "summary.json")
        if not isinstance(summary, dict):
            raise ValueError("summary.json 根值必须是对象")
        samples_bytes = samples_path.read_bytes() if samples_path.exists() else b""
        records = [read_json(line, f"samples.jsonl 第 {index} 行") for index, line in enumerate(samples_bytes.splitlines(), 1) if line.strip()]
        if any(not isinstance(record, dict) for record in records):
            raise ValueError("samples.jsonl 每行必须是对象")
        analysis = summarize(summary, records)
        analysis["source_sha256"] = {
            "summary.json": hashlib.sha256(summary_bytes).hexdigest(),
            "samples.jsonl": hashlib.sha256(samples_bytes).hexdigest() if samples_path.exists() else None,
        }
        analysis["samples_file_present"] = samples_path.exists()
        write_atomic(args.directory / "summary.analysis.json", json.dumps(analysis, ensure_ascii=False, indent=2, allow_nan=False) + "\n")
        write_atomic(args.directory / "summary.table.md", render_markdown(analysis))
    except (OSError, ValueError, TypeError, KeyError) as error:
        print(f"资源汇总失败：{error}", file=sys.stderr)
        return 2
    print(f"已汇总 {analysis['completed_phase_count']} 个登记采样阶段：summary.analysis.json、summary.table.md")
    return 0


def cli() -> int:
    # Redirected Python output on Windows can default to a legacy code page.
    # Configure the standalone CLI before argparse emits help/errors, so UTF-8
    # evidence files are not followed by a UnicodeEncodeError in the status log.
    # Imported main()/summarize() callers keep ownership of their own streams.
    for stream in (sys.stdout, sys.stderr):
        reconfigure = getattr(stream, "reconfigure", None)
        if callable(reconfigure):
            reconfigure(encoding="utf-8")
    return main()


if __name__ == "__main__":
    raise SystemExit(cli())
