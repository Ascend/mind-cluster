set -ex

cd ${ATOMGIT_WORKSPACE}
if [ 1 == 1 ]; then
    apt-get -y install dos2unix
    # pip3 install pandas==1.3.5
    # pip3 install ply==3.11
    pip3 install  numpy==1.26.4
    pip3 install openpyxl
    # pip3 install joblib==1.4.2
    pip3 install paramiko
    pip3 install scp
    pip3 install cryptography
    pip3 install openpyxl
    export PATH=/opt/buildtools/python-3.11.4/bin:$PATH
    cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/ascend-faultdiag/test/ && dos2unix *.sh && chmod +x * && bash run_dt.sh
else
    echo "***********************************no need compile***********************************"
fi
