#!/bin/bash
# Deploy container-manager as a systemd service
# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e

# ======================== Constants ========================
readonly COMPONENT_NAME="container-manager"
readonly SERVICE_NAME="${COMPONENT_NAME}"
readonly TIMER_NAME="${COMPONENT_NAME}.timer"
readonly SERVICE_DESC="Ascend container manager"

# Installation paths (aligned with installation guide)
readonly INSTALL_BIN_DIR="/usr/local/bin"
readonly INSTALL_LOG_DIR="/var/log/mindx-dl/${COMPONENT_NAME}"
readonly SYSTEMD_UNIT_DIR="/etc/systemd/system"

# Source build paths
CUR_DIR=$(dirname "$(readlink -f "$0")")
readonly BINARY_PATH="${CUR_DIR}/${COMPONENT_NAME}"

# Default runtime parameters (aligned with code defaults)
readonly DEFAULT_RUNTIME_TYPE="docker"
readonly DEFAULT_SOCK_PATH="/run/docker.sock"
readonly DEFAULT_CTR_STRATEGY="never"
readonly DEFAULT_LOG_LEVEL=0
readonly DEFAULT_LOG_FILE="${INSTALL_LOG_DIR}/${COMPONENT_NAME}.log"
readonly DEFAULT_MAX_AGE=7
readonly DEFAULT_MAX_BACKUPS=30

# Timer default
readonly DEFAULT_TIMER_DELAY="60s"

# ======================== Color Output ========================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ======================== Pre-check ========================
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

check_systemd() {
    if ! command -v systemctl &>/dev/null; then
        log_error "systemctl not found, systemd is required"
        exit 1
    fi
}

check_binary() {
    if [ ! -f "${BINARY_PATH}" ]; then
        log_error "Binary not found at ${BINARY_PATH}"
        log_error "Please place the ${COMPONENT_NAME} binary in the same directory as this script"
        exit 1
    fi
}

# Validate parameter values against code constraints
validate_log_level() {
    local level="$1"
    if ! [[ "$level" =~ ^-?[0-9]+$ ]] || [ "$level" -lt -1 ] || [ "$level" -gt 3 ]; then
        log_error "Invalid log-level: ${level}, must be in range [-1, 3]"
        exit 1
    fi
}

validate_max_age() {
    local age="$1"
    if ! [[ "$age" =~ ^[0-9]+$ ]] || [ "$age" -lt 7 ] || [ "$age" -gt 700 ]; then
        log_error "Invalid max-age: ${age}, must be in range [7, 700]"
        exit 1
    fi
}

validate_max_backups() {
    local backups="$1"
    if ! [[ "$backups" =~ ^[0-9]+$ ]] || [ "$backups" -lt 1 ] || [ "$backups" -gt 30 ]; then
        log_error "Invalid max-backups: ${backups}, must be in range (0, 30]"
        exit 1
    fi
}

validate_path_not_symlink() {
    local path="$1"
    local param_name="$2"
    if [ -L "${path}" ]; then
        log_error "Invalid ${param_name}: ${path} must not be a symlink"
        exit 1
    fi
}

