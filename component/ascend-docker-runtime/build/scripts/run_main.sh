#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2020-2022. All rights reserved.
# Description: ascend-docker-runtime run package script
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

RT_LOWER_CASE="ascend-docker-runtime"
RT_FIRST_CASE="Ascend-docker-runtime"

args=($@)
start_arg="${args[0]}"
start_script=${start_arg#*--}

ASCEND_RUNTIME_CONFIG_DIR=/etc/${RT_LOWER_CASE}.d
DOCKER_CONFIG_DIR=/etc/docker
CONTAINERD_CONFIG_DIR=/etc/containerd
CRIO_CONFIG_DIR=/etc/crio
CONFIG_FILE_PATH=""
INSTALL_SCENE=docker
INJECTION_MODE=""
INSTALL_PATH=/usr/local/Ascend/Ascend-Docker-Runtime
readonly INSTALL_LOG_DIR=/var/log/${RT_LOWER_CASE}
readonly INSTALL_LOG_PATH=${INSTALL_LOG_DIR}/installer.log
readonly INSTALL_LOG_PATH_BAK=${INSTALL_LOG_DIR}/installer_bak.log
readonly LOG_SIZE_THRESHOLD=$((20*1024*1024))
readonly PACKAGE_VERSION=REPLACE_VERSION
readonly UNIFIED_BUS_DIR="/sys/class/unified_bus"

readonly PACKAGE_COMMIT=REPLACE_COMMIT
readonly PACKAGE_BRANCH=REPLACE_BRANCH
readonly PACKAGE_GO=REPLACE_GO

umask 027

function check_log {
    if [[ ! -d ${INSTALL_LOG_DIR} ]]; then
        mkdir -p -m 750 ${INSTALL_LOG_DIR}
    fi

    check_sub_path ${INSTALL_LOG_DIR}
    if [[ $? != 0 ]]; then
        echo "[ERROR] ${INSTALL_LOG_DIR} is invalid"
        exit 1
    fi

    if [[ ! -f ${INSTALL_LOG_PATH} ]]; then
        touch ${INSTALL_LOG_PATH}
        chmod 640 ${INSTALL_LOG_PATH}
        return
    fi

    local log_size="$(ls -l ${INSTALL_LOG_PATH} | awk '{ print $5 }')"
    if [[ ${log_size} -ge ${LOG_SIZE_THRESHOLD} ]]; then
        mv -f ${INSTALL_LOG_PATH} ${INSTALL_LOG_PATH_BAK}
        chmod 400 ${INSTALL_LOG_PATH_BAK}
        > ${INSTALL_LOG_PATH}
        chmod 640 ${INSTALL_LOG_PATH}
    fi
}

function log {
    local ip="${SSH_CLIENT%% *}"
    if [ "${ip}" = "" ]; then
        ip="localhost"
    fi
    echo "$1 $2"
    echo "$1 [$(date +'%Y/%m/%d %H:%M:%S')] [uid: ${UID}] [${ip}] [${RT_LOWER_CASE}] $2" >> ${INSTALL_LOG_PATH}
}

function check_path {
    local path="$1"
    if [[ ${#path} -gt 1024 ]] || [[ ${#path} -le 0 ]]; then
        echo "[ERROR] parameter is invalid, length not in 1~1024"
        return 1
    fi
    if [[ -n $(echo "${path}" | grep -Ev '^[a-zA-Z0-9./_-]*$') ]]; then
        echo "[ERROR] parameter is invalid, char not all in 'a-zA-Z0-9./_-'"
        return 1
    fi
    path=$(realpath -m -s "${path}")
    while [[ ! -e "${path}" ]]; do
        path=$(dirname "${path}")
    done
    while true; do
        if [[ "${path}" == "/" ]]; then
            break
        fi
        check_path_permission "${path}"
        if [[ $? != 0 ]]; then
            return 1
        fi
        path=$(dirname "${path}")
    done
}

function check_sub_path {
    local path="$1"
    while [[ ! -e "${path}" ]]; do
        return 1
    done
    for file in $(find "${path}"); do
        check_path_permission "${file}"
        if [[ $? != 0 ]]; then
            return 1
        fi
    done
}

function check_path_permission {
    local path="$1"
    if [[ -L "${path}" ]]; then
        echo "[ERROR] ${path} is soft link"
        return 1
    fi
    if [[ $(stat -c %u "${path}") != 0 ]] || [[ "$(stat -c %g ${path})" != 0 ]]; then
        echo "[ERROR] user or group of ${path} is not root"
        return 1
    fi
    local permission=$(stat -c %A "${path}")
    if [[ $(echo "${permission}" | cut -c6) == w ]] || [[ $(echo "${permission}" | cut -c9) == w ]]; then
        echo "[ERROR] group or other of ${path} has write permission"
        return 1
    fi
}

function print_version {
    echo "${RT_LOWER_CASE} version: ${PACKAGE_VERSION}"
}

function print_help {
    echo "Error input
Usage: ./${RT_FIRST_CASE}_${PACKAGE_VERSION}_linux-$(uname -m).run [options]
Options:
  --help | -h                   Print this message
  --install                     Install into this system
  --install-path                Specify the installation path, which must be absolute path
  --uninstall                   Uninstall the installed ${RT_FIRST_CASE} tool
  --upgrade                     Upgrade the installed ${RT_FIRST_CASE} tool
  --install-type=<type>         Only A500, A500A2, A200ISoC, A200IA2 and A200 need to specify
                                the installation type of ${RT_FIRST_CASE}
                                (eg: --install-type=A200IA2, when your product is A200I A2 or A200I DK A2)
  --version                     Query ${RT_FIRST_CASE} version
  --install-scene=<scene>       Installation scenario, only docker, containerd, crio or isula(eg: --install-scene=docker, default: docker)
  --injection-mode=<mode>       Injection mode for NPU devices, cdi or legacy (eg: --injection-mode=cdi, default: legacy)
  --config-file-path            Specifies the path of the Docker, containerd or CRI-O configuration file
                                (eg: --config-file-path=/etc/containerd/config.toml).
                                If this parameter is not specified, the default configuration file path
                                of docker, containerd or crio is used. For docker, the path is /etc/docker/daemon.json.
                                For containerd, the path is /etc/containerd/config.toml.
                                For crio, the path is /etc/crio/crio.conf.d/99-ascend-runtime.conf.
"
}

function check_platform {
  plat="$(uname -m)"
  if [[ $start_script =~ $plat ]]; then
    echo "[INFO] platform($plat) matched!"
    return 0
  else
    echo "[ERROR] platform($plat) mismatch for $start_script, please check it"
    return 1
  fi
}

function save_install_args() {
    # default injection-mode to legacy when not explicitly specified,
    # so the install.info records "injection-mode=legacy" instead of an
    # empty value when --injection-mode is omitted.
    if [[ -z "${INJECTION_MODE}" ]]; then
        log "[INFO]" "injection-mode not specified, defaulting to legacy"
        INJECTION_MODE=legacy
    fi
    {
      echo -e "version=v${PACKAGE_VERSION}"
      echo -e "arch=$(uname -m)"
      echo -e "os=linux"
      echo -e "gitCommit=${PACKAGE_COMMIT}"
      echo -e "gitBranch=${PACKAGE_BRANCH}"
      echo -e "goVersion=${PACKAGE_GO}"
      echo -e "path=${INSTALL_PATH}"
      echo -e "build=${RT_FIRST_CASE}_${PACKAGE_VERSION}-$(uname -m)"
      echo -e "a500=${a500}"
      echo -e "a500a2=${a500a2}"
      echo -e "a200=${a200}"
      echo -e "a200isoc=${a200isoc}"
      echo -e "a200ia2=${a200ia2}"
      echo -e "install-scene=${INSTALL_SCENE}"
      echo -e "config-file-path=${CONFIG_FILE_PATH}"
      echo -e "injection-mode=${INJECTION_MODE}"
    } > "${INSTALL_PATH}"/ascend_docker_runtime_install.info
    chmod 640 ${INSTALL_PATH}/ascend_docker_runtime_install.info
}

function add_so() {
    check_path "/etc/os-release"
    if [[ $? != 0 ]]; then
        echo "[ERROR] /etc/os-release is invalid"
        return 1
    fi
    if grep -qi "ubuntu" "/etc/os-release"; then
      echo "[info] os is Ubuntu"
      echo -e "\n/usr/lib/aarch64-linux-gnu/libcrypto.so.1.1" >> ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
      echo "/usr/lib/aarch64-linux-gnu/libyaml-0.so.2" >> ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    elif grep -qi "euler" "/etc/os-release"; then
      echo "[info] os is Euler/OpenEuler"
      echo -e "\n/usr/lib64/libcrypto.so.1.1" >> ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
      echo "/usr/lib64/libyaml-0.so.2" >> ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    else
      echo "[ERROR] not support this os"
      return 1
    fi
}

function check_eula()
{
    local eula_file="$(cd "$(dirname "$0")" && pwd)/agreement.txt"
    if [[ -f "${eula_file}" ]]; then
        cat "${eula_file}"
        echo ""
    fi
    echo "Do you accept the EULA to install ${RT_FIRST_CASE}?[Y/N]"
    while true
    do
        read -r yn
        case "${yn}" in
            [Yy])
                log "[INFO]" "user accepted EULA"
                break
                ;;
            [Nn])
                log "[INFO]" "install cancelled, user declined EULA"
                exit 0
                ;;
            *)
                echo "Do you accept the EULA to install ${RT_FIRST_CASE}?[Y/N]"
                ;;
        esac
    done
}

function prepare_ub_driver_list()
{
    if [[ -d "${UNIFIED_BUS_DIR}" ]]; then
        log "[INFO] unified bus dir ${UNIFIED_BUS_DIR} exists, will mount the ub_driver.list content"
        echo -e "\033[31m[WARNING] mounting the ub_driver.list content may involve glibc version compatibility between the host and the container. Please ensure the host glibc version is compatible with the glibc version in the container to avoid runtime failures. To skip mounting the ub_driver.list content, set ASCEND_UB_DRV_MOUNT=False.\033[0m"
        check_path ${ASCEND_RUNTIME_CONFIG_DIR}/ub_driver.list
        if [[ $? != 0 ]]; then
            log "[ERROR]" "check failed, ${ASCEND_RUNTIME_CONFIG_DIR}/ub_driver.list is invalid"
            return 1
        fi
        cp -f ./ub_driver.list ${ASCEND_RUNTIME_CONFIG_DIR}/ub_driver.list
        chmod 440 ${ASCEND_RUNTIME_CONFIG_DIR}/ub_driver.list
    fi
}

function install()
{
    check_eula
    echo "[INFO] installing ${RT_LOWER_CASE}"
    check_platform
    if [[ $? != 0 ]]; then
        log "[ERROR]" "install failed, run package and os not matched in arch"
        exit 1
    fi

    if [[ ! ${INSTALL_PATH} =~ ^/ ]]; then
        echo "[ERROR]: install path: ${INSTALL_PATH} is a relative path, please use absolute path"
        exit 1
    fi

    check_path "${INSTALL_PATH}"
    if [[ $? != 0 ]]; then
        log "[ERROR]" "install failed, ${INSTALL_PATH} is invalid"
        exit 1
    fi
    [[ ! -d "${INSTALL_PATH}" ]] && mkdir -p -m 750 "${INSTALL_PATH}"
    [[ ! -d "${INSTALL_PATH}/assets" ]] && mkdir -p -m 750 "${INSTALL_PATH}/assets"
    [[ ! -d "${INSTALL_PATH}/script" ]] && mkdir -p -m 750 "${INSTALL_PATH}/script"

    check_sub_path "${INSTALL_PATH}"
    if [[ $? != 0 ]]; then
        log "[ERROR]" "install failed, ${INSTALL_PATH} or ${INSTALL_PATH}/assets or ${INSTALL_PATH}/script is invalid"
        exit 1
    fi

    cp -f ./ascend-docker-runtime ${INSTALL_PATH}/ascend-docker-runtime
    cp -f ./ascend-docker-hook ${INSTALL_PATH}/ascend-docker-hook
    cp -f ./ascend-docker-cli ${INSTALL_PATH}/ascend-docker-cli
    cp -f ./ascend-docker-plugin-install-helper ${INSTALL_PATH}/ascend-docker-plugin-install-helper
    cp -f ./ascend-docker-destroy ${INSTALL_PATH}/ascend-docker-destroy
    cp -f ./README.md ${INSTALL_PATH}/README.md
    chmod 550 ${INSTALL_PATH}/ascend-docker-runtime
    chmod 550 ${INSTALL_PATH}/ascend-docker-hook
    chmod 550 ${INSTALL_PATH}/ascend-docker-cli
    chmod 550 ${INSTALL_PATH}/ascend-docker-plugin-install-helper
    chmod 550 ${INSTALL_PATH}/ascend-docker-destroy
    chmod 640 ${INSTALL_PATH}/README.md

    cp -f ./assets/20230118566.png ${INSTALL_PATH}/assets/20230118566.png
    cp -f ./assets/20210329102949456.png ${INSTALL_PATH}/assets/20210329102949456.png
    chmod 640 ${INSTALL_PATH}/assets/20230118566.png ${INSTALL_PATH}/assets/20210329102949456.png

    cp -f ./uninstall.sh ${INSTALL_PATH}/script/uninstall.sh
    chmod 500 ${INSTALL_PATH}/script/uninstall.sh

    check_path ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    if [[ $? != 0 ]]; then
        log "[ERROR]" "install failed, ${ASCEND_RUNTIME_CONFIG_DIR}/base.list is invalid"
        exit 1
    fi
    [[ ! -d ${ASCEND_RUNTIME_CONFIG_DIR} ]] && mkdir -p -m 750 ${ASCEND_RUNTIME_CONFIG_DIR}

    if [ "${a500}" == "y" ]; then
        cp -f ./base.list_A500 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    elif [ "${a200}" == "y" ]; then
        cp -f ./base.list_A200 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    elif [ "${a200isoc}" == "y" ]; then
        cp -f ./base.list_A200ISoC ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    elif [ "${a500a2}" == "y" ]; then
        cp -f ./base.list_A500A2 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        add_so
        if [[ $? != 0 ]]; then
            log "[ERROR]" "install failed, a500a2 not support this os"
            exit 1
        fi
    elif [ "${a200ia2}" == "y" ]; then
        cp -f ./base.list_A200IA2 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        add_so
        if [[ $? != 0 ]]; then
            log "[ERROR]" "install failed, a200ia2 not support this os"
            exit 1
        fi
    else
        cp -f ./base.list ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    fi
    chmod 440 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list

    if ! prepare_ub_driver_list; then
        log "[ERROR]" "install failed, prepare ub_driver.list failed"
        exit 1
    fi

    echo "[INFO] install executable files success"

    if [[ ${CONFIG_FILE_PATH} == "" ]]; then
        if [ "${INSTALL_SCENE}" == "docker" ] || [ "${INSTALL_SCENE}" == "isula" ]; then
                echo "[INFO] install scene is 'docker'."
                check_path ${DOCKER_CONFIG_DIR}/daemon.json
                if [[ $? != 0 ]]; then
                    log "[ERROR]" "install failed, ${DOCKER_CONFIG_DIR}/daemon.json is invalid"
                    exit 1
                fi
                [[ ! -d ${DOCKER_CONFIG_DIR} ]] && mkdir -p -m 750 ${DOCKER_CONFIG_DIR}

                SRC="${DOCKER_CONFIG_DIR}/daemon.json.${PPID}"
                DST="${DOCKER_CONFIG_DIR}/daemon.json"
            elif [[ "${INSTALL_SCENE}" == "containerd" ]]; then
                echo "[INFO] install scene is 'containerd'."
                check_path ${CONTAINERD_CONFIG_DIR}/config.toml
                if [[ $? != 0 ]]; then
                    log "[ERROR]" "install failed, ${CONTAINERD_CONFIG_DIR}/config.toml is invalid"
                    exit 1
                fi
                [[ ! -d ${CONTAINERD_CONFIG_DIR} ]] && mkdir -p -m 750 ${CONTAINERD_CONFIG_DIR}

                SRC="${CONTAINERD_CONFIG_DIR}/config.toml.${PPID}"
                DST="${CONTAINERD_CONFIG_DIR}/config.toml"
                if [ ! -e ${DST} ]; then
                  echo "[INFO] containerd config file does not exist, default ${DST} will be created"
                  containerd config default > ${DST}
                fi
            elif [[ "${INSTALL_SCENE}" == "crio" ]]; then
                echo "[INFO] install scene is 'crio'."
                [[ ! -d ${CRIO_CONFIG_DIR}/crio.conf.d ]] && mkdir -p -m 750 ${CRIO_CONFIG_DIR}/crio.conf.d

                SRC="${CRIO_CONFIG_DIR}/crio.conf.d/99-ascend-runtime.conf.${PPID}"
                DST="${CRIO_CONFIG_DIR}/crio.conf.d/99-ascend-runtime.conf"
            else
                log "[ERROR]" "install failed, invalid value '${INSTALL_SCENE}' of 'install-scene' "
                exit 1
            fi
    else
        SRC="${CONFIG_FILE_PATH}.${PPID}"
        DST="${CONFIG_FILE_PATH}"
    fi

    # exit when return code is not 0, if use 'set -e'
    ./ascend-docker-plugin-install-helper add ${DST} ${SRC} ${INSTALL_PATH}/ascend-docker-runtime ${RESERVEDEFAULT} ${INSTALL_SCENE}  > /dev/null
    if [[ $? != 0 ]]; then
        log "[ERROR]" "install failed, './ascend-docker-plugin-install-helper add ${DST} ${SRC} ${INSTALL_PATH}/ascend-docker-runtime ${RESERVEDEFAULT} ${INSTALL_SCENE} ' return non-zero"
        exit 1
    fi
    mv -f ${SRC} ${DST}
    log "[INFO]" "${DST} modify success"
    chmod 600 ${DST}

    save_install_args
    echo "[INFO] ${RT_LOWER_CASE} has been installed in: ${INSTALL_PATH}"
    echo "[INFO] The version of ${RT_LOWER_CASE} is: ${PACKAGE_VERSION}"
    echo '[INFO] please reboot daemon and container engine to take effect'
    log "[INFO]" "${RT_LOWER_CASE} install success"
}

function uninstall()
{
    echo "[INFO] uninstalling ${RT_LOWER_CASE} ${PACKAGE_VERSION}"

    if [ ! -d "${INSTALL_PATH}" ]; then
        log "[WARNING]" "uninstall skipping, the specified install path does not exist"
        exit 0
    fi

    check_path "${INSTALL_PATH}"
    if [[ $? != 0 ]]; then
        log "[ERROR]" "uninstall failed, ${INSTALL_PATH} or ${INSTALL_PATH}/script is invalid"
        exit 1
    fi

    "${INSTALL_PATH}"/script/uninstall.sh ${ISULA} ${INSTALL_SCENE} ${CONFIG_FILE_PATH}
    if [[ $? != 0 ]]; then
        log "[ERROR]" "uninstall failed, '${INSTALL_PATH}/script/uninstall.sh ${ISULA} ${INSTALL_SCENE} ${CONFIG_FILE_PATH}' return non-zero"
        exit 1
    fi

    log "[INFO]" "${RT_LOWER_CASE} uninstall success"
}

function upgrade()
{
    check_eula
    echo "[INFO] upgrading ${RT_LOWER_CASE}"
    check_platform
    if [[ $? != 0 ]]; then
        log "[ERROR]" "upgrade failed, run package and os not matched in arch"
        exit 1
    fi

    if [ ! -d "${INSTALL_PATH}" ]; then
        log "[ERROR]" "upgrade failed, the specified install path does not exist, stopping upgrading"
        exit 1
    fi

    if [ ! -d "${ASCEND_RUNTIME_CONFIG_DIR}" ]; then
        log "[ERROR]" "upgrade failed, the configuration directory does not exist"
        exit 1
    fi

    check_path "${INSTALL_PATH}" && check_sub_path "${INSTALL_PATH}"
    if [[ $? != 0 ]]; then
        log "[ERROR]" "upgrade failed, ${INSTALL_PATH} or ${INSTALL_PATH}/script is invalid"
        exit 1
    fi

    cp -f ./ascend-docker-runtime ${INSTALL_PATH}/ascend-docker-runtime
    cp -f ./ascend-docker-hook ${INSTALL_PATH}/ascend-docker-hook
    cp -f ./ascend-docker-cli ${INSTALL_PATH}/ascend-docker-cli
    cp -f ./ascend-docker-plugin-install-helper ${INSTALL_PATH}/ascend-docker-plugin-install-helper
    cp -f ./ascend-docker-destroy ${INSTALL_PATH}/ascend-docker-destroy
    cp -f ./uninstall.sh ${INSTALL_PATH}/script/uninstall.sh
    chmod 550 ${INSTALL_PATH}/ascend-docker-runtime
    chmod 550 ${INSTALL_PATH}/ascend-docker-hook
    chmod 550 ${INSTALL_PATH}/ascend-docker-cli
    chmod 550 ${INSTALL_PATH}/ascend-docker-plugin-install-helper
    chmod 550 ${INSTALL_PATH}/ascend-docker-destroy
    chmod 500 ${INSTALL_PATH}/script/uninstall.sh

    check_path ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
    if [[ $? != 0 ]]; then
        log "[ERROR]" "upgrade failed, ${ASCEND_RUNTIME_CONFIG_DIR}/base.list is invalid"
        exit 1
    fi

    # Preserve the previously configured injection-mode when upgrading without an
    # explicit --injection-mode argument, so save_install_args does not overwrite
    # the install.info record with an empty value.
    if [[ -z "${INJECTION_MODE}" ]] && [ -f "${INSTALL_PATH}"/ascend_docker_runtime_install.info ]; then
        INJECTION_MODE=$(grep "^injection-mode=" "${INSTALL_PATH}"/ascend_docker_runtime_install.info | head -n 1 | cut -d"=" -f2-)
    fi

    if [ -f "${INSTALL_PATH}"/ascend_docker_runtime_install.info ]; then
        if [ "$(grep "a500=y" "${INSTALL_PATH}"/ascend_docker_runtime_install.info)" == "a500=y" ];then
            a500=y
            cp -f ./base.list_A500 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        elif [ "$(grep "a500a2=y" "${INSTALL_PATH}"/ascend_docker_runtime_install.info)" == "a500a2=y" ]; then
            a500a2=y
            cp -f ./base.list_A500A2 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
            add_so
            if [[ $? != 0 ]]; then
                log "[ERROR]" "upgrade failed, a500a2 not support this os"
                exit 1
            fi
        elif [ "$(grep "a200=y" "${INSTALL_PATH}"/ascend_docker_runtime_install.info)" == "a200=y" ]; then
            a200=y
            cp -f ./base.list_A200 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        elif [ "x$(grep "a200isoc=y" "${INSTALL_PATH}"/ascend_docker_runtime_install.info)" == "xa200isoc=y" ]; then
            a200isoc=y
            cp -f ./base.list_A200ISoC ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        elif [ "x$(grep "a200ia2=y" "${INSTALL_PATH}"/ascend_docker_runtime_install.info)" == "xa200ia2=y" ]; then
            a200ia2=y
            cp -f ./base.list_A200IA2 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
            add_so
            if [[ $? != 0 ]]; then
                log "[ERROR]" "upgrade failed, a200a2 not support this os"
                exit 1
            fi
        else
            cp -f ./base.list ${ASCEND_RUNTIME_CONFIG_DIR}/base.list
        fi
        save_install_args
    fi
    chmod 440 ${ASCEND_RUNTIME_CONFIG_DIR}/base.list

    if ! prepare_ub_driver_list; then
        log "[ERROR]" "upgrade failed, prepare ub_driver.list failed"
        exit 1
    fi

    echo "[INFO] ${RT_LOWER_CASE} has been installed in: ${INSTALL_PATH}"
    echo "[INFO] upgrade ${RT_LOWER_CASE} success"
    echo "[INFO] The version of ${RT_LOWER_CASE} is: v${PACKAGE_VERSION}"
    log "[INFO]" "${RT_LOWER_CASE} upgrade success"
}
INSTALL_SCENE_FLAG=n
CONFIG_FILE_PATH_FLAG=n
INSTALL_FLAG=n
INSTALL_PATH_FLAG=n
UNINSTALL_FLAG=n
UPGRADE_FLAG=n
a500=n
a200=n
a200isoc=n
a500a2=n
a200ia2=n
ISULA=none
RESERVEDEFAULT=no
INJECTION_MODE_FLAG=n
need_help=y

check_log

# must run with root permission
if [ "${UID}" != "0" ]; then
    log "[ERROR]" "failed, please run with root permission"
    exit 1
fi

while true
do
    case "$3" in
        --install-scene=*)
            if [ "${INSTALL_SCENE_FLAG}" == "y" ]; then
                log "[ERROR]" "failed, '--install-scene' Repeat parameter!"
                exit 1
            fi
            if [ "${ISULA}" == "isula" ]; then
                log "[ERROR]" "failed, incompatible parameters: '--install-scene' !"
                exit 1
            fi
            need_help=n
            INSTALL_SCENE_FLAG=y
            if [ "$3" == "--install-scene=docker" ]; then
                INSTALL_SCENE=docker
            elif [ "$3" == "--install-scene=containerd" ]; then
                INSTALL_SCENE=containerd
            elif [ "$3" == "--install-scene=crio" ]; then
                INSTALL_SCENE=crio
            elif [ "$3" == "--install-scene=isula" ]; then
                INSTALL_SCENE=isula
                DOCKER_CONFIG_DIR="/etc/isulad"
                RESERVEDEFAULT=yes
            else
                log "[ERROR]" "failed, please check the parameter of --install-scene=<scene>"
                exit 1
            fi
            shift
            ;;
        --injection-mode=*)
            if [ "${INJECTION_MODE_FLAG}" == "y" ]; then
                log "[ERROR]" "failed, '--injection-mode' Repeat parameter!"
                exit 1
            fi
            need_help=n
            INJECTION_MODE_FLAG=y
            if [ "$3" == "--injection-mode=cdi" ]; then
                INJECTION_MODE=cdi
            elif [ "$3" == "--injection-mode=legacy" ]; then
                INJECTION_MODE=legacy
            else
                log "[ERROR]" "failed, please check the parameter of --injection-mode=<mode>, only cdi or legacy are supported"
                exit 1
            fi
            shift
            ;;
        --config-file-path=*)
            if [ "${CONFIG_FILE_PATH_FLAG}" == "y" ]; then
                log "[ERROR]" "failed, '--config-file-path' Repeat parameter!"
                exit 1
            fi
            need_help=n
            CONFIG_FILE_PATH_FLAG=y
            CONFIG_FILE_PATH=$(echo $3 | cut -d"=" -f2)
            if [[ ! -e "$CONFIG_FILE_PATH" ]]; then
                log "[ERROR]" "failed, file '$CONFIG_FILE_PATH' does not exist."
                exit 1
            fi
            shift
            ;;
        --install)
            if [ "${INSTALL_FLAG}" == "y" ]; then
                log "[ERROR]" "install failed, '--install' Repeat parameter!"
                exit 1
            fi
            need_help=n
            INSTALL_FLAG=y
            shift
            ;;
        --uninstall)
            if [ "${UNINSTALL_FLAG}" == "y" ]; then
                log "[ERROR]" "uninstall failed, '--uninstall' Repeat parameter!"
                exit 1
            fi
            need_help=n
            UNINSTALL_FLAG=y
            shift
            ;;
        --install-path=*)
            if [ "${INSTALL_PATH_FLAG}" == "y" ]; then
                log "[ERROR]" "failed, '--install-path' Repeat parameter!"
                exit 1
            fi
            need_help=n
            INSTALL_PATH_FLAG=y
            INSTALL_PATH=$(echo $3 | cut -d"=" -f2)
            INSTALL_PATH=$(echo ${INSTALL_PATH}/Ascend-Docker-Runtime | sed "s/\/*$//g")
            shift
            ;;
        --upgrade)
            if [ "${UPGRADE_FLAG}" == "y" ]; then
                log "[ERROR]" "upgrade failed, '--upgrade' Repeat parameter!"
                exit 1
            fi
            need_help=n
            UPGRADE_FLAG=y
            shift
            ;;
        --install-type=*)
            if [ "${a500}" == "y" ] || [ "${a200}" == "y" ] || [ "${a200isoc}" == "y" ] ||
            [ "${a200ia2}" == "y" ] || [ "${a500a2}" == "y" ]; then
                log "[ERROR]" "failed, '--install-type' Repeat parameter!"
                exit 1
            fi
            need_help=n

            if [ "$3" == "--install-type=A500" ]; then
                a500=y
            elif [ "$3" == "--install-type=A200" ]; then
                a200=y
            elif [ "$3" == "--install-type=A200ISoC" ]; then
                a200isoc=y
            elif [ "$3" == "--install-type=A500A2" ]; then
                a500a2=y
            elif [ "$3" == "--install-type=A200IA2" ]; then
                a200ia2=y
            else
                log "[ERROR]" "failed, please check the parameter of --install-type=<type>"
                exit 1
            fi
            shift
            ;;
        --version)
            need_help=n
            print_version
            exit 0
            shift
            ;;
        *)
            if [ "x$3" != "x" ]; then
                log "[ERROR]" "failed, unsupported parameters: $3"
                print_help
                exit 1
            fi
            break
            ;;
    esac
