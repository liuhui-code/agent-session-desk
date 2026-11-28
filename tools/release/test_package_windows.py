import json
import tempfile
import unittest
import zipfile
from pathlib import Path

from package_windows import extract_terminal


class PackageInputTests(unittest.TestCase):
    def test_terminal_archive_cannot_write_outside_fresh_host(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            archive = root / "bad.zip"
            with zipfile.ZipFile(archive, "w") as z:
                z.writestr("terminal/../outside.exe", b"fixture")
            with self.assertRaises(ValueError):
                extract_terminal(archive, root / "host", None)
            self.assertFalse((root / "outside.exe").exists())

    def test_terminal_distribution_keeps_official_notices_and_creates_no_runtime_data(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            archive = root / "official-fixture.zip"
            with zipfile.ZipFile(archive, "w") as z:
                z.writestr("terminal/WindowsTerminal.exe", b"fixture exe")
                z.writestr("terminal/NOTICE.html", b"fixture notice")
            extract_terminal(archive, root / "host", None)
            self.assertEqual((root / "host/NOTICE.html").read_bytes(), b"fixture notice")
            self.assertFalse((root / "host/settings").exists())


if __name__ == "__main__":
    unittest.main()
