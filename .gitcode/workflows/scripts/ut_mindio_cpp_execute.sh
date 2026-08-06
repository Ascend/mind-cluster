set -x
gcc --version
g++ --version

if [ 1 == 1 ]; then
    apt-get -y install dos2unix
    pip3 install  numpy==1.26.4
    pip3 install lxml
    pip3 install beautifulsoup4
    pip3 install requests

    pip3 install wheel
    pip3 install --upgrade setuptools
    gcc --version

    whoami
    pwd
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
    [ ! -f 'run_gtest_ut.sh' ] && echo "can not found ut entry file - run_gtest_ut.sh, exit normally." && exit 0

    sed -i 's/set -e/set -ex/g' ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/scripts/run_gtest_ut.sh
    bash run_gtest_ut.sh
    ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/test/build/gcover_report/result | grep index.html
    ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/acp/test/build/gcover_report/result | grep test_detail.xml
fi

if grep -q "component/mindio/tft/src/csrc" "${ATOMGIT_WORKSPACE}/change.txt"; then
    apt-get -y install dos2unix
    pip3 install  numpy==1.26.4
    pip3 install lxml
    pip3 install beautifulsoup4
    pip3 install requests

    pip3 install wheel
    pip3 install --upgrade setuptools
    gcc --version

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

    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/test
    ls -al
    dos2unix *.sh && chmod +x *
    [ ! -f 'run_gtest_dt.sh' ] && echo "can not found dt entry file - run_gtest_dt.sh, exit normally." && exit 0
    bash run_gtest_dt.sh
    ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/test/build/report |grep report.xml
    ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/mindio/tft/test/build/gcover_report |grep index.html
fi
