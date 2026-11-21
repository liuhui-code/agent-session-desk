#!/usr/bin/env python3
"""Validate and render test-design documents; never execute product tests."""

import argparse
import csv
import io
import json
import math
import re
import sys
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TESTING = Path("docs/testing")
REQUIRED = {
    "id", "title", "features", "requirements", "adrs", "priority", "stage",
    "kinds", "level", "status", "execution", "spec_state", "decision_refs",
    "environments", "fixtures", "preconditions", "steps", "expected", "evidence",
}
OPTIONAL = {"parameters", "notes"}
ENUMS = {
    "priority": {"P0", "P1", "P2"},
    "stage": {"M0", "M1", "M2", "M3", "M4", "POST_MVP"},
    "level": {"unit", "contract", "integration", "windows_e2e", "manual", "soak", "release"},
    "status": {"designed"},
    "execution": {"not_run"},
    "spec_state": {"ready", "needs_decision", "deferred"},
}
KINDS = {
    "functional", "boundary", "negative", "state", "concurrency", "recovery",
    "compatibility", "usability", "stability", "performance", "security",
    "release", "property", "fuzz",
}
LIST_FIELDS = (
    "features", "requirements", "adrs", "kinds", "decision_refs", "environments",
    "fixtures", "preconditions", "steps", "expected", "evidence",
)
EMPTY_ALLOWED = {"requirements", "decision_refs"}
LABELS = {
    "ready": "依据明确",
    "needs_decision": "待决策",
    "deferred": "后置",
}
SUITE_LABELS = {
    "launch-terminal-ui": "启动、终端、UI、焦点与 Windows",
    "sessions-recovery-storage": "会话、后台、恢复与存储",
    "output-messaging": "输出、投递、队列、控制 CLI 与回复",
    "resilience-resources-release": "资源、长稳、控制边界、发布与属性测试",
    "attention-notifications": "人工介入、待处理事项、通知与后台监测",
}


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON key: " + key)
        result[key] = value
    return result


def read_json(path):
    def reject_nonfinite(value):
        raise ValueError("nonfinite JSON number: " + value)
    return json.loads(
        path.read_text(encoding="utf-8"),
        object_pairs_hook=unique_object,
        parse_constant=reject_nonfinite,
    )


def source_ids(root):
    ledger = (root / "docs/FEATURE_LEDGER.md").read_text(encoding="utf-8")
    plan = (root / "docs/PRODUCT_PLAN.md").read_text(encoding="utf-8")
    decisions = (root / "docs/DECISIONS.md").read_text(encoding="utf-8")
    features = set(re.findall(r"^### (FL-\d{3})", ledger, re.M))
    requirements = set(re.findall(r"^\| (FR-\d{2}) \|", plan, re.M))
    open_ids = set(re.findall(r"^\| (O-\d{3}) \|", decisions, re.M))
    adrs = {
        "ADR-" + path.name[:4]: path
        for path in sorted((root / "docs/adr").glob("[0-9][0-9][0-9][0-9]-*.md"))
    }
    return features, requirements, open_ids, adrs


