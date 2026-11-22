"""Production source size guardrail, independent of the current file lengths."""
import importlib.util
from pathlib import Path
import tempfile
import unittest


class SourceLimitsTest(unittest.TestCase):
    def test_limit_and_test_exemption(self):
        spec = importlib.util.spec_from_file_location(
            "source_limits", Path(__file__).with_name("check_source_limits.py"))
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "ok.go").write_text("package sample\n" * 500)
            (root / "large_test.go").write_text("// test\n" * 800)
            self.assertEqual(module.violations(root), [])
            (root / "large.go").write_text("package sample\n" * 501)
            self.assertEqual(module.violations(root), [("large.go", 501)])

    def test_scripts_share_the_production_limit(self):
        spec = importlib.util.spec_from_file_location("source_limits", Path(__file__).with_name("check_source_limits.py"))
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "tool.py").write_text("# source\n" * 501)
            (root / "tool.ps1").write_text("# source\n" * 501)
            self.assertEqual(module.violations(root), [("tool.ps1", 501), ("tool.py", 501)])


if __name__ == "__main__":
    unittest.main()