validate_path_safe_for_unit() {
    local value="$1"
    local param_name="$2"
    # Reject characters that break systemd unit files or shell safety
    if [[ "${value}" =~ [[:space:]\'\"\\$\`\;\|\&\<\>] ]]; then
        log_error "Invalid ${param_name}: value contains disallowed characters"
        exit 1
    fi
    if [[ "${value}" =~ $'\n' || "${value}" =~ $'\r' ]]; then
        log_error "Invalid ${param_name}: value contains newline characters"
        exit 1
    fi
}

validate_timer_delay() {
    local delay="$1"
    # Accept formats: 60s, 2min, 1h, 30s, 90sec, 5mins etc.
    if ! [[ "${delay}" =~ ^[0-9]+(s|m|h|sec|secs|min|mins|hr|hrs)?$ ]]; then
        log_error "Invalid timerDelay: ${delay}, expected format like 60s, 2min, 1h"
        exit 1
    fi
}

validate_fault_config() {
    local path="$1"
    if [ ! -f "${path}" ]; then
        log_error "Fault config file not found: ${path}"
        exit 1
    fi
    validate_path_not_symlink "${path}" "faultConfigPath"
    local perm
    perm=$(stat -c '%a' "${path}" 2>/dev/null || echo "777")
    local perm_last3="${perm: -3}"
    if [ "${perm_last3:0:1}" -gt 6 ] || [ "${perm_last3:1:1}" -gt 4 ] || [ "${perm_last3:2:1}" -gt 0 ]; then
        log_error "Fault config file permission too open: ${perm}, must not exceed 640"
        exit 1
    fi
}

# ======================== Service Unit File ========================
generate_service_file() {
    local runtime_type="${1:-${DEFAULT_RUNTIME_TYPE}}"
    local sock_path="${2:-${DEFAULT_SOCK_PATH}}"
    local ctr_strategy="${3:-${DEFAULT_CTR_STRATEGY}}"
    local log_level="${4:-${DEFAULT_LOG_LEVEL}}"
    local max_age="${5:-${DEFAULT_MAX_AGE}}"
    local max_backups="${6:-${DEFAULT_MAX_BACKUPS}}"
    local fault_cfg_path="${7:-}"
    local log_path="${8:-${DEFAULT_LOG_FILE}}"

    local exec_args="run"
    exec_args+=" -ctrStrategy=${ctr_strategy}"
    exec_args+=" -logPath=${log_path}"
    if [ -n "${runtime_type}" ]; then
        exec_args+=" -runtimeType=${runtime_type}"
    fi
    if [ -n "${sock_path}" ]; then
        exec_args+=" -sockPath=${sock_path}"
    fi
    if [ "${log_level}" != "0" ]; then
        exec_args+=" -logLevel=${log_level}"
    fi
    if [ "${max_age}" != "7" ]; then
        exec_args+=" -maxAge=${max_age}"
    fi
    if [ "${max_backups}" != "30" ]; then
        exec_args+=" -maxBackups=${max_backups}"
    fi
    if [ -n "${fault_cfg_path}" ]; then
        exec_args+=" -faultConfigPath=${fault_cfg_path}"
    fi

    # Aligned with installation guide service config
    cat <<EOF
[Unit]
Description=${SERVICE_DESC}
Documentation=hiascend.com

[Service]
ExecStart=${INSTALL_BIN_DIR}/${COMPONENT_NAME} ${exec_args}
Restart=always
RestartSec=2
KillMode=process
Environment="GOGC=50"
Environment="GOMAXPROCS=2"
Environment="GODEBUG=madvdontneed=1"
Type=simple
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF
}

# ======================== Timer Unit File ========================
generate_timer_file() {
    local delay="${1:-${DEFAULT_TIMER_DELAY}}"

    cat <<EOF
[Unit]
Description=Timer for container manager Service

[Timer]
OnBootSec=${delay}
Unit=${SERVICE_NAME}.service

[Install]
WantedBy=timers.target
EOF
}

# ======================== Install ========================
do_install() {
    local runtime_type="${DEFAULT_RUNTIME_TYPE}"
    local sock_path="${DEFAULT_SOCK_PATH}"
    local ctr_strategy="${DEFAULT_CTR_STRATEGY}"
    local log_level="${DEFAULT_LOG_LEVEL}"
    local max_age="${DEFAULT_MAX_AGE}"
    local max_backups="${DEFAULT_MAX_BACKUPS}"
    local fault_cfg_path=""
    local log_path="${DEFAULT_LOG_FILE}"
    local timer_delay="${DEFAULT_TIMER_DELAY}"
    local skip_confirm=false

    # Parse install arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --runtimeType=*)
                runtime_type="${1#*=}"
                shift
                ;;
            --sockPath=*)
                sock_path="${1#*=}"
                shift
                ;;
            --ctrStrategy=*)
                ctr_strategy="${1#*=}"
                shift
                ;;
            --logLevel=*)
                log_level="${1#*=}"
                shift
                ;;
            --maxAge=*)
                max_age="${1#*=}"
                shift
                ;;
            --maxBackups=*)
                max_backups="${1#*=}"
                shift
                ;;
            --faultConfig=*)
                fault_cfg_path="${1#*=}"
                shift
                ;;
            --logPath=*)
                log_path="${1#*=}"
                shift
                ;;
            --timerDelay=*)
                timer_delay="${1#*=}"
                shift
                ;;
            -y|--yes)
                skip_confirm=true
                shift
                ;;
            *)
                log_error "Unknown install option: $1"
                show_install_help
                exit 1
                ;;
        esac
    done

    # Validate parameters
    if [[ "${runtime_type}" != "docker" && "${runtime_type}" != "containerd" ]]; then
        log_error "Invalid runtime-type: ${runtime_type}, must be 'docker' or 'containerd'"
        exit 1
    fi

    if [[ "${ctr_strategy}" != "never" && "${ctr_strategy}" != "singleRecover" && "${ctr_strategy}" != "ringRecover" ]]; then
        log_error "Invalid ctr-strategy: ${ctr_strategy}, must be 'never', 'singleRecover', or 'ringRecover'"
        exit 1
    fi

    validate_log_level "${log_level}"
    validate_max_age "${max_age}"
    validate_max_backups "${max_backups}"

    # Sanitize path values before writing to systemd unit
    validate_path_safe_for_unit "${sock_path}" "sockPath"
    validate_path_safe_for_unit "${log_path}" "logPath"
    if [ -n "${fault_cfg_path}" ]; then
        validate_path_safe_for_unit "${fault_cfg_path}" "faultConfig"
    fi
    validate_timer_delay "${timer_delay}"

    # If runtime is containerd and sock_path is still docker default, auto-switch
    if [[ "${runtime_type}" == "containerd" && "${sock_path}" == "${DEFAULT_SOCK_PATH}" ]]; then
        sock_path="/run/containerd/containerd.sock"
        log_info "Runtime is containerd, using default socket: ${sock_path}"
    fi

    # Validate sock path is not a symlink
    if [ -e "${sock_path}" ]; then
        validate_path_not_symlink "${sock_path}" "sockPath"
    fi

    # Validate fault config if specified
    if [ -n "${fault_cfg_path}" ]; then
        validate_fault_config "${fault_cfg_path}"
    fi

    # Resolve display names for logLevel and ctrStrategy
    local log_level_desc=""
    case "${log_level}" in
        -1) log_level_desc="debug" ;;
        0)  log_level_desc="info" ;;
        1)  log_level_desc="warning" ;;
        2)  log_level_desc="error" ;;
        3)  log_level_desc="critical" ;;
    esac

    local ctr_strategy_desc=""
    case "${ctr_strategy}" in
        never)         ctr_strategy_desc="not recover containers" ;;
        singleRecover) ctr_strategy_desc="recover single faulty chip container" ;;
        ringRecover)   ctr_strategy_desc="recover all related chip containers" ;;
    esac

    log_info "Installing ${COMPONENT_NAME}..."
    log_info "  Runtime type : ${runtime_type}"
    log_info "  Socket path  : ${sock_path}"
    log_info "  CTR strategy : ${ctr_strategy} (${ctr_strategy_desc})"
    log_info "  Log level    : ${log_level} (${log_level_desc})"
    log_info "  Log path     : ${log_path}"
    if [ -n "${fault_cfg_path}" ]; then
        log_info "  Fault config : ${fault_cfg_path}"
    fi

    # Check prerequisites
    check_binary

    if [ -f "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service" ]; then
        if [ "${skip_confirm}" = false ]; then
            log_warn "Service is already installed. Reinstalling will overwrite existing configuration."
            echo ""
            echo -e "  ${YELLOW}Do you want to continue? [y/N]${NC}"
            read -r confirm
            if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
                log_info "Installation cancelled"
                exit 0
            fi
        fi
    fi

    # Stop existing service if running
    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        log_warn "Service is running, stopping it first..."
        systemctl stop "${SERVICE_NAME}" || true
    fi

    # Create log directory
    local log_dir
    log_dir=$(dirname "${log_path}")
    log_info "Creating log directory..."
    mkdir -p "${log_dir}"
    chown root:root "${log_dir}"
    chmod 0750 "${log_dir}"

    # Install binary
    log_info "Installing binary to ${INSTALL_BIN_DIR}/"
    cp -f "${BINARY_PATH}" "${INSTALL_BIN_DIR}/${COMPONENT_NAME}"
    chmod 0500 "${INSTALL_BIN_DIR}/${COMPONENT_NAME}"

    # Generate and install systemd service unit
    log_info "Generating systemd service unit..."
    generate_service_file "${runtime_type}" "${sock_path}" "${ctr_strategy}" \
        "${log_level}" "${max_age}" "${max_backups}" "${fault_cfg_path}" "${log_path}" \
        > "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service"
    chmod 0640 "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service"

    # Generate and install timer
    log_info "Generating systemd timer (delay=${timer_delay})..."
    generate_timer_file "${timer_delay}" > "${SYSTEMD_UNIT_DIR}/${TIMER_NAME}"
    chmod 0640 "${SYSTEMD_UNIT_DIR}/${TIMER_NAME}"

    # Reload systemd
    systemctl daemon-reload

    # Enable auto-start and start service now
    systemctl enable "${SERVICE_NAME}"
    systemctl start "${SERVICE_NAME}"

    systemctl enable "${TIMER_NAME}"
    systemctl start "${TIMER_NAME}"
    log_info "Timer enabled (starts ${timer_delay} after boot)"

    log_info "Installation completed, verifying..."

    # Verify service state
    local svc_state="unknown"
    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        svc_state="${GREEN}active (running)${NC}"
    elif systemctl is-failed --quiet "${SERVICE_NAME}" 2>/dev/null; then
        svc_state="${RED}failed${NC}"
    else
        svc_state="${YELLOW}inactive${NC}"
    fi
    echo -e "  Service       : ${svc_state}"

    if systemctl is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
        echo "  Auto-start    : enabled"
    else
        echo "  Auto-start    : disabled"
    fi

    if systemctl is-active --quiet "${TIMER_NAME}" 2>/dev/null; then
        echo -e "  Timer         : ${GREEN}active${NC}"
    else
        echo -e "  Timer         : ${YELLOW}inactive${NC}"
    fi

    local version
    version=$("${INSTALL_BIN_DIR}/${COMPONENT_NAME}" -v 2>/dev/null || echo "unknown")
    echo "  Binary        : ${INSTALL_BIN_DIR}/${COMPONENT_NAME}  (${version})"

    local all_ok=true
    if ! systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        all_ok=false
    fi
    if ! systemctl is-active --quiet "${TIMER_NAME}" 2>/dev/null; then
        all_ok=false
    fi

    echo ""
    if [ "${all_ok}" = true ]; then
        echo -e "  ${GREEN}✓ All checks passed${NC}"
    else
        log_warn "Some checks failed. Check logs with: journalctl -u ${SERVICE_NAME} -f"
    fi

    echo ""
    echo "Usage:"
    echo "  systemctl status ${SERVICE_NAME}   # Check running status"
    echo "  journalctl -u ${SERVICE_NAME} -f   # View live logs"
    echo "  $(basename "$0") uninstall          # Uninstall"
}

