"""Read go test -bench output; expose only static names and iteration counts."""
import re
import sys

groups = {
    "APIRequest": "GetDashboardStats GetCases GetClients LoginRequest",
    "CacheMiddleware": "CacheHit CacheMiss",
    "DatabaseQueries": "SimpleQuery PaginatedQuery ComplexQuery",
    "RedisOperations": "SET GET Pipeline",
    "MiddlewarePerformance": "PerformanceMiddleware RateLimitMiddleware CORSMiddleware",
    "ConcurrentRequests": "Concurrency_1 Concurrency_10 Concurrency_50 Concurrency_100",
}
required = {f"Benchmark{group}/{name}" for group, names in groups.items() for name in names.split()}
counts = {}
failed = False
pending = None
with open(sys.argv[1], encoding="utf-8", errors="replace") as output:
    for line in output:
        # Go prints the name before calibrating. Application logs can separate it from its numeric row.
        label = re.match(r"^(Benchmark[\w/]+?)(?:-\d+)?(?:\s+|$)", line)
        if label:
            pending = label[1] if label[1] in required else None
        elif re.match(r"^(?:Benchmark|goos:|pkg:|ok\s|PASS\b|FAIL\b|\s*--- (?:FAIL|SKIP):)", line):
            pending = None
        metrics = line[label.end():] if label else line
        result = re.fullmatch(r"\s*([1-9]\d*)\s+[\d.]+\s+ns/op(?:\s+[\d.]+\s+[^\s]+)*\s*", metrics)
        if pending and result:
            counts[pending] = int(result[1])
            pending = None
        failure = re.match(r"^\s*--- FAIL: ([\w/]+)(?: |$)", line)
        if not failed and failure and failure[1] in required:
            print("BENCH_FAILURE", failure[1])
            failed = True
        package = re.match(r"^FAIL\s+(law-oa-go(?:/[A-Za-z0-9._-]+)*)\s", line)
        if not failed and package:
            print("BENCH_PACKAGE_FAILURE", package[1])
            failed = True
for name, iterations in sorted(counts.items()):
    print("BENCH_RESULT", name, iterations)
missing = sorted(required - counts.keys())
if missing:
    print("BENCH_MISSING", *missing)
if failed or missing:
    sys.exit(1)
