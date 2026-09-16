set -ex

cd ${ATOMGIT_WORKSPACE}
ls -al ./

unzip -o "artifacts_x86_64.zip"
unzip -o "artifacts_aarch64.zip"

export BASE_DIR=${ATOMGIT_WORKSPACE}/tests/st/testcases/
cp -rf /workspace/ST/test.sh ./
ls -al "${ATOMGIT_WORKSPACE}"
sh test.sh ${ATOMGIT_WORKSPACE}
