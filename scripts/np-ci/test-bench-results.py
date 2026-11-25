"""Regression checks for package ownership and Go's split benchmark rows."""
import contextlib
import io
import runpy
import unittest
from pathlib import Path

checker = runpy.run_path(str(Path(__file__).with_name("bench-results.py")))
target = "law-oa-go/tests/performance"
other = "law-oa-go/other"
names = sorted(checker["required"])
row = "     2  2.0 ns/op 16 B/op 1 allocs/op\n"


def package(name, results):
    return f"pkg: {name}\n" + results + f"PASS\nok  {name}  1s\n"


class BenchmarkOutputTest(unittest.TestCase):
    def test_package_ownership_and_calibration(self):
        prefix = "".join(name + "-2 " + row for name in names[:-1])
        split = names[-1] + "-2\tprivate application output\n" + row
        for output, valid in [
            (package(target, prefix + split), True),
            (package(target, prefix) + package(other, split), False),
            (package(other, prefix + split), False),
            ("pkg: " + target + "\n" + prefix + split, False),
            (package(target, prefix + split.replace(row, "BenchmarkOther-2 " + row)), False),
            (package(target, prefix + split.replace("2  2.0", "0  2.0")), False),
            (package(target, prefix + split) + "FAIL  law-oa-go/other  1s\n", False),
        ]:
            with self.subTest(valid=valid), contextlib.redirect_stdout(io.StringIO()) as evidence:
                self.assertEqual(not checker["verify"](output.splitlines(keepends=True)), valid)
                self.assertNotIn("private application output", evidence.getvalue())


if __name__ == "__main__":
    unittest.main()
