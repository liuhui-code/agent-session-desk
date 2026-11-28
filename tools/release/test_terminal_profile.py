import json
import tempfile
import unittest
from pathlib import Path

from terminal_profile import write_settings


class PortableTerminalTests(unittest.TestCase):
    def test_fresh_host_has_english_plain_terminal_and_no_sample_startup(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            profile = write_settings(root)
            settings = json.loads((root / "settings/settings.json").read_text())
            self.assertTrue((root / ".portable").is_file())
            self.assertEqual(settings["language"], "en-US")
            plain = next(p for p in settings["profiles"]["list"]
                         if p["guid"] == settings["defaultProfile"])
            self.assertEqual(plain["name"], "Terminal")
            self.assertEqual(plain["tabTitle"], "Terminal")
            self.assertIn("powershell.exe", plain["commandline"])
            self.assertNotIn("EncodedCommand", plain["commandline"])
            self.assertNotIn("startupActions", settings)
            self.assertNotIn("launchMode", settings)
            self.assertNotEqual(profile, settings["defaultProfile"])
            self.assertEqual(len(settings["profiles"]["list"]), 2)


if __name__ == "__main__":
    unittest.main()
