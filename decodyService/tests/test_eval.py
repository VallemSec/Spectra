import unittest

from helpers import safe_eval


class EvalTest(unittest.TestCase):
    def test_eval(self):
        self.assertTrue(safe_eval("1 == 1"))

    def test_invalid_eval(self):
        self.assertFalse(safe_eval("1 == 2"))

    def test_raise_exception_eval(self):
        with self.assertRaises(ValueError):
            safe_eval("1 + 3")
