#!/bin/bash

function get_version_number() {
    TARGET_BRANCH=$1
    local VERSION_NUMBER=$(date +%Y%m%d)
    case "$TARGET_BRANCH" in
        "release_v26.2.0")
            VERSION_NUMBER=26.2.0
            ;;
        "branch_v26.1.0")
            VERSION_NUMBER=26.1.0
            ;;
        "branch_v26.0.0")
            VERSION_NUMBER=26.0.0
            ;;
        "master")
            VERSION_NUMBER=26.3.0
            ;;
        *)
            echo "Please set the corresponding version number for branch ${TARGET_BRANCH} in advance" >&2
            return 1
            ;;
    esac
    echo "$VERSION_NUMBER"
}
