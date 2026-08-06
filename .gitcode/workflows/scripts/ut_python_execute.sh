set -ex
ls -al /opt/buildtools

export GOROOT=/opt/buildtools/go_1.21.12
export GO111MODULE='auto'
export GONOSUMDB=*
export GOPROXY="https://goproxy.cn,direct"
export PYTHON_HOME=/usr/local/python3.9.11/bin
export PATH=${PYTHON_HOME}:$GOROOT/bin:$PATH
export GOPATH=/opt/buildtools/

file_path="change.txt"
cd ${ATOMGIT_WORKSPACE}
if grep -q "taskd" "${ATOMGIT_WORKSPACE}/$file_path"; then
    echo "path: ${PATH}"
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/taskd/tests/ut/python
    dos2unix *.sh && chmod +x *
    bash -x run_test.sh
fi

if grep -q "component/mindio/acp/python_whl/mindio_acp/mindio_acp" "${ATOMGIT_WORKSPACE}/$file_path"; then
    echo "path: ${PATH}"
    pip3 install --upgrade pip
    pip3 --version

    pip3 install wheel
    pip3 install --upgrade setuptools
    pip3 install pytest # -i https://pypi.tuna.tsinghua.edu.cn/simple
    pip3 install pytest-cov
    pip3 install pytest-mock
    pip3 install numpy

    which python3.10
    ls -al /usr/bin/python3
    ls -al /usr/bin/python3.10
    rm -fr /usr/bin/python3
    rm -fr /usr/bin/python3.10

    export PYTHON_HOME=/opt/buildtools/python-3.9.11
    export Python3_ROOT=$PYTHON_HOME
    export PATH=$PYTHON_HOME/bin:$PATH
    export LD_LIBRARY_PATH=$PYTHON_HOME/lib:$LD_LIBRARY_PATH
    export CMAKE_PREFIX_PATH=$PYTHON_HOME:$CMAKE_PREFIX_PATH
    ln -sf $PYTHON_HOME/bin/python3 /usr/bin/python3
    ls -l $PYTHON_HOME/bin/python3
    python3 --version

    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/scripts
    dos2unix *.sh && chmod +x *
    [ ! -f 'run_python_ut.sh' ] && echo "can not found dt entry file - run_python_ut.sh， exit normally." && exit 0
    bash -x run_python_ut.sh
    pip3 install lxml
    pip3 install beautifulsoup4
    mv ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/output/* ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/scripts
    cat final.xml
fi