done

# it is not allowed to input only install-path
if [ "${INSTALL_PATH_FLAG}" == "y" ] && \
   [ "${INSTALL_FLAG}" == "n" ] && \
   [ "${UNINSTALL_FLAG}" == "n" ] && \
   [ "${UPGRADE_FLAG}" == "n" ]; then
      log "[ERROR]" "failed, only input <install_path> command. When use --install-path you also need input --install or --uninstall or --upgrade"
      exit 1
fi

# it is not allowed to input only install-scene
if [ "${INSTALL_SCENE_FLAG}" == "y" ] && \
   [ "${INSTALL_FLAG}" == "n" ] && \
   [ "${UNINSTALL_FLAG}" == "n" ] && \
   [ "${UPGRADE_FLAG}" == "n" ]; then
      log "[ERROR]" "failed, only input <install-scene> command. When use --install-scene you also need input --install or --uninstall or --upgrade"
      exit 1
fi

# it is not allowed to input only injection-mode
if [ "${INJECTION_MODE_FLAG}" == "y" ] && \
   [ "${INSTALL_FLAG}" == "n" ] && \
   [ "${UPGRADE_FLAG}" == "n" ]; then
      log "[ERROR]" "failed, only input <injection-mode> command. When use --injection-mode you also need input --install or --upgrade"
      exit 1
fi

if [ "${INSTALL_FLAG}" == "y" ]; then
    install
    exit 0
fi

if [ "${UNINSTALL_FLAG}" == "y" ]; then
    uninstall
    exit 0
fi

if [ "${UPGRADE_FLAG}" == "y" ]; then
    upgrade
    exit 0
fi

if [ "${need_help}" == "y" ]; then
  print_help
  exit 0
fi
