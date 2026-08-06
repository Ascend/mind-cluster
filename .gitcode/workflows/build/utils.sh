#!/bin/bash

function get_version_number() {
    TARGET_BRANCH=$1
    local VERSION_NUMBER=$(date +%Y%m%d)
    case "$TARGET_BRANCH" in
        "branch_v26.1.0")
            VERSION_NUMBER=26.1.0
            ;;
        "branch_v26.0.0")
            VERSION_NUMBER=26.0.1
            ;;
        "branch_v7.3.0")
            VERSION_NUMBER=7.3.2
            ;;
        "master")
            VERSION_NUMBER=26.2.0
            ;;
        *)
            echo "Please set the corresponding version number for branch ${TARGET_BRANCH} in advance" >&2
            return 1
            ;;
    esac
    echo "$VERSION_NUMBER"
}