def validate(root):
    """Return (suites, registry, ADR paths, errors); do not mutate input files."""
    errors = []
    suites = []
    seen = set()
    try:
        features, requirements, open_ids, adrs = source_ids(root)
        registry = read_json(root / TESTING / "catalog-registry.json")
    except (OSError, ValueError, TypeError) as error:
        return [], {}, {}, [str(error)]
    if not isinstance(registry, dict) or registry.get("schema_version") != 1:
        return [], {}, {}, ["catalog registry must be an object with schema_version=1"]
    for group in ("environments", "fixtures", "decisions", "budgets"):
        if not isinstance(registry.get(group), dict):
            errors.append("registry missing object: " + group)
    if errors:
        return [], registry, adrs, errors

    valid_decisions = open_ids | set(registry["decisions"])
    for group in ("environments", "fixtures"):
        for key, entry in registry[group].items():
            if not isinstance(entry, dict) or not isinstance(entry.get("path"), str):
                errors.append("registry path missing: " + key)
                continue
            target = entry["path"].split("#", 1)[0]
            if not (root / TESTING / target).is_file():
                errors.append("registry target does not exist: " + key + " -> " + target)
    for key, entry in registry["decisions"].items():
        if not key.startswith("TBD-") or not isinstance(entry, dict):
            errors.append("invalid decision registry entry: " + key)
            continue
        for name in ("title", "question", "stage", "status"):
            if not isinstance(entry.get(name), str) or not entry[name].strip():
                errors.append(key + " missing " + name)
        if entry.get("status") not in ("open", "resolved"):
            errors.append(key + ": decision status must be open or resolved")
        if entry.get("status") == "resolved" and (
            not isinstance(entry.get("resolution"), str) or not entry["resolution"].strip()
        ):
            errors.append(key + ": resolved decision requires a resolution reference")
        related = entry.get("related")
        if not isinstance(related, list) or any(not isinstance(x, str) or x not in open_ids for x in related):
            errors.append(key + ": related must reference existing O IDs")
    for key, entry in registry["budgets"].items():
        if not re.fullmatch(r"B_[A-Z0-9_]+", key) or not isinstance(entry, dict):
            errors.append("invalid budget registry entry: " + key)
            continue
        for name in ("title", "unit", "stage"):
            if not isinstance(entry.get(name), str) or not entry[name].strip():
                errors.append(key + " missing " + name)
        if "value" not in entry:
            errors.append(key + " missing explicit value (null when undecided)")
        else:
            value = entry["value"]
            if value is not None and (
                type(value) not in (int, float)
                or value < 0
                or (isinstance(value, float) and not math.isfinite(value))
            ):
                errors.append(key + ": budget value must be null or a finite nonnegative number")
            elif value is not None and entry.get("unit", "").startswith("待"):
                errors.append(key + ": numeric budget requires a resolved unit")
        refs = entry.get("decision_refs")
        if not isinstance(refs, list) or not refs or any(not isinstance(x, str) or x not in valid_decisions for x in refs):
            errors.append(key + ": decision_refs must reference registered decisions")

    if errors:
        return [], registry, adrs, errors

    references = {
        "features": features, "requirements": requirements, "adrs": set(adrs),
        "decision_refs": valid_decisions, "environments": set(registry["environments"]),
        "fixtures": set(registry["fixtures"]),
    }
    paths = sorted((root / TESTING / "cases").glob("*.json"))
    if not paths:
        errors.append("no case sources")
    suite_names = set()
    for path in paths:
        try:
            data = read_json(path)
        except (OSError, ValueError, TypeError) as error:
            errors.append(str(path) + ": " + str(error))
            continue
        if not isinstance(data, dict) or data.get("schema_version") != 1:
            errors.append(path.name + ": expected schema_version=1 object")
            continue
        name = data.get("suite")
        if not isinstance(name, str) or name not in SUITE_LABELS or name in suite_names:
            errors.append(path.name + ": unknown or duplicate suite")
        if isinstance(name, str):
            suite_names.add(name)
        cases = data.get("cases")
        if not isinstance(cases, list) or not cases:
            errors.append(path.name + ": cases must be a nonempty list")
            continue
        suites.append((path, data))
        for index, case in enumerate(cases):
            prefix = path.name + ":" + str(index + 1)
            if not isinstance(case, dict):
                errors.append(prefix + ": case must be object")
                continue
            case_id = case.get("id")
            if not isinstance(case_id, str) or not re.fullmatch(r"TC-[A-Z][A-Z0-9]*-\d{3}", case_id):
                errors.append(prefix + ": invalid ID")
            elif case_id in seen:
                errors.append("duplicate case ID: " + case_id)
            else:
                seen.add(case_id)
                prefix = case_id
            for key in sorted(REQUIRED - set(case)):
                errors.append(prefix + ": missing " + key)
            for key in sorted(set(case) - REQUIRED - OPTIONAL):
                errors.append(prefix + ": unknown field " + key)
            if not isinstance(case.get("title"), str) or not case["title"].strip():
                errors.append(prefix + ": title must be nonempty string")
            for key, allowed in ENUMS.items():
                value = case.get(key)
                if not isinstance(value, str) or value not in allowed:
                    errors.append(prefix + ": invalid " + key)
            list_fields_valid = True
            for key in LIST_FIELDS:
                value = case.get(key)
                if not isinstance(value, list) or any(not isinstance(x, str) or not x.strip() for x in value):
                    errors.append(prefix + ": " + key + " must be a string list")
                    list_fields_valid = False
                    continue
                if not value and key not in EMPTY_ALLOWED:
                    errors.append(prefix + ": empty " + key)
                if len(value) != len(set(value)):
                    errors.append(prefix + ": duplicate item in " + key)
                if key in references:
                    for ref in value:
                        if ref not in references[key]:
                            errors.append(prefix + ": unknown " + key + " reference " + ref)
            if not list_fields_valid:
                continue
            if isinstance(case.get("kinds"), list):
                for kind in case["kinds"]:
                    if not isinstance(kind, str) or kind not in KINDS:
                        errors.append(prefix + ": invalid kind " + str(kind))
            if case.get("spec_state") == "needs_decision" and not case.get("decision_refs"):
                errors.append(prefix + ": needs_decision requires decision_refs")
            if case.get("spec_state") == "ready" and any(
                x.startswith("TBD-") and registry["decisions"].get(x, {}).get("status") != "resolved"
                for x in case.get("decision_refs", [])
            ):
                errors.append(prefix + ": ready cannot depend on an unresolved TBD")
            if case.get("spec_state") == "deferred" and case.get("stage") != "POST_MVP":
                errors.append(prefix + ": deferred case must be POST_MVP")
            if "FL-018" in case.get("features", []) and case.get("spec_state") != "deferred":
                errors.append(prefix + ": FL-018 remains deferred in this baseline")
            for budget in set(re.findall(r"\bB_[A-Z0-9_]+", json.dumps(case, ensure_ascii=False))):
                if budget not in registry["budgets"]:
                    errors.append(prefix + ": unregistered budget " + budget)
                elif registry["budgets"][budget].get("value") is None and case.get("spec_state") == "ready":
                    errors.append(prefix + ": ready case uses unresolved budget " + budget)
                elif case.get("spec_state") == "ready" and any(
                    ref.startswith("TBD-") and registry["decisions"].get(ref, {}).get("status") != "resolved"
                    for ref in registry["budgets"][budget]["decision_refs"]
                ):
                    errors.append(prefix + ": budget depends on unresolved contract " + budget)
            parameters = case.get("parameters", [])
            if not isinstance(parameters, list):
                errors.append(prefix + ": parameters must be a list")
            else:
                for parameter in parameters:
                    if (
                        not isinstance(parameter, dict)
                        or set(parameter) != {"name", "values"}
                        or not isinstance(parameter.get("name"), str)
                        or not parameter["name"].strip()
                        or not isinstance(parameter.get("values"), list)
                        or not parameter["values"]
                    ):
                        errors.append(prefix + ": invalid parameter entry")
            if "notes" in case and (not isinstance(case["notes"], str) or not case["notes"].strip()):
                errors.append(prefix + ": notes must be nonempty string")

    valid_cases = [
        case for _, data in suites for case in data["cases"]
        if isinstance(case, dict) and all(
            isinstance(case.get(field), list) and all(isinstance(x, str) for x in case[field])
            for field in LIST_FIELDS
        )
    ]
    for field, expected_ids in (("features", features), ("requirements", requirements)):
        covered = {
            ref for case in valid_cases for ref in case.get(field, [])
            if isinstance(ref, str)
        }
        missing = expected_ids - covered
        if missing:
            errors.append("missing design coverage for " + field + ": " + ", ".join(sorted(missing)))
    return suites, registry, adrs, errors