# ======================== Uninstall ========================
do_uninstall() {
    if [ $# -gt 0 ]; then
        log_error "Uninstall does not accept any options"
        show_uninstall_help
        exit 1
    fi

    log_info "Uninstalling ${COMPONENT_NAME}..."

    local actual_log_dir="${INSTALL_LOG_DIR}"
    if [ -f "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service" ]; then
        local log_path_in_unit
        log_path_in_unit=$(sed -n 's/.*-logPath=\([^ ]*\).*/\1/p' "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service" 2>/dev/null || true)
        if [ -n "${log_path_in_unit}" ]; then
            actual_log_dir=$(dirname "${log_path_in_unit}")
        fi
    fi

    # Stop and disable service
    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        log_info "Stopping service..."
        systemctl stop "${SERVICE_NAME}" || true
    fi

    if systemctl is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
        log_info "Disabling service..."
        systemctl disable "${SERVICE_NAME}" || true
    fi

    # Stop and disable timer
    if systemctl is-active --quiet "${TIMER_NAME}" 2>/dev/null; then
        log_info "Stopping timer..."
        systemctl stop "${TIMER_NAME}" || true
    fi

    if systemctl is-enabled --quiet "${TIMER_NAME}" 2>/dev/null; then
        log_info "Disabling timer..."
        systemctl disable "${TIMER_NAME}" || true
    fi

    # Remove systemd units
    local need_reload=false
    if [ -f "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service" ]; then
        log_info "Removing systemd service unit..."
        rm -f "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service"
        need_reload=true
    fi

    if [ -f "${SYSTEMD_UNIT_DIR}/${TIMER_NAME}" ]; then
        log_info "Removing systemd timer unit..."
        rm -f "${SYSTEMD_UNIT_DIR}/${TIMER_NAME}"
        need_reload=true
    fi

    if [ "${need_reload}" = true ]; then
        systemctl daemon-reload
    fi

    # Remove binary
    if [ -f "${INSTALL_BIN_DIR}/${COMPONENT_NAME}" ]; then
        log_info "Removing binary..."
        rm -f "${INSTALL_BIN_DIR}/${COMPONENT_NAME}"
    fi

    # Preserve log directory by default
    if [ -d "${actual_log_dir}" ]; then
        log_info "Log directory preserved at ${actual_log_dir}"
    fi

    log_info "Uninstallation completed successfully"
}

# ======================== Upgrade ========================
do_upgrade() {
    if [ $# -gt 0 ]; then
        log_error "Upgrade does not accept any options"
        show_upgrade_help
        exit 1
    fi

    log_info "Upgrading ${COMPONENT_NAME}..."

    if [ ! -f "${SYSTEMD_UNIT_DIR}/${SERVICE_NAME}.service" ]; then
        log_error "Service is not installed. Please run 'install' first."
        exit 1
    fi

    check_binary

    local current_version
    current_version=$("${INSTALL_BIN_DIR}/${COMPONENT_NAME}" -v 2>/dev/null || echo "unknown")
    local target_version
    target_version=$("${BINARY_PATH}" -v 2>/dev/null || echo "unknown")

    log_info "Current version : ${current_version}"
    log_info "Target version  : ${target_version}"

    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        log_info "Stopping service..."
        systemctl stop "${SERVICE_NAME}"
    else
        log_warn "Service is not running"
    fi

    # Replace binary
    log_info "Replacing binary..."
    cp -f "${BINARY_PATH}" "${INSTALL_BIN_DIR}/${COMPONENT_NAME}"
    chmod 0500 "${INSTALL_BIN_DIR}/${COMPONENT_NAME}"

    # Reload systemd and restart service
    systemctl daemon-reload
    log_info "Starting service..."
    systemctl start "${SERVICE_NAME}"

    # Verify
    local version
    version=$("${INSTALL_BIN_DIR}/${COMPONENT_NAME}" -v 2>/dev/null || echo "unknown")
    log_info "Binary upgraded to: ${version}"

    if systemctl is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
        log_info "Upgrade completed successfully"
    else
        log_error "Service failed to start after upgrade. Check logs with: journalctl -u ${SERVICE_NAME} -f"
        exit 1
    fi
}

# ======================== Help ========================
show_help() {
    cat <<EOF
${COMPONENT_NAME} deployment tool - install or uninstall the service

Usage: $(basename "$0") <command> [options]

Commands:
  install       Install ${COMPONENT_NAME} as a systemd service
  upgrade       Upgrade ${COMPONENT_NAME} binary and restart service
  uninstall     Uninstall ${COMPONENT_NAME} service

Run '$(basename "$0") <command> --help' for more information on a command.
EOF
}

show_install_help() {
    cat <<EOF
Usage: $(basename "$0") install [options]

Install ${COMPONENT_NAME} as a systemd service.

Options:
  --runtimeType=<type>      Container runtime type (default: ${DEFAULT_RUNTIME_TYPE})
                            Valid values: docker, containerd
  --sockPath=<path>         Container runtime socket path (default: ${DEFAULT_SOCK_PATH})
                            Must not be a symlink
  --ctrStrategy=<strategy>  Faulty container recovery strategy (default: ${DEFAULT_CTR_STRATEGY})
                            never          - not recover containers
                            singleRecover  - recover single faulty chip container
                            ringRecover    - recover all related chip containers
  --logLevel=<level>        Log level (default: ${DEFAULT_LOG_LEVEL})
                            -1=debug, 0=info, 1=warning, 2=error, 3=critical
  --maxAge=<days>           Log backup retention days (default: ${DEFAULT_MAX_AGE}, range: [7, 700])
  --maxBackups=<count>      Max backup log files (default: ${DEFAULT_MAX_BACKUPS}, range: (0, 30])
  --faultConfig=<path>      Custom fault code configuration file path
                            File permission must not exceed 640, must not be a symlink
  --logPath=<path>          Log file path (default: ${DEFAULT_LOG_FILE})
                            Single log file rotates when exceeding 20MB
  --timerDelay=<duration>   Boot timer delay duration (default: ${DEFAULT_TIMER_DELAY})
                            Ensures NPU devices are ready before service starts
  -y, --yes                 Skip confirmation prompt (for scripted/automated use)

Examples:
  # Install with defaults (Docker runtime, no auto-recovery)
  $(basename "$0") install

  # Install with containerd and single recover strategy
  $(basename "$0") install --runtimeType=containerd --ctrStrategy=singleRecover

  # Install with timer (60s delay after boot), custom fault config and log path
  $(basename "$0") install --ctrStrategy=ringRecover --timerDelay=60s --faultConfig=/path/to/faultCode.json --logPath=/var/log/mindx-dl/container-manager/container-manager.log
EOF
}

show_upgrade_help() {
    cat <<EOF
Usage: $(basename "$0") upgrade

Upgrade ${COMPONENT_NAME} by replacing the binary and restarting the service.
The existing service configuration (systemd unit, startup parameters) is preserved.

Note: Use 'upgrade' instead of 'install' for in-place binary updates.
      Running 'install' will overwrite the existing service configuration.

Example:
  $(basename "$0") upgrade
EOF
}

show_uninstall_help() {
    cat <<EOF
Usage: $(basename "$0") uninstall

Uninstall ${COMPONENT_NAME} service.
Log files are preserved by default.

Example:
  $(basename "$0") uninstall
EOF
}

# ======================== Main ========================
main() {
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi

    local command="$1"
    shift

    case "${command}" in
        install)
            check_root
            check_systemd
            if [[ "$1" == "-h" || "$1" == "--help" ]]; then
                show_install_help
                exit 0
            fi
            do_install "$@"
            ;;
        uninstall)
            check_root
            check_systemd
            if [[ "$1" == "-h" || "$1" == "--help" ]]; then
                show_uninstall_help
                exit 0
            fi
            do_uninstall "$@"
            ;;
        upgrade)
            check_root
            check_systemd
            if [[ "$1" == "-h" || "$1" == "--help" ]]; then
                show_upgrade_help
                exit 0
            fi
            do_upgrade "$@"
            ;;
        -h|--help)
            show_help
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
