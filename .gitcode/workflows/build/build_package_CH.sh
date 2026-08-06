#!/bin/bash
SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/utils.sh"

set -e
serviceName=$1
TARGET_BRANCH=$2
BUILDNUMBER=$2
buildtype=$3

version=$(get_version_number "$TARGET_BRANCH")
if [ $? -ne 0 ]; then
    echo "get Version_Number failed, please check."
    exit 1
fi

ostype=`arch`
uploadpackge(){
    if [ -d ${ATOMGIT_WORKSPACE}/$1 ];then
		cd ${ATOMGIT_WORKSPACE}/$1
	fi
	if [ "${serviceName}" == "ascend-docker-runtime" ]; then
		cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
        package_name=Alan-docker-runtime_${version}_linux-${ostype}.run
		echo "SHA256 Alan-docker-runtime_${version}_linux-${ostype}.run" > Alan-docker-runtime_${version}_linux-${ostype}.run.sha256sum
		SHA256_PG=$(sha256sum Alan-docker-runtime_${version}_linux-${ostype}.run)
		SHA256=(${SHA256_PG// / })
		sed -i "s#SHA256#${SHA256}#g" Alan-docker-runtime_${version}_linux-${ostype}.run.sha256sum
		cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
		commit_id=$(git rev-parse HEAD)
        echo servicename:${serviceName} >  Alan-docker-runtime_srcbaseline.txt
        echo branch:${branch} >>  Alan-docker-runtime_srcbaseline.txt
		echo commitid:${commit_id} >>  Alan-docker-runtime_srcbaseline.txt
        cp -rf Alan-docker-runtime_srcbaseline.txt ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
        touch ${package_name}.json
        cat>${package_name}.json<<EOF
        {
            "sha256Sum": "${SHA256}",
            "repoInfo": [{
                "repoUrl": "http://gitcode.com/ascend/mind-cluster.git",
                "repoBranch": "${branch}",
                "commitId": "${commit_id}"
                }],
            "buildTime": "$(date +%Y%m%d%H%M%S)"
        }
EOF
        cp -rf ${package_name}.json ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
	else
		if [ "${serviceName}" == "ascend-for-volcano" ]; then
			servicepackName="volcano"
            rm -rf ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Dockerfile-controller ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Dockerfile-scheduler
			echo  "*************${servicepackName}**************************"
		elif [ "${serviceName}" == "ascend-device-plugin" ]; then
			servicepackName="device-plugin"
		elif [ "${serviceName}" == "npu-exporter" ];then
			servicepackName="npu-exporter"
        elif [ "${serviceName}" == "noded" ];then
			servicepackName="noded"
        elif [ "${serviceName}" == "ascend-operator" ];then
			servicepackName="operator"
		else
			servicepackName="${serviceName}"
		fi
        cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
        package_name=Alan-${servicepackName}_${version}_linux-${ostype}.zip
        zip -r Alan-${servicepackName}_${version}_linux-${ostype}.zip ./*
        echo "SHA256 Alan-${servicepackName}_${version}_linux-${ostype}.zip" > Alan-${servicepackName}_${version}_linux-${ostype}.zip.sha256sum
        SHA256_PG=$(sha256sum Alan-${servicepackName}_${version}_linux-${ostype}.zip)
        SHA256=(${SHA256_PG// / })
        sed -i "s#SHA256#${SHA256}#g" Alan-${servicepackName}_${version}_linux-${ostype}.zip.sha256sum
        cd ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
        echo servicename:${serviceName} >   Alan-${servicepackName}_srcbaseline.txt
        echo branch:${branch} >>   Alan-${servicepackName}_srcbaseline.txt
        echo commitid:${commit_id} >>   Alan-${servicepackName}_srcbaseline.txt
        cp -rf Alan-${servicepackName}_srcbaseline.txt ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
        touch ${package_name}.json
        cat>${package_name}.json<<EOF
        {
            "sha256Sum": "${SHA256}",
            "repoInfo": [{
                "repoUrl": "http://gitcode.com/ascend/mind-cluster.git",
                "repoBranch": "${branch}",
                "commitId": "${commit_id}"
                }],
            "buildTime": "$(date +%Y%m%d%H%M%S)"
        }
EOF
        cp -rf ${package_name}.json ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
        chmod +x ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/*
        ls -al ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/
	fi
}


echo "*******************${serviceName}****${ATOMGIT_WORKSPACE}******************"

if [ "${serviceName}" == "ascend-for-volcano" ]; then
	uploadpackge ${serviceName}
    if [ "${buildtype}" == "version" ]; then
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.zip obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.sha256sum obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.json obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
    else
        echo "*****************************pr no need uplod obs****************************"
    fi
elif [ "${serviceName}" == "ascend-docker-runtime" ]; then
	echo "*******************${serviceName}**********************"
	uploadpackge ${serviceName}
    if [ "${buildtype}" == "version" ]; then
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.run obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.sha256sum obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.json obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
    else
        echo "*****************************pr no need uplod obs****************************"
    fi
else
	echo "*******************${serviceName}**********************"
	uploadpackge ${serviceName}
    if [ "${buildtype}" == "version" ]; then
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.zip obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.sha256sum obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
        obsutil cp ${ATOMGIT_WORKSPACE}/mind-cluster/component/$1/output/Alan*.json obs://mindcluster/mind-cluster/daily/version/${version}/${BUILDNUMBER}/ -f -r
    else
        echo "*****************************pr no need uplod obs****************************"
    fi
fi