def cell(value):
    return str(value).replace("|", r"\|").replace("\n", " ")


def links(ids, target):
    return "、".join("[" + value + "](" + target(value) + ")" for value in ids) or "—"


def numbered(title, values, ordered=False):
    result = ["#### " + title, ""]
    result += [(str(i + 1) + ". " if ordered else "- ") + value for i, value in enumerate(values)]
    return result + [""]


def render_outputs(root, suites, registry, adrs):
    outputs = {}
    all_cases = [(path, case) for path, data in suites for case in data["cases"]]
    locations = {
        case["id"]: "generated/" + path.stem + ".md#" + case["id"].lower()
        for path, case in all_cases
    }
    feature_ids, requirement_ids, _, _ = source_ids(root)

    for path, data in suites:
        rows = [
            "# " + SUITE_LABELS[data["suite"]], "",
            "> 从 [JSON 源](../cases/" + path.name + ") 自动生成。请修改源并运行 §python tools/test_catalog.py render§。", "",
            "**全部为测试设计，当前未执行产品或 Windows 测试。** ready 仅表示预期依据明确。", "",
            "[总索引](../CASE_INDEX.md) · [覆盖表](../COVERAGE.md) · [字段约定](../CASE_FORMAT.md)", "",
            "| ID | 场景 | 优先级 | 阶段 | 规范状态 |",
            "| --- | --- | --- | --- | --- |",
        ]
        for case in data["cases"]:
            rows.append("| [" + case["id"] + "](#" + case["id"].lower() + ") | " +
                        cell(case["title"]) + " | " + case["priority"] + " | " +
                        case["stage"] + " | " + LABELS[case["spec_state"]] + " |")
        for case in data["cases"]:
            rows += ["", "### " + case["id"], "", "**" + case["title"] + "**", ""]
            rows += [
                "- 功能：" + links(case["features"], lambda x: "../COVERAGE.md#" + x.lower()),
                "- 需求：" + links(case["requirements"], lambda x: "../COVERAGE.md#" + x.lower()),
                "- 决策：" + links(case["adrs"], lambda x: "../../adr/" + adrs[x].name),
                "- 优先级 / 阶段 / 层次：" + " / ".join([case["priority"], case["stage"], case["level"]]),
                "- 类型：" + "、".join(case["kinds"]),
                "- 设计 / 执行：§" + case["status"] + " / " + case["execution"] + "§",
                "- 规范状态：" + LABELS[case["spec_state"]] + "（§" + case["spec_state"] + "§）",
                "- 待决依据：" + links(case["decision_refs"], lambda x: "../BUDGETS_AND_DECISIONS.md" if x.startswith("TBD-") else "../../DECISIONS.md"),
                "- 环境：" + links(case["environments"], lambda x: "../" + registry["environments"][x]["path"]),
                "- 夹具：" + links(case["fixtures"], lambda x: "../" + registry["fixtures"][x]["path"]),
                "",
            ]
            for field, title, ordered in (
                ("preconditions", "前置条件", False), ("steps", "操作步骤", True),
                ("expected", "预期与不变量", False), ("evidence", "需要的证据", False),
            ):
                rows += numbered(title, case[field], ordered)
            if case.get("parameters"):
                rows += ["#### 参数变体", ""]
                for parameter in case["parameters"]:
                    rows += ["- " + parameter["name"] + "：§" +
                             json.dumps(parameter["values"], ensure_ascii=False) + "§"]
                rows += [""]
            if case.get("notes"):
                rows += ["说明：" + case["notes"], ""]
        outputs[TESTING / "generated" / (path.stem + ".md")] = "\n".join(rows).replace("§", chr(96)).rstrip() + "\n"

    index = [
        "# 测试用例索引", "",
        "> 自动生成；JSON 为编辑源。所有用例 designed / not_run，参数变体不额外计数。", "",
        "[测试入口](README.md) · [设计覆盖](COVERAGE.md) · [CSV 索引](case-index.csv)", "",
    ]
    for path, data in suites:
        index += ["## " + SUITE_LABELS[data["suite"]], "",
                  "| ID | 场景 | FL | 优先级 | 阶段 | 层次 | 规范 |",
                  "| --- | --- | --- | --- | --- | --- | --- |"]
        for case in data["cases"]:
            index.append("| [" + case["id"] + "](" + locations[case["id"]] + ") | " +
                         " | ".join(cell(x) for x in (
                             case["title"], "、".join(case["features"]), case["priority"],
                             case["stage"], case["level"], LABELS[case["spec_state"]],
                         )) + " |")
        index += [""]
    outputs[TESTING / "CASE_INDEX.md"] = "\n".join(index).rstrip() + "\n"

    cases = [case for _, case in all_cases]
    specification = Counter(case["spec_state"] for case in cases)
    coverage = [
        "# 用例设计覆盖", "",
        "> 自动生成；这里的覆盖是设计映射，不是代码覆盖率、实际通过率或全部边界已穷尽。", "",
        "- 用例：" + str(len(cases)) + "；全部 designed / not_run。",
        "- 依据明确：" + str(specification["ready"]) + "；待决策：" +
        str(specification["needs_decision"]) + "；后置：" + str(specification["deferred"]) + "。",
        "- FL：" + str(len(feature_ids)) + "；FR：" + str(len(requirement_ids)) + "，均有设计引用。",
        "- 已记录 ADR：" + str(len(adrs)) + "；本库引用 ADR：" +
        str(len({ref for case in cases for ref in case["adrs"]})) + "。下方逐项列出数量，零引用不等于已覆盖。",
        "- 本轮设计基线执行：0；没有产品通过结论。本表不统计后续独立运行报告，实际证据见 [results/](results/)。",
        "",
        "[总索引](CASE_INDEX.md) · [测试策略](TEST_STRATEGY.md) · [未决项](BUDGETS_AND_DECISIONS.md)", "",
        "## 套件", "", "| 套件 | 数量 |", "| --- | --- |",
    ]
    for path, data in suites:
        coverage.append("| [" + SUITE_LABELS[data["suite"]] + "](generated/" + path.stem +
                        ".md) | " + str(len(data["cases"])) + " |")
    for field, label in (("priority", "优先级"), ("stage", "阶段"), ("level", "层次")):
        counts = Counter(case[field] for case in cases)
        coverage += ["", "## " + label, "", "| " + label + " | 数量 |", "| --- | --- |"]
        coverage += ["| " + key + " | " + str(value) + " |" for key, value in sorted(counts.items())]
    counts = Counter(kind for case in cases for kind in case["kinds"])
    coverage += ["", "## 类型标签", "", "同一用例可有多个标签，数量不能相加当作独立用例总数。", "",
                 "| 标签 | 数量 |", "| --- | --- |"]
    coverage += ["| " + key + " | " + str(value) + " |" for key, value in sorted(counts.items())]
    for field, ids, label in (("features", feature_ids, "Feature"), ("requirements", requirement_ids, "需求")):
        coverage += ["", "## " + label + "映射", "",
                     "| 编号 | 用例数 | 依据明确 | 待决策 | 后置 |", "| --- | --- | --- | --- | --- |"]
        for ref in sorted(ids):
            matching = [case for case in cases if ref in case[field]]
            counter = Counter(case["spec_state"] for case in matching)
            coverage.append("| [" + ref + "](#" + ref.lower() + ") | " + " | ".join(str(x) for x in (
                len(matching), counter["ready"], counter["needs_decision"], counter["deferred"],
            )) + " |")
        for ref in sorted(ids):
            matching = [case for case in cases if ref in case[field]]
            coverage += ["", "### " + ref, "",
                         links([case["id"] for case in matching], lambda x: locations[x]), ""]
    coverage += ["", "## ADR 引用", "", "| ADR | 用例数 |", "| --- | --- |"]
    for ref, path in sorted(adrs.items()):
        coverage.append("| [" + ref + "](../adr/" + path.name + ") | " +
                        str(sum(ref in case["adrs"] for case in cases)) + " |")
    coverage += ["", "FL-018 为后置探索，不能因已有两条用例就将其计入首版实现范围。", ""]
    outputs[TESTING / "COVERAGE.md"] = "\n".join(coverage).rstrip() + "\n"

    stream = io.StringIO(newline="")
    writer = csv.writer(stream, lineterminator="\n")
    writer.writerow(["id", "title", "features", "requirements", "adrs", "priority", "stage",
                     "kinds", "level", "spec_state", "status", "execution", "decision_refs",
                     "source", "readable"])
    for path, case in all_cases:
        fields = ("id", "title", "features", "requirements", "adrs", "priority", "stage",
                  "kinds", "level", "spec_state", "status", "execution", "decision_refs")
        writer.writerow([(";".join(case[x]) if isinstance(case[x], list) else case[x]) for x in fields] +
                        ["cases/" + path.name, locations[case["id"]]])
    outputs[TESTING / "case-index.csv"] = stream.getvalue()
    return outputs


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("check", "render", "summary"), nargs="?", default="check")
    parser.add_argument("--generated", action="store_true", help="also check generated files for drift")
    args = parser.parse_args(argv)
    suites, registry, adrs, errors = validate(ROOT)
    if errors:
        for error in errors:
            print("ERROR: " + error, file=sys.stderr)
        return 1
    outputs = render_outputs(ROOT, suites, registry, adrs)
    if args.generated or args.command == "render":
        expected_generated = {path for path in outputs if path.parent == TESTING / "generated"}
        existing_generated = {
            path.relative_to(ROOT) for path in (ROOT / TESTING / "generated").glob("*.md")
        }
        orphaned = sorted(existing_generated - expected_generated)
        if orphaned:
            for path in orphaned:
                print("ERROR: orphaned generated document; review before removal: " + str(path), file=sys.stderr)
            return 1
    if args.command == "render":
        for path, content in outputs.items():
            target = ROOT / path
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(content, encoding="utf-8")
        print("Rendered " + str(len(outputs)) + " files from JSON sources.")
    elif args.generated:
        drift = [str(path) for path, content in outputs.items()
                 if not (ROOT / path).is_file() or (ROOT / path).read_text(encoding="utf-8") != content]
        if drift:
            for path in drift:
                print("ERROR: generated file missing or stale: " + path, file=sys.stderr)
            return 1
    cases = [case for _, data in suites for case in data["cases"]]
    print(json.dumps({
        "cases": len(cases),
        "suites": len(suites),
        "spec_state": dict(sorted(Counter(case["spec_state"] for case in cases).items())),
        "feature_design_coverage": len({ref for case in cases for ref in case["features"]}),
        "requirement_design_coverage": len({ref for case in cases for ref in case["requirements"]}),
        "product_tests_executed_by_this_tool": 0,
    }, ensure_ascii=False))
    print("Catalog validation only; no product or Windows tests were executed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
