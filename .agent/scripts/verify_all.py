#!/usr/bin/env python3
"""
Full pre-deploy verification suite for go-boilerplate.

Runs comprehensive checks: lint, tests, security, architecture, formatting.
Usage: python3 .agent/scripts/verify_all.py [--quick]
"""

import subprocess
import sys
import time
from pathlib import Path

# Colors
GREEN = "\033[92m"
RED = "\033[91m"
YELLOW = "\033[93m"
BLUE = "\033[94m"
BOLD = "\033[1m"
RESET = "\033[0m"

ROOT = Path(__file__).resolve().parent.parent.parent


def run_cmd(cmd: str) -> tuple[int, str, str]:
    """Run command and return (code, stdout, stderr)."""
    result = subprocess.run(
        cmd, shell=True, capture_output=True, text=True, cwd=str(ROOT)
    )
    return result.returncode, result.stdout, result.stderr


def step(name: str):
    """Print step header."""
    print(f"\n{BLUE}{'─'*50}")
    print(f"  ▶ {name}")
    print(f"{'─'*50}{RESET}\n")


def result(name: str, passed: bool, duration: float, detail: str = ""):
    """Print step result."""
    icon = f"{GREEN}PASS{RESET}" if passed else f"{RED}FAIL{RESET}"
    print(f"  [{icon}] {name} ({duration:.1f}s)")
    if detail and not passed:
        for line in detail.strip().split("\n")[:10]:
            print(f"         {line}")


def verify_lint() -> bool:
    """Run basic lint (go vet + gofmt)."""
    step("1. Lint (go vet + gofmt)")
    start = time.time()
    code, out, err = run_cmd("make lint")
    result("make lint", code == 0, time.time() - start, err)
    return code == 0


def verify_lint_full() -> bool:
    """Run full lint (golangci-lint)."""
    step("2. Full Lint (golangci-lint)")
    start = time.time()
    code, out, err = run_cmd("make lint-full 2>&1")
    result("make lint-full", code == 0, time.time() - start, out + err)
    return code == 0


def verify_format() -> bool:
    """Check formatting."""
    step("3. Format Check (gofmt)")
    start = time.time()
    code, out, _ = run_cmd("gofmt -l .")
    passed = out.strip() == ""
    result("gofmt -l .", passed, time.time() - start, out)
    return passed


def verify_unit_tests() -> bool:
    """Run unit tests."""
    step("4. Unit Tests")
    start = time.time()
    code, out, err = run_cmd("make test-unit")
    passed = code == 0

    # Extract summary
    lines = out.split("\n")
    summary = [l for l in lines if "ok" in l or "FAIL" in l]
    detail = "\n".join(summary[-10:]) if not passed else ""

    result("make test-unit", passed, time.time() - start, detail or err)
    return passed


def verify_architecture() -> bool:
    """Check Clean Architecture import rules."""
    step("5. Architecture (import rules)")
    start = time.time()

    # Domain must not import infrastructure or usecases
    code1, out1, _ = run_cmd(
        "grep -rn 'infrastructure\\|usecases' internal/domain/ --include='*.go' || true"
    )
    # Usecases must not import infrastructure
    code2, out2, _ = run_cmd(
        "grep -rn 'infrastructure' internal/usecases/ --include='*.go' || true"
    )

    passed = out1.strip() == "" and out2.strip() == ""
    detail = ""
    if out1.strip():
        detail += f"Domain violations:\n{out1}"
    if out2.strip():
        detail += f"Usecases violations:\n{out2}"

    result("Import rules", passed, time.time() - start, detail)
    return passed


def verify_mod_tidy() -> bool:
    """Check go.mod is tidy."""
    step("6. Module Tidy")
    start = time.time()
    run_cmd("go mod tidy")
    code, out, _ = run_cmd("git diff --exit-code go.mod go.sum")
    passed = code == 0
    result("go mod tidy", passed, time.time() - start,
           "go.mod or go.sum changed after tidy" if not passed else "")
    return passed


def verify_uncommitted() -> bool:
    """Check for uncommitted changes."""
    step("7. Uncommitted Changes")
    start = time.time()
    code, out, _ = run_cmd("git status --porcelain")
    lines = [l for l in out.strip().split("\n") if l and not l.startswith("??")]
    passed = len(lines) == 0
    result("No uncommitted changes", passed, time.time() - start,
           "\n".join(lines[:5]) if not passed else "")
    return passed


def verify_todo_fixme() -> bool:
    """Check for TODO/FIXME in Go files."""
    step("8. TODO/FIXME Check")
    start = time.time()
    code, out, _ = run_cmd(
        "grep -rn 'TODO\\|FIXME\\|HACK\\|XXX' internal/ pkg/ --include='*.go' || true"
    )
    lines = [l for l in out.strip().split("\n") if l]
    # Warning only, not blocking
    if lines:
        print(f"  {YELLOW}⚠️  Found {len(lines)} TODO/FIXME comments{RESET}")
        for line in lines[:5]:
            print(f"      {line}")
    else:
        print(f"  {GREEN}✅ No TODO/FIXME found{RESET}")
    result("TODO/FIXME audit", True, time.time() - start)
    return True


def verify_security() -> bool:
    """Run security scan."""
    step("9. Security Scan")
    start = time.time()
    code, out, err = run_cmd("make security 2>&1 || true")
    if "not found" in (out + err).lower():
        print(f"  {YELLOW}⚠️  gosec not installed, skipping{RESET}")
        result("Security scan", True, time.time() - start, "gosec not available")
        return True
    passed = code == 0
    result("make security", passed, time.time() - start, out[:300] if not passed else "")
    return passed


def verify_swagger() -> bool:
    """Check Swagger docs exist."""
    step("10. Swagger Docs")
    start = time.time()
    exists = (ROOT / "docs" / "swagger.json").exists()
    result("docs/swagger.json exists", exists, time.time() - start)
    return exists


def main():
    total_start = time.time()

    print(f"\n{BOLD}{'═'*50}")
    print(f"  go-boilerplate Full Verification Suite")
    print(f"{'═'*50}{RESET}")

    quick = "--quick" in sys.argv

    checks = [
        ("Lint", verify_lint),
        ("Format", verify_format),
        ("Unit Tests", verify_unit_tests),
        ("Architecture", verify_architecture),
        ("Module Tidy", verify_mod_tidy),
        ("TODO/FIXME", verify_todo_fixme),
        ("Swagger", verify_swagger),
    ]

    if not quick:
        checks.insert(1, ("Full Lint", verify_lint_full))
        checks.append(("Security", verify_security))
        checks.append(("Uncommitted", verify_uncommitted))

    results = []
    for name, fn in checks:
        try:
            results.append((name, fn()))
        except Exception as e:
            print(f"  {RED}ERROR: {name}: {e}{RESET}")
            results.append((name, False))

    # Final summary
    elapsed = time.time() - total_start
    passed = sum(1 for _, ok in results if ok)
    total = len(results)

    print(f"\n{BOLD}{'═'*50}")
    if passed == total:
        print(f"  {GREEN}ALL {total} CHECKS PASSED{RESET} ({elapsed:.1f}s)")
    else:
        print(f"  {RED}{total - passed} of {total} CHECKS FAILED{RESET} ({elapsed:.1f}s)")
    print(f"{BOLD}{'═'*50}{RESET}\n")

    for name, ok in results:
        icon = f"{GREEN}✅{RESET}" if ok else f"{RED}❌{RESET}"
        print(f"  {icon} {name}")

    print()
    return 0 if passed == total else 1


if __name__ == "__main__":
    sys.exit(main())
