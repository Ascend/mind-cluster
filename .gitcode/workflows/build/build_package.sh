#!/bin/bash
SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/utils.sh"

set -e
serviceName=$1
TARGET_BRANCH=$2
BUILDNUMBER=$3
buildtype=$4
package_flag=$5

version=$(get_version_number "$TARGET_BRANCH")
if [ $? -ne 0 ]; then
    echo "get Version_Number failed, please check."
    exit 1
fi

ostype=`arch`
uploadpackge(){
    if [ -d ${WORKSPACE}/$1 ];then
        cd ${WORKSPACE}/$1
    fi

    if [ "${serviceName}" == "ascend-docker-runtime" ]; then
        cd ${WORKSPACE}/mind-cluster/component/$1/output/
        package_name=Ascend-docker-runtime_${version}_linux-${ostype}.run
        echo "SHA256 Ascend-docker-runtime_${version}_linux-${ostype}.run" > Ascend-docker-runtime_${version}_linux-${ostype}.run.sha256sum
        SHA256_PG=$(sha256sum Ascend-docker-runtime_${version}_linux-${ostype}.run)
        SHA256=(${SHA256_PG// / })
        sed -i "s#SHA256#${SHA256}#g" Ascend-docker-runtime_${version}_linux-${ostype}.run.sha256sum
        cd ${WORKSPACE}/mind-cluster/component/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
        echo servicename:${serviceName} >  Ascend-docker-runtime_srcbaseline.txt
        echo branch:${branch} >>  Ascend-docker-runtime_srcbaseline.txt
        echo commitid:${commit_id} >>  Ascend-docker-runtime_srcbaseline.txt
        cp -rf Ascend-docker-runtime_srcbaseline.txt ${WORKSPACE}/mind-cluster/component/$1/output/
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
        cp -rf ${package_name}.json ${WORKSPACE}/mind-cluster/component/$1/output/
    elif [ "${serviceName}" == "helm-deploy-tool" ];then
        cd ${WORKSPACE}/mind-cluster/$1/output/
        package_name=Ascend-helm-deploy-tool_${version}_linux.zip
        zip -r ${package_name} ./*
        SHA256_PG=$(sha256sum ${package_name})
        SHA256=(${SHA256_PG// / })
        echo "${SHA256} ${package_name}" > ${package_name}.sha256sum
        cd ${WORKSPACE}/mind-cluster/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
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
        cp -rf ${package_name}.json ${WORKSPACE}/mind-cluster/$1/output/
    elif [ "${serviceName}" == "ascend-deployer" ];then
        cd ${WORKSPACE}/$1/dist/
        package_name=${package_flag}_${version}_linux.zip
        zip -r ${package_flag}_${version}_linux.zip ./*
        echo "SHA256 ${package_flag}_${version}_linux.zip" > ${package_flag}_${version}_linux.zip.sha256sum
        SHA256_PG=$(sha256sum ${package_flag}_${version}_linux.zip)
        SHA256=(${SHA256_PG// / })
        sed -i "s#SHA256#${SHA256}#g" ${package_flag}_${version}_linux.zip.sha256sum
        cd ${WORKSPACE}/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
        echo servicename:${serviceName} >    ${package_flag}_srcbaseline.txt
        echo branch:${branch} >>    ${package_flag}_srcbaseline.txt
        echo commitid:${commit_id} >>    ${package_flag}_srcbaseline.txt
        cp -rf ${package_flag}_srcbaseline.txt ${WORKSPACE}/$1/dist/
        touch ${package_name}.json
        cat>${package_name}.json<<EOF
        {
            "sha256Sum": "${SHA256}",
            "repoInfo": [{
                "repoUrl": "http://gitcode.com/ascend/ascend-deployer.git",
                "repoBranch": "${branch}",
                "commitId": "${commit_id}"
                }],
            "buildTime": "$(date +%Y%m%d%H%M%S)"
        }
EOF
        cp -rf ${package_name}.json ${WORKSPACE}/$1/dist/
        echo "*****************************pr no need uplod obs****************************"
    elif [ "${serviceName}" == "MindCluster-AscendNPUBurn" ];then
        cd ${WORKSPACE}/$1/dist/
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
        for tar_file in *.whl; do
         [ -f "$tar_file" ] || continue
         package_name="$tar_file"
         SHA256_PG=$(sha256sum "${package_name}")
         SHA256=(${SHA256_PG// / })
         echo "${SHA256}  ${package_name}" > "${package_name}.sha256sum"
         cat>${package_name}.json<<EOF
         {
            "sha256Sum": "${SHA256}",
            "repoInfo": [{
                "repoUrl": "http://gitcode.com/ascend/${serviceName}.git",
                "repoBranch": "${branch}",
                "commitId": "${commit_id}"
                }],
            "buildTime": "$(date +%Y%m%d%H%M%S)"
         }
EOF
        echo "*****************************pr no need uplod obs****************************"
        done
    else
        if [ "${serviceName}" == "ascend-for-volcano" ]; then
            servicepackName="volcano"
            rm -rf ${WORKSPACE}/mind-cluster/component/$1/output/Dockerfile-controller ${WORKSPACE}/mind-cluster/component/$1/output/Dockerfile-scheduler
            rm -rf ${WORKSPACE}/mind-cluster/component/$1/output/alpine ${WORKSPACE}/mind-cluster/component/$1/output/openeuler

            echo  "*************${servicepackName}**************************"
        elif [ "${serviceName}" == "ascend-device-plugin" ]; then
            servicepackName="device-plugin"
        elif [ "${serviceName}" == "npu-exporter" ];then
            servicepackName="npu-exporter"
        elif [ "${serviceName}" == "noded" ];then
            servicepackName="noded"
        elif [ "${serviceName}" == "ascend-faultdiag" ];then
            servicepackName="faultdiag"
        else
            servicepackName="${serviceName}"
        fi
        cd ${WORKSPACE}/mind-cluster/component/$1/output/
        package_name=Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip
        zip -r Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip ./*
        echo "SHA256 Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip" > Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip.sha256sum
        SHA256_PG=$(sha256sum Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip)
        SHA256=(${SHA256_PG// / })
        sed -i "s#SHA256#${SHA256}#g" Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip.sha256sum
        cd ${WORKSPACE}/mind-cluster/component/$1
        branch=$(git rev-parse --abbrev-ref HEAD)
        commit_id=$(git rev-parse HEAD)
        echo servicename:${serviceName} >   Ascend-mindxdl-${servicepackName}_srcbaseline.txt
        echo branch:${branch} >>   Ascend-mindxdl-${servicepackName}_srcbaseline.txt
        echo commitid:${commit_id} >>   Ascend-mindxdl-${servicepackName}_srcbaseline.txt
        cp -rf Ascend-mindxdl-${servicepackName}_srcbaseline.txt ${WORKSPACE}/mind-cluster/component/$1/output/
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
        cp -rf ${package_name}.json ${WORKSPACE}/mind-cluster/component/$1/output/
    fi
}


echo "*******************${serviceName}****${WORKSPACE}******************"

if [ "${serviceName}" == "ascend-for-volcano" ]; then
    uploadpackge ${serviceName}
    echo "*****************************pr no need uplod obs****************************"
elif [ "${serviceName}" == "ascend-deployer" ]; then
    echo "*******************${serviceName}**********************"
    uploadpackge ${serviceName}
elif [ "${serviceName}" == "MindCluster-AscendNPUBurn" ]; then
    echo "*******************${serviceName}**********************"
    uploadpackge ${serviceName}
elif [ "${serviceName}" == "ascend-docker-runtime" ]; then
    echo "*******************${serviceName}**********************"
    uploadpackge ${serviceName}
    echo "*****************************pr no need uplod obs****************************"
elif [ "${serviceName}" == "helm-deploy-tool" ]; then
    echo "*******************${serviceName}**********************"
    uploadpackge ${serviceName}
    echo "*****************************pr no need uplod obs****************************"
else
    echo "*******************${serviceName}**********************"
    uploadpackge ${serviceName}
    echo "*****************************pr no need uplod obs****************************"
fi
