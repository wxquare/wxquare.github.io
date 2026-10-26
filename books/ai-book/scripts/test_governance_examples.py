"""Execute governance teaching functions from Markdown, not copied test doubles."""

import ast
from decimal import Decimal, localcontext
from math import comb, isclose, log, sqrt
from pathlib import Path
import re
import unittest


BOOK = Path(__file__).resolve().parents[1]
GOVERNANCE = BOOK / "src/part3/06-agent-governance-optimization.md"
STATISTICS = BOOK / "src/part2/04-data-eval.md"


def extract(path, function_name):
    blocks = re.findall(r"^```python\n(.*?)^```\s*$", path.read_text(), re.M | re.S)
    matches = []
    for block in blocks:
        tree = ast.parse(block)
        if any(isinstance(node, ast.FunctionDef) and node.name == function_name
               for node in tree.body):
            matches.append(block)
    if len(matches) != 1:
        raise AssertionError(f"{path}: expected one block defining {function_name}")
    namespace = {}
    exec(compile(matches[0], str(path), "exec"), namespace)
    return namespace


GATE = extract(GOVERNANCE, "release_gate")["release_gate"]
STATS = extract(STATISTICS, "paired_binary_interval")
PAIRED = STATS["paired_binary_interval"]
ZERO = STATS["zero_event_bound"]
CRITERIA = extract(GOVERNANCE, "evaluate_criteria")["evaluate_criteria"]


def independent_cell_bounds(k, n, tail):
    """Independent Decimal oracle: direct binomial polynomial, no log-gamma."""
    with localcontext() as ctx:
        ctx.prec = 60
        tail = Decimal(str(tail))

        def root(last, target):
            low, high = Decimal(0), Decimal(1)
            for _ in range(160):
                p = (low + high) / 2
                cdf = sum(Decimal(comb(n, i)) * p ** i * (1 - p) ** (n - i)
                          for i in range(last + 1))
                if cdf > target:
                    low = p
                else:
                    high = p
            return (low + high) / 2

        lower = Decimal(0) if k == 0 else root(k - 1, 1 - tail)
        upper = Decimal(1) if k == n else root(k, tail)
        return float(lower), float(upper)


def fixture():
    policy = dict(
        candidate="incident-agent-v43", baseline="incident-agent-v42",
        slice="incident-readonly", policy_version="prod-policy-v5",
        eval_suite="incident-regression:v17", evaluator_version="incident-evaluator:v3",
        quality_alpha=0.025, safety_alpha=0.025, total_alpha=0.05,
        allowed_drop=0.02, max_zero_event_upper=0.01,
        min_n=320, online_min_n=200, online_window_age=600,
        max_age={"eval": 3600, "failure": 60, "policy": 60, "online": 60})
    payloads = {
        "eval": dict(
            eval_suite=policy["eval_suite"], evaluator_version=policy["evaluator_version"],
            smoke="passed", missing_regression=False,
            required_repair_ids=["eval-incident-0099"],
            repair_results={"eval-incident-0099": "passed"},
            critical_unknown=0, unknown=0, invalid=0, unresolved_new_failures=0,
            paired=PAIRED([(1, 1)] * 320, alpha=0.025),
            safety_bound=ZERO(500, alpha=0.025), cost_over=False, latency_over=False),
        "failure": dict(open_p0=0, affected_p1=0),
        "policy": dict(phase="offline", coverage_complete=True,
                       safety_violations=0, approval_bypass=0, critical_unknown=0),
        "online": dict(window_start=800, window_end=990, mature_n=200,
                       critical_unknown=0, safety_violations=0, stop_triggered=False,
                       cost_over=False, latency_over=False),
    }
    sources = {"eval": "eval-runner", "failure": "failure-registry",
               "policy": "policy-auditor", "online": "online-metrics"}
    bundle = {}
    for kind, data in payloads.items():
        bundle[kind] = dict(
            schema="governance-signal/v1", source=sources[kind],
            candidate=policy["candidate"], baseline=policy["baseline"],
            slice=policy["slice"], policy_version=policy["policy_version"],
            produced_at=995, expires_at=1100, data=data)
    return bundle, policy


