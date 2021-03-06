#!/bin/sh
#@auth cl
#@time 20201214
# export CI_BUILD_REF_NAME=develop/1.0.0

TARGET=httpsproxy

WORKSPACEPATH=$(cd ../`dirname $0`; pwd)
cd ${WORKSPACEPATH}
SERVERDIR=${WORKSPACEPATH}
TARGETPATH=${WORKSPACEPATH}/build/target/${TARGET}


function make {
    SRC_PATH=${SERVERDIR}
    TARGET=$1
    DST_PATH=${TARGETPATH}
    if [ ! ${TARGET} ]; then
        echo -e "\033[31m [make] make ${TARGET} is nil! \033[0m"
        return 1
    fi

    cd ${SRC_PATH}
    TARGET_FILE=${SRC_PATH}/${TARGET}
    if [ -f ${TARGET_FILE} ]; then
        rm -rf ${TARGET_FILE}
    fi
    echo -e "\033[33m === start make ${TARGET_FILE} ... \033[0m"
    go build
    if [ $? != 0 ]; then
        echo -e "\033[31m [make] make ${TARGET} is failed! \033[0m"
        return 1
    fi
    md5sum ${SRC_PATH}/${TARGET}
    if [ ! -f ${TARGET_FILE} ]; then
        echo -e "\033[31m [make] make file ${TARGET_FILE} is not exits! \033[0m"
        return 1
    fi
    copy_to_target ${TARGET_FILE} ${DST_PATH}
    if [ $? != 0 ]; then
        echo -e "\033[31m [make] copy ${TARGET} is failed! \033[0m"
        return 1
    fi
    md5sum ${DST_PATH}/${TARGET}
    echo -e "\033[32m +++ [make] make and copy ${TARGET} is success! \033[0m"
    return 0
}

function copy_to_target {
    SRC_FILE=$1
    DST_PATH=$2
    DST_FILE=${DST_PATH}/${TARGET}
    if [ ! -d ${DST_PATH} ]; then
        mkdir -p ${DST_PATH}
    else
        if [ -f ${DST_FILE} ]; then
            rm -rf ${DST_FILE}
        fi
    fi
    cp ${SRC_FILE} ${DST_PATH}
    if [ $? != 0 ]; then
        return 1
    fi
    return 0
}

function main {
    echo -e "\033[34m ################### 克隆protocol文件 ################### \033[0m"
    clone_public
    echo -e "\033[34m ######################### 编译 ######################### \033[0m"
    make ${TARGET}

    cp -rf ${SERVERDIR}/conf ${TARGETPATH}
    cp ${SERVERDIR}/build/install.sh ${SERVERDIR}/build/target/
    cp ${SERVERDIR}/build/readme.md ${SERVERDIR}/build/target/
    cp ${SERVERDIR}/build/appversion ${SERVERDIR}/build/target/
    # 设定打包时间
    pack_date=`date +%F_%T_%N`
    sed -i "s/Build\w\+/Build$pack_date/g" ${SERVERDIR}/build/target/appversion
    echo -e "\033[33m === ${SERVERDIR}/build/target/appversion  \033[0m"
    cat ${SERVERDIR}/build/target/appversion
    echo

    TOOLSPATH=${SERVERDIR}/build/target/tools/
    if [ ! -d ${TOOLSPATH} ]; then
        mkdir -p ${TOOLSPATH}
    fi
    cp -rf ${SERVERDIR}/build/tools/* ${TOOLSPATH}

    SUPERPATH=${SERVERDIR}/build/target/supervisor/
    if [ ! -d ${SUPERPATH} ]; then
        mkdir -p ${SUPERPATH}
    fi
    cp -rf ${SERVERDIR}/build/supervisor/* ${SUPERPATH}

    SSLPATH=${SERVERDIR}/build/target/ssl/
    if [ ! -d ${SSLPATH} ]; then
        mkdir -p ${SSLPATH}
    fi
    cp -rf ${SERVERDIR}/build/ssl/* ${SSLPATH}
}