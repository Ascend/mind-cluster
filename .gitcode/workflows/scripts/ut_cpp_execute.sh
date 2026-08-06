set -ex

file_path="change.txt"
if grep -q "ascend-docker-runtime" "${ATOMGIT_WORKSPACE}/$file_path"; then
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-docker-runtime/cli/test/dt
    sh build.sh
    ls -la ./xml
    cat ./xml/test_detail.xml
    ls -la ./html
    cat ./html/index.html
else
    echo "**********************no change ascend-docker-runtime*********************"
fi