class ReleaseGateTests(unittest.TestCase):
    def setUp(self):
        self.bundle, self.policy = fixture()

    def decide(self, expected, stage="offline", now=1000):
        result, reason = GATE(self.bundle, self.policy, stage=stage, now=now)
        self.assertEqual(expected, result, reason)

    def test_offline_allow_without_online(self):
        del self.bundle["online"]
        self.decide("allow")

    def test_scale_allow_requires_canary_policy(self):
        self.bundle["policy"]["data"]["phase"] = "canary"
        self.decide("allow", "scale")

    def test_stage_requirements(self):
        self.decide("block", "scale")
        self.bundle["policy"]["data"]["phase"] = "canary"
        del self.bundle["online"]
        self.decide("block", "scale")
        self.decide("block", "unknown")

    def test_decimal_alpha_budget_exact_allocation_allows(self):
        self.policy.update(quality_alpha=.025, safety_alpha=.005, total_alpha=.03)
        self.bundle["eval"]["data"]["safety_bound"] = ZERO(1000, .005)
        self.decide("allow")

    def test_decimal_alpha_budget_small_remainder_allows(self):
        self.policy.update(quality_alpha=.025, safety_alpha=.005,
                           total_alpha=.03000000000000001)
        self.bundle["eval"]["data"]["safety_bound"] = ZERO(1000, .005)
        self.decide("allow")

    def test_decimal_alpha_budget_small_excess_blocks(self):
        self.policy.update(quality_alpha=.025, safety_alpha=.005,
                           total_alpha=.029999999999999995)
        self.bundle["eval"]["data"]["safety_bound"] = ZERO(1000, .005)
        self.assertEqual(GATE(self.bundle, self.policy, stage="offline", now=1000),
                         ("block", "joint error budget exceeded"))

    def test_decimal_alpha_budget_does_not_round_away_tiny_excess(self):
        self.policy.update(quality_alpha=.025, safety_alpha=5e-324, total_alpha=.025)
        self.bundle["eval"]["data"]["safety_bound"] = ZERO(1000, 5e-324)
        self.assertEqual(GATE(self.bundle, self.policy, stage="offline", now=1000),
                         ("block", "joint error budget exceeded"))

    def test_all_required_signals_missing(self):
        for stage in ("offline", "scale"):
            for kind in ("eval", "failure", "policy", "online"):
                if kind == "online" and stage == "offline":
                    continue
                with self.subTest(stage=stage, kind=kind):
                    self.bundle, self.policy = fixture()
                    self.bundle["policy"]["data"]["phase"] = "canary"
                    del self.bundle[kind]
                    self.decide("block", stage)

    def test_every_envelope_and_payload_field_required(self):
        for kind in fixture()[0]:
            signal = fixture()[0][kind]
            for field in signal:
                with self.subTest(kind=kind, envelope=field):
                    self.bundle, self.policy = fixture()
                    del self.bundle[kind][field]
                    self.decide("block")
            for field in signal["data"]:
                with self.subTest(kind=kind, payload=field):
                    self.bundle, self.policy = fixture()
                    del self.bundle[kind]["data"][field]
                    self.decide("block")

    def test_every_statistical_field_required(self):
        for name in ("paired", "safety_bound"):
            for key in fixture()[0]["eval"]["data"][name]:
                with self.subTest(report=name, field=key):
                    self.bundle, self.policy = fixture()
                    del self.bundle["eval"]["data"][name][key]
                    self.decide("block")

    def test_all_versions_and_sources(self):
        for kind in fixture()[0]:
            for field in ("candidate", "baseline", "slice", "policy_version",
                          "schema", "source"):
                with self.subTest(kind=kind, field=field):
                    self.bundle, self.policy = fixture()
                    self.bundle[kind][field] = "wrong-version-or-source"
                    self.decide("block")
        for field in ("eval_suite", "evaluator_version"):
            self.bundle, self.policy = fixture()
            self.bundle["eval"]["data"][field] = "wrong-version"
            self.decide("block")

    def test_signal_time_boundaries_and_max_age(self):
        for kind in fixture()[0]:
            for field, value in (("produced_at", 1001), ("expires_at", 1000),
                                 ("produced_at", -10000), ("expires_at", float("inf"))):
                with self.subTest(kind=kind, field=field, value=value):
                    self.bundle, self.policy = fixture()
                    self.bundle[kind][field] = value
                    self.decide("block")
        self.bundle, self.policy = fixture()
        self.bundle["policy"]["produced_at"] = 940
        self.decide("allow")
        self.bundle["policy"]["produced_at"] = 939
        self.decide("block")

    def test_online_window_validation_even_if_offline(self):
        for updates in (
            dict(window_start=990), dict(window_end=1001),
            dict(window_start=0, window_end=100), dict(window_end=float("nan")),
        ):
            with self.subTest(updates=updates):
                self.bundle, self.policy = fixture()
                self.bundle["online"]["data"].update(updates)
                self.decide("block")

    def test_policy_and_failure_hard_blocks(self):
        for kind, field in (
            ("failure", "open_p0"), ("failure", "affected_p1"),
            ("policy", "safety_violations"), ("policy", "approval_bypass"),
            ("policy", "critical_unknown"), ("eval", "critical_unknown"),
            ("online", "critical_unknown"), ("online", "safety_violations"),
        ):
            with self.subTest(kind=kind, field=field):
                self.bundle, self.policy = fixture()
                self.bundle[kind]["data"][field] = 1
                self.decide("block")
        self.bundle, self.policy = fixture()
        self.bundle["policy"]["data"]["coverage_complete"] = False
        self.decide("block")
        self.bundle, self.policy = fixture()
        self.bundle["online"]["data"]["stop_triggered"] = True
        self.bundle["eval"]["data"]["unknown"] = 1
        self.decide("block")

    def test_smoke_and_repairs_identity_not_just_count(self):
        for state in ("failed", "unknown", "invalid"):
            self.bundle, self.policy = fixture()
            self.bundle["eval"]["data"]["smoke"] = state
            self.decide("block")
        for repairs in ({}, {"another-id": "passed"},
                        {"eval-incident-0099": "unknown"}):
            self.bundle, self.policy = fixture()
            self.bundle["eval"]["data"]["repair_results"] = repairs
            self.decide("block")
        self.bundle, self.policy = fixture()
        self.bundle["eval"]["data"].update(required_repair_ids=[], repair_results={})
        self.decide("allow")

    def test_review_branches(self):
        for field in ("unknown", "invalid", "unresolved_new_failures"):
            with self.subTest(field=field):
                self.bundle, self.policy = fixture()
                self.bundle["eval"]["data"][field] = 1
                self.decide("review")
        for kind in ("eval", "online"):
            for field in ("cost_over", "latency_over"):
                self.bundle, self.policy = fixture()
                self.bundle[kind]["data"][field] = True
                self.decide("review")
        self.bundle, self.policy = fixture()
        self.policy["min_n"] = 321
        self.decide("review")
        self.bundle, self.policy = fixture()
        self.policy["allowed_drop"] = 0.001
        self.decide("review")
        self.bundle, self.policy = fixture()
        self.policy["max_zero_event_upper"] = 0.001
        self.decide("review")
        self.bundle, self.policy = fixture()
        self.bundle["policy"]["data"]["phase"] = "canary"
        self.bundle["online"]["data"]["mature_n"] = 199
        self.decide("review", "scale")

    def test_schema_types_do_not_coerce(self):
        for kind, field, value in (
            ("failure", "open_p0", False), ("failure", "affected_p1", -1),
            ("policy", "coverage_complete", 1), ("policy", "phase", []),
            ("eval", "unknown", "0"), ("eval", "smoke", "completed"),
            ("eval", "required_repair_ids", ["x", "x"]),
            ("online", "cost_over", "false"), ("online", "mature_n", 200.5),
        ):
            with self.subTest(kind=kind, field=field):
                self.bundle, self.policy = fixture()
                self.bundle[kind]["data"][field] = value
                self.decide("block")
        for field in fixture()[1]:
            with self.subTest(policy=field):
                self.bundle, self.policy = fixture()
                del self.policy[field]
                self.decide("block")
        self.bundle, self.policy = fixture()
        self.policy["total_alpha"] = 0.04
        self.decide("block")
        self.decide("block", now=float("nan"))

    def test_statistical_protocol_and_internal_consistency(self):
        for name, key, value in (
            ("paired", "method", "hoeffding"), ("paired", "sides", "one-sided"),
            ("paired", "alpha", 0.05), ("paired", "unit", "model_attempt"),
            ("paired", "low", float("nan")), ("paired", "low", 0.1),
            ("paired", "gain", 321), ("paired", "delta", 0.1),
            ("paired", "n", True), ("safety_bound", "n", 0),
            ("safety_bound", "sides", "two-sided"), ("safety_bound", "alpha", 0.05),
            ("safety_bound", "method", "unknown"), ("safety_bound", "upper", 0),
        ):
            with self.subTest(report=name, key=key, value=value):
                self.bundle, self.policy = fixture()
                self.bundle["eval"]["data"][name][key] = value
                self.decide("block")


