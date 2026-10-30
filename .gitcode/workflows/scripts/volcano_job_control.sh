# Helpers for the three parallel Volcano builds.
# build_job_compile.sh sources this file. volcano_job_control_test.sh exercises it.

volcano_pid_state() {
    local pid="$1"
    local stat rest
    if [ -z "${pid}" ] || [ ! -r "/proc/${pid}/stat" ]; then
        printf '%s\n' gone
        return 0
    fi
    stat=$(cat "/proc/${pid}/stat" 2>/dev/null) || stat=""
    if [ -z "${stat}" ]; then
        printf '%s\n' gone
        return 0
    fi
    rest="${stat##*) }"
    printf '%s\n' "${rest%% *}"
}

record_volcano_job() {
    local idx="$1"
    local status="$2"
    local now
    if [ "${volcano_collected[$idx]}" -eq 1 ]; then
        return 0
    fi
    status=${status//[^0-9]/}
    case "${status}" in
        ''|*[!0-9]*) status=1 ;;
    esac
    now=$(date +%s)
    volcano_collected[$idx]=1
    volcano_status[$idx]=${status}
    volcano_elapsed[$idx]=$(( now - volcano_started[$idx] ))
}

collect_finished_volcano_jobs() {
    local idx pid status status_file pgid_file pgid state gstate
    for idx in "${!volcano_pids[@]}"; do
        if [ "${volcano_collected[$idx]}" -eq 1 ]; then
            continue
        fi
        pid="${volcano_pids[$idx]}"
        status_file="${ATOMGIT_WORKSPACE}/volcano-build/${volcano_vers[$idx]}.status"
        pgid_file="${ATOMGIT_WORKSPACE}/volcano-build/${volcano_vers[$idx]}.pgid"
        pgid=""
        if [ -f "${pgid_file}" ]; then
            pgid=$(cat "${pgid_file}")
            pgid=${pgid//[^0-9]/}
        fi
        # setsid forks only when it is already a group leader. The real session then
        # differs from $!. Keep waiting while that session is alive.
        if [ -n "${pgid}" ] && [ "${pgid}" != "${pid}" ]; then
            gstate=$(volcano_pid_state "${pgid}")
            if [ ! -f "${status_file}" ] && [ "${gstate}" != "gone" ] && [ "${gstate}" != "Z" ]; then
                continue
            fi
        elif [ ! -f "${status_file}" ]; then
            state=$(volcano_pid_state "${pid}")
            if [ "${state}" != "gone" ] && [ "${state}" != "Z" ]; then
                continue
            fi
        fi
        status=0
        if [ -n "${pid}" ]; then
            wait "${pid}" 2>/dev/null || status=$?
        fi
        if [ -f "${status_file}" ]; then
            status=$(cat "${status_file}")
        elif [ -n "${pgid}" ] && [ "${pgid}" != "${pid}" ]; then
            status=1
        fi
        record_volcano_job "${idx}" "${status}"
    done
}

volcano_still_running() {
    local idx names=""
    for idx in "${!volcano_pids[@]}"; do
        if [ "${volcano_collected[$idx]}" -eq 0 ]; then
            names="${names} ${volcano_vers[$idx]}"
        fi
    done
    printf '%s' "${names# }"
}

# Signal the whole build session, including timeout, build.sh, and go.
stop_volcano_jobs() {
    local sig="$1"
    local idx pid pgid pgid_file
    for idx in "${!volcano_pids[@]}"; do
        pid="${volcano_pids[$idx]}"
        pgid_file="${ATOMGIT_WORKSPACE}/volcano-build/${volcano_vers[$idx]}.pgid"
        pgid=""
        if [ -f "${pgid_file}" ]; then
            pgid=$(cat "${pgid_file}")
            pgid=${pgid//[^0-9]/}
        fi
        if [ -z "${pgid}" ]; then
            pgid="${pid}"
        fi
        case "${pgid}" in
            ''|*[!0-9]*) continue ;;
        esac
        if [ "${pgid}" = "$$" ]; then
            echo "refuse to signal the parent shell" >&2
            continue
        fi
        kill "-${sig}" -- "-${pgid}" 2>/dev/null || true
        if [ -n "${pid}" ] && [ "${pid}" != "${pgid}" ]; then
            kill "-${sig}" -- "-${pid}" 2>/dev/null || true
        fi
    done
}

report_volcano_jobs() {
    local force_logs="$1"
    local idx ver status elapsed log
    fail=0
    for idx in "${!volcano_vers[@]}"; do
        ver="${volcano_vers[$idx]}"
        log="${ATOMGIT_WORKSPACE}/volcano-build/${ver}.log"
        if [ "${volcano_collected[$idx]}" -ne 1 ]; then
            fail=1
            echo "volcano ${ver} interrupted before it reported a status"
            if [ -f "${log}" ]; then
                echo "===== volcano ${ver} ====="
                cat "${log}"
            fi
            continue
        fi
        status="${volcano_status[$idx]}"
        elapsed="${volcano_elapsed[$idx]}"
        if [ "${status}" -eq 0 ]; then
            echo "volcano ${ver} finished in ${elapsed}s"
        elif [ "${status}" -eq 124 ] || [ "${status}" -eq 137 ]; then
            fail=1
            echo "volcano ${ver} timed out after ${volcano_timeout} (exit ${status}, ${elapsed}s)"
        elif [ "${status}" -eq 130 ] || [ "${status}" -eq 143 ]; then
            fail=1
            echo "volcano ${ver} interrupted (exit ${status}, ${elapsed}s)"
        else
            fail=1
            echo "volcano ${ver} failed with exit ${status} after ${elapsed}s"
        fi
        if [ "${force_logs}" -eq 1 ] || [ "${status}" -ne 0 ]; then
            if [ -f "${log}" ]; then
                echo "===== volcano ${ver} ====="
                cat "${log}"
            fi
        fi
    done
}

reap_volcano_jobs() {
    local attempts="$1"
    local try running
    try=0
    while [ "${try}" -lt "${attempts}" ]; do
        collect_finished_volcano_jobs
        running=$(volcano_still_running)
        if [ -z "${running}" ]; then
            return 0
        fi
        try=$(( try + 1 ))
        if [ "${try}" -lt "${attempts}" ]; then
            sleep 1
        fi
    done
}

# Drop the traps first so exit from this function cannot re-enter it.
stop_volcano_sessions() {
    trap - EXIT INT TERM
    set +x
    stop_volcano_jobs TERM
    reap_volcano_jobs 5
    if [ -n "$(volcano_still_running)" ]; then
        stop_volcano_jobs KILL
        reap_volcano_jobs 5
    fi
    report_volcano_jobs 1
}

on_volcano_interrupt() {
    stop_volcano_sessions
    exit "$1"
}
