#!/usr/bin/env bash
# Exercises volcano_job_control.sh. Needs Linux /proc and setsid:
#   bash .gitcode/workflows/scripts/volcano_job_control_test.sh
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
compile="${script_dir}/build_job_compile.sh"
helper="${script_dir}/volcano_job_control.sh"

if ! command -v setsid >/dev/null 2>&1; then
    echo "setsid is required" >&2
    exit 1
fi
if [ ! -d /proc ]; then
    echo "/proc is required" >&2
    exit 1
fi

fail() {
    echo "FAIL: $*" >&2
    exit 1
}

# The compile script must keep cleanup on EXIT and drop it before the success report.
grep -q 'source "${ATOMGIT_WORKSPACE}/${servicename}/.gitcode/workflows/scripts/volcano_job_control.sh"' "${compile}" || fail "compile script does not source the helper"
grep -q "trap 'stop_volcano_sessions' EXIT" "${compile}" || fail "missing EXIT trap"
awk '
    /trap - EXIT INT TERM/ { cleared = NR }
    /report_volcano_jobs 0/ && cleared && NR > cleared { found = 1 }
    END { exit found ? 0 : 1 }
' "${compile}" || fail "EXIT trap is still set while printing the success report"

# Volcano builds take GOROOT and PYTHON_HOME from the environment, then from go and python3.
awk '
    /grep -q "component\/ascend-for-volcano"/ { in_volcano = 1 }
    in_volcano && /grep -q "component\/ascend-operator"/ { in_volcano = 0 }
    in_volcano && /GOROOT= go env GOROOT/ { got_go = 1 }
    in_volcano && /dirname "\$\{py\}"/ { got_py = 1 }
    in_volcano && /GOROOT="\$\{GOROOT\}" PYTHON_HOME="\$\{PYTHON_HOME\}"/ { got_pass = 1 }
    in_volcano && /\/opt\/buildtools\/go_/ { hardcoded = 1 }
    in_volcano && /python3\.7\.5/ { hardcoded = 1 }
    END { exit (got_go && got_py && got_pass && !hardcoded) ? 0 : 1 }
' "${compile}" || fail "volcano toolchain is not taken from the environment"

# utils.sh is stored with CR. Source it only after that CR is stripped.
awk '
    /grep -q "component\/ascend-for-volcano"/ { in_volcano = 1 }
    in_volcano && /grep -q "component\/ascend-operator"/ { in_volcano = 0 }
    in_volcano && /dos2unix "\$\{utils_sh\}"/ && !unix_at { unix_at = NR }
    in_volcano && /source "\$\{utils_sh\}"/ && !source_at { source_at = NR }
    END { exit (unix_at && source_at && unix_at < source_at) ? 0 : 1 }
' "${compile}" || fail "utils.sh is sourced before CR is stripped"

# shellcheck source=volcano_job_control.sh
source "${helper}"

state=$(volcano_pid_state "$$")
[ "${state}" != "gone" ] || fail "could not read this shell from /proc"

work=$(mktemp -d)
trap 'rm -rf "${work}"' EXIT
ATOMGIT_WORKSPACE="${work}"
mkdir -p "${ATOMGIT_WORKSPACE}/volcano-build"
volcano_vers=(v1.9.0)
volcano_timeout=20m
volcano_pids=()
volcano_started=()
volcano_collected=()
volcano_status=()
volcano_elapsed=()
export ATOMGIT_WORKSPACE volcano_timeout

# A reaped launcher must not hide the status file.
bash -c "printf '%s\n' 9 > \"${ATOMGIT_WORKSPACE}/volcano-build/v1.9.0.status\"; exit 0" &
volcano_pids[0]=$!
volcano_started[0]=$(date +%s)
volcano_collected[0]=0
wait "${volcano_pids[0]}" || true
collect_finished_volcano_jobs
[ "${volcano_status[0]}" = "9" ] || fail "status file was not kept, got ${volcano_status[0]:-empty}"
report_volcano_jobs 0 >/dev/null
[ "${fail}" = "1" ] || fail "exit 9 was treated as success"

# Do not signal the shell that owns the builds.
sleep 30 &
kept=$!
printf '%s\n' "$$" > "${ATOMGIT_WORKSPACE}/volcano-build/v1.9.0.pgid"
volcano_pids[0]=${kept}
volcano_collected[0]=0
stop_volcano_jobs TERM
if ! kill -0 "${kept}" 2>/dev/null; then
    fail "stop_volcano_jobs signaled the parent shell"
fi
kill "${kept}" 2>/dev/null || true
wait "${kept}" 2>/dev/null || true

# A set -e exit must stop the setsid session, including its grandchild.
child_file="${work}/child"
inner="${work}/inner.sh"
cat > "${inner}" <<'EOF'
#!/usr/bin/env bash
set -e
set +m
source "$1"
ATOMGIT_WORKSPACE=$2
child_file=$3
mkdir -p "${ATOMGIT_WORKSPACE}/volcano-build"
volcano_vers=(v1.9.0)
volcano_timeout=20m
volcano_pids=()
volcano_started=()
volcano_collected=()
volcano_status=()
volcano_elapsed=()
volcano_started[0]=$(date +%s)
volcano_collected[0]=0
export ATOMGIT_WORKSPACE volcano_timeout
setsid bash -c "printf '%s\n' \"\$\$\" > \"${ATOMGIT_WORKSPACE}/volcano-build/v1.9.0.pgid\"; sleep 60 & printf '%s\n' \"\$!\" > \"${child_file}\"; wait" \
    > "${ATOMGIT_WORKSPACE}/volcano-build/v1.9.0.log" 2>&1 </dev/null &
volcano_pids[0]=$!
for _ in 1 2 3 4 5 6 7 8 9 10; do
    [ -f "${child_file}" ] && break
    sleep 0.2
done
trap 'stop_volcano_sessions' EXIT
false
EOF
set +e
bash "${inner}" "${helper}" "${work}/exit-case" "${child_file}" >"${work}/exit.out" 2>"${work}/exit.err"
inner_status=$?
set -e
[ "${inner_status}" -ne 0 ] || fail "inner script should have failed"
for _ in 1 2 3 4 5; do
    [ -f "${child_file}" ] && break
    sleep 0.2
done
[ -f "${child_file}" ] || fail "grandchild pid was not recorded"
grandchild=$(tr -d '[:space:]' < "${child_file}")
if kill -0 "${grandchild}" 2>/dev/null; then
    kill -KILL -- "-$(tr -d '[:space:]' < "${work}/exit-case/volcano-build/v1.9.0.pgid")" 2>/dev/null || true
    fail "grandchild ${grandchild} still alive after EXIT cleanup"
fi

echo "PASS"