class StatisticsTests(unittest.TestCase):
    def test_cells_against_independent_decimal_oracle(self):
        for n, k, tail in ((20, 4, .05), (320, 0, .00625),
                           (320, 2, .00625), (320, 5, .00625), (20, 20, .05)):
            with self.subTest(n=n, k=k, tail=tail):
                observed = STATS["_cell_bounds"](k, n, tail)
                expected = independent_cell_bounds(k, n, tail)
                for a, b in zip(observed, expected):
                    self.assertAlmostEqual(a, b, places=11)
        low, high = STATS["_cell_bounds"](4, 20, .05)
        self.assertAlmostEqual(low, .071354, places=6)
        self.assertAlmostEqual(high, .401028117331975, places=12)

    def test_320_paired_cases_and_noninferiority(self):
        pairs = [(1, 1)] * 280 + [(0, 0)] * 33 + [(1, 0)] * 5 + [(0, 1)] * 2
        result = PAIRED(pairs, .025)
        gain_lo, gain_hi = independent_cell_bounds(2, 320, .025 / 4)
        loss_lo, loss_hi = independent_cell_bounds(5, 320, .025 / 4)
        self.assertEqual((result["loss"], result["gain"]), (5, 2))
        self.assertEqual(result["delta"], -3 / 320)
        self.assertAlmostEqual(result["low"], gain_lo - loss_hi, places=11)
        self.assertAlmostEqual(result["high"], gain_hi - loss_lo, places=11)
        self.assertLess(result["low"], -.02)
        self.assertEqual(result["sides"], "two-sided")
        reversed_result = PAIRED([(b, a) for a, b in pairs], .025)
        self.assertAlmostEqual(reversed_result["low"], -result["high"], places=11)
        self.assertAlmostEqual(reversed_result["high"], -result["low"], places=11)

    def test_zero_flip_not_zero_width_and_hoeffding_reference(self):
        result = PAIRED([(1, 1)] * 320, .05)
        expected = 1 - (.05 / 4) ** (1 / 320)
        self.assertAlmostEqual(result["low"], -expected, places=11)
        self.assertAlmostEqual(result["high"], expected, places=11)
        self.assertAlmostEqual(sqrt(2 * log(2 / .05) / 320), .1518403655, places=9)
        self.assertTrue(isclose(sqrt(2 * log(1 / .025) / 320),
                                sqrt(2 * log(2 / .05) / 320)))

    def test_zero_event_upper_independent_decimal_identity(self):
        for n, alpha in ((100, .05), (320, .025), (500, .025), (1, .05)):
            with self.subTest(n=n, alpha=alpha):
                result = ZERO(n, alpha)
                with localcontext() as ctx:
                    ctx.prec = 60
                    expected = 1 - (Decimal(str(alpha)).ln() / n).exp()
                self.assertAlmostEqual(result["upper"], float(expected), places=14)
                self.assertAlmostEqual((1 - result["upper"]) ** n, alpha, places=12)
                self.assertEqual(result["sides"], "one-sided")

    def test_invalid_statistical_inputs(self):
        for pairs, alpha in (([], .05), ([(2, 0)], .05), ([(1, 1)], 0),
                             ([(1, 1)], float("nan"))):
            with self.assertRaises(ValueError):
                PAIRED(pairs, alpha)
        for n, alpha in ((0, .05), (-1, .05), (True, .05), (1.5, .05), (10, 1), (10, 0)):
            with self.assertRaises(ValueError):
                ZERO(n, alpha)

    def test_exact_mcnemar_does_not_establish_noninferiority(self):
        # n10=5, n01=2: two-sided exact binomial p, not a CI or an NI test.
        p = min(1, 2 * sum(comb(7, i) for i in range(3)) / 2 ** 7)
        self.assertEqual(p, .453125)


class CriteriaTests(unittest.TestCase):
    def test_four_states_and_config_precedence(self):
        for criteria, expected in (
            ({"a": "passed"}, "passed"), ({"a": "unknown"}, "unknown"),
            ({"a": "failed", "b": "unknown"}, "failed"),
            ({"a": "invalid", "b": "failed"}, "invalid"), ({}, "invalid"),
            ({"a": "completed"}, "invalid"),
        ):
            self.assertEqual(CRITERIA(criteria), expected)
        self.assertEqual(CRITERIA({"a": "passed"}, False), "invalid")


if __name__ == "__main__":
    unittest.main(verbosity=2)
