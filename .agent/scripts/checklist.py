#!/usr/bin/env python3
"""
Quality Gate Checklist for go-boilerplate.

Runs automated checks to validate code quality before commits.
Usage: python3 .agent/scripts/checklist.py [--fix]
"""

import subprocess
import sys
import os
import re
from pathlib import Path

# Colors for terminal output
GREEN = "\033[92m"
RED = "\033[91m"
YELLOW = "\033[93m"
BOLD = "\033[1m"
RESET = "\033[0m"

# Project root
ROOT = Path(__file__).resolve().parent.parent.parent


def run_cmd(cmd: str, cwd: str = None) -> tuple[int, str, str]:
    """Run a shell command and return (returncode, stdout, stderr)."""
    result = subprocess.run(
        cmd, shell=True, capture_output=True, text=True,
        cwd=cwd or str(ROOT)
    )
    return result.returncode, result.stdout, result.stderr


def print_check(name: str, passed: bool, detail: str = ""):
    """Print check result with color."""
    icon = f"{GREEN}✅{RESET}" if passed else f"{RED}❌{RESET}"
    print(f"  {icon} {name}")
    if detail and not passed:
        for line in detail.strip().split("\n")[:5]:
            print(f"      {YELLOW}{line}{RESET}")


def check_clean_architecture() -> bool:
    """Verify no prohibited imports across layers."""
    print(f"\n{BOLD}1. Clean Architecture Imports{RESET}")
    passed = True

    # Domain must not import usecases or infrastructure
    code, out, _ = run_cmd(
        "grep -rn 'infrastructure\\|usecases' internal/domain/ --include='*.go' || true"
    )
    domain_clean = out.strip() == ""
    print_check("Domain has no prohibited imports", domain_clean, out)
    if not domain_clean:
        passed = False

    # Usecases must not import infrastructure
    code, out, _ = run_cmd(
        "grep -rn 'infrastructure' internal/usecases/ --include='*.go' || true"
    )
    usecases_clean = out.strip() == ""
    print_check("Use Cases has no prohibited imports", usecases_clean, out)
    if not usecases_clean:
        passed = False

    return passed


def check_error_shadowing() -> bool:
    """Check for potential error variable shadowing."""
    print(f"\n{BOLD}2. Error Variable Shadowing{RESET}")

    code, out, _ = run_cmd(
        r"grep -rn 'if err :=' internal/ --include='*.go' || true"
    )
    lines = [l for l in out.strip().split("\n") if l]

    # Group by file to find multiple err := in same file
    file_counts: dict[str, int] = {}
    for line in lines:
        fname = line.split(":")[0]
        file_counts[fname] = file_counts.get(fname, 0) + 1

    shadowed_files = {f: c for f, c in file_counts.items() if c > 2}
    passed = len(shadowed_files) == 0

    if shadowed_files:
        detail = "\n".join(f"{f}: {c} occurrences" for f, c in shadowed_files.items())
        print_check("No error shadowing detected", False, detail)
    else:
        print_check("No error shadowing detected", True)

    return passed


def check_lint() -> bool:
    """Run make lint."""
    print(f"\n{BOLD}3. Lint (go vet + gofmt){RESET}")

    code, out, err = run_cmd("make lint")
    passed = code == 0
    print_check("make lint", passed, err if not passed else "")
    return passed


def check_tests() -> bool:
    """Run unit tests."""
    print(f"\n{BOLD}4. Unit Tests{RESET}")

    code, out, err = run_cmd("make test-unit")
    passed = code == 0

    if passed:
        # Count tests
        test_lines = [l for l in out.split("\n") if "PASS" in l or "ok" in l]
        print_check(f"make test-unit", True)
    else:
        fail_lines = [l for l in (out + err).split("\n") if "FAIL" in l]
        print_check("make test-unit", False, "\n".join(fail_lines[:5]))

    return passed


def check_migrations() -> bool:
    """Verify migration files are valid."""
    print(f"\n{BOLD}5. Migrations{RESET}")

    migration_dir = ROOT / "internal" / "infrastructure" / "db" / "postgres" / "migration"
    if not migration_dir.exists():
        print_check("Migration directory exists", False, str(migration_dir))
        return False

    sql_files = sorted(migration_dir.glob("*.sql"))
    if not sql_files:
        print_check("Migration files exist", True)
        return True

    # Check naming convention
    valid = True
    for f in sql_files:
        if not re.match(r"\d+_\w+\.(sql)", f.name):
            print_check(f"Migration naming: {f.name}", False, "Expected: NNNN_name.sql")
            valid = False

    if valid:
        print_check(f"Migrations valid ({len(sql_files)} files)", True)

    return valid


def check_swagger() -> bool:
    """Check if Swagger docs exist."""
    print(f"\n{BOLD}6. Swagger Docs{RESET}")

    swagger_json = ROOT / "docs" / "swagger.json"
    exists = swagger_json.exists()
    print_check("docs/swagger.json exists", exists)

    if exists:
        # Check if docs might be stale
        code, out, _ = run_cmd("swag init -g cmd/api/main.go -o /tmp/swag-check --parseDependency --parseInternal 2>/dev/null && diff -q docs/swagger.json /tmp/swag-check/swagger.json 2>/dev/null || echo 'STALE'")
        stale = "STALE" in out
        if stale:
            print_check("Swagger docs up to date", False, "Run: swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal")
        else:
            print_check("Swagger docs up to date", True)
        return exists and not stale

    return exists


def check_security() -> bool:
    """Run security checks."""
    print(f"\n{BOLD}7. Security{RESET}")

    code, out, err = run_cmd("make security 2>&1 || true")
    # Check if gosec is available
    if "gosec" in (out + err).lower() and "not found" in (out + err).lower():
        print_check("gosec available", False, "Install: go install github.com/securego/gosec/v2/cmd/gosec@latest")
        return True  # Non-blocking

    passed = code == 0 or "no issues" in out.lower()
    print_check("Security scan (gosec)", passed, out[:200] if not passed else "")
    return passed


def main():
    """Run all checks and report summary."""
    print(f"\n{BOLD}{'='*50}")
    print(f"  go-boilerplate Quality Gate Checklist")
    print(f"{'='*50}{RESET}\n")

    fix_mode = "--fix" in sys.argv

    checks = [
        ("Clean Architecture", check_clean_architecture),
        ("Error Shadowing", check_error_shadowing),
        ("Lint", check_lint),
        ("Unit Tests", check_tests),
        ("Migrations", check_migrations),
        ("Swagger", check_swagger),
        ("Security", check_security),
    ]

    results = []
    for name, check_fn in checks:
        try:
            results.append((name, check_fn()))
        except Exception as e:
            print_check(name, False, str(e))
            results.append((name, False))

    # Summary
    passed = sum(1 for _, ok in results if ok)
    total = len(results)

    print(f"\n{BOLD}{'='*50}")
    print(f"  Summary: {passed}/{total} checks passed")
    print(f"{'='*50}{RESET}\n")

    for name, ok in results:
        icon = f"{GREEN}✅{RESET}" if ok else f"{RED}❌{RESET}"
        print(f"  {icon} {name}")

    print()
    return 0 if passed == total else 1


if __name__ == "__main__":
    sys.exit(main())
