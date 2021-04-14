#!/bin/bash
#@auth cl
#@time 20201214

TARGET=httpsproxy

VERSION=`date +%s`
SUPERPATH="/data/supervisor"
ROOTPATH=/data/sa/service
INSTALLPATH=${ROOTPATH}/${TARGET}


function install_supervisor {
    echo -e "\033[32m ################################# 安装检测supervisor服务 #################################\033[0m"
    # 守护进程supervisor安装
    if [ ! -f /usr/local/bin/supervisord ] && [ ! -f /usr/bin/supervisord ]; then
        pip3 install supervisor
        #ln -s /usr/local/python3/bin/supervisord   /usr/bin/supervisord
        #ln -s /usr/local/python3/bin/supervisorctl /usr/bin/supervisorctl
    else
        echo -e "\033[33m /usr/bin/supervisord had exist \033[0m"
    fi
    echo -e "\033[32m === /usr/bin/supervisord install success === \033[0m"

    # 守护进程supervisor配置
    if [ ! -d "/etc/supervisor/" ];then
        mkdir -m 755 -p /etc/supervisor/
    else
        echo -e "\033[33m /etc/supervisor/ had exist \033[0m"
    fi
    cp supervisord.conf /etc/supervisor/

    # 守护进程supervisor log目录
    if [ ! -d "/etc/supervisor/log" ];then
        mkdir -m 755 -p /etc/supervisor/log
    else
        echo -e "\033[33m /etc/supervisor/log had exist \033[0m"
    fi

    if [ ! -d "/var/run/" ];then
        mkdir -m 777 -p /var/run
    else
        echo -e "\033[33m /var/run/ had exist \033[0m"
    fi

    if [ ! -d "/var/log/" ];then
        mkdir -m 777 -p /var/log
    else
        echo -e "\033[33m /var/log/ had exist \033[0m"
    fi

    # 开机启动配置
    if [ ! -f /usr/lib/systemd/system/supervisord.service ]; then
        cp supervisord.service /usr/lib/systemd/system/
        systemctl enable supervisord
        systemctl is-enabled supervisord
    else
        cp supervisord.service /usr/lib/systemd/system/
        systemctl daemon-reload
        echo -e "\033[33m /usr/lib/systemd/system/supervisord.service had exist \033[0m"
    fi
    echo -e "\033[32m === 开机启动配置 load success === \033[0m"
}

function check_process() {
    count=`ps -ef |grep $1 |grep -v "grep" |wc -l`
    if [ 0 == $count ];then
        return 0
    fi
    return 1
}

function install_service() {
    SAAS_COMMON_CONFIG=$1
    if [ ! -d ${SUPERPATH}/log ]; then
        mkdir -p ${SUPERPATH}/log
    fi
    if [ -f ${SUPERPATH}/${TARGET}.ini ]; then
        mv ${SUPERPATH}/${TARGET}.ini ${SUPERPATH}/${TARGET}.ini_bak_${VERSION}
        echo -e "\033[33m ${SUPERPATH}/${TARGET}.ini had exist \033[0m"
    fi
    echo -e "\033[33m modify the SAAS_COMMON_CONFIG=${SAAS_COMMON_CONFIG} \033[0m"
    sed -i "s/environment=SAAS_COMMON_CONFIG=\w\+/environment=SAAS_COMMON_CONFIG=${SAAS_COMMON_CONFIG}/g" ./server/${TARGET}.ini
    cp ./server/${TARGET}.ini ${SUPERPATH}
    cat ./server/${TARGET}.ini
    echo
    echo -e "\033[32m === ${TARGET}.ini load success === \033[0m"
}

function install_all() {
    #supervisor安装、配置加载
    echo -e "\033[32m ################################# 安装supervisor服务 #################################\033[0m"
    install_supervisor

    echo -e "\033[32m ################################# 安装${TARGET}服务 #################################\033[0m"
    install_service $1
     #启动supervisor
    check_process supervisord
    if [ $? == 0 ]; then
        supervisord -c /etc/supervisor/supervisord.conf
    else
        echo -e "\033[33m supervisord had exist! \033[0m"
    fi
    ps -ef|grep supervisord
    echo -e "\033[32m #################### install supervisor done #####################\033[0m"
}

function start_supervisor() {
    echo -e "\033[32m ################################# 启动supervisor服务 #################################\033[0m"
    check_process supervisord
    if [ $? == 0 ]; then
        supervisord -c /etc/supervisor/supervisord.conf
    else
        echo -e "\033[33m supervisord had exist! \033[0m"
    fi
    ps -ef|grep supervisord
    echo -e "\033[32m #################### start supervisor done #####################\033[0m"
}

function print_help {
    echo -e "\033[35m ######################### 帮助 ######################### \033[0m"
    echo -e "\033[35m #sh install.sh {param} \033[0m"
    echo -e "\033[35m {param}: \033[0m"
    echo -e "\033[35m        -d         : 安装${TARGET} supervisor服务dev配置文件 \033[0m"
    echo -e "\033[35m        -t         : 安装${TARGET} supervisor服务test配置文件 \033[0m"
    echo -e "\033[35m        -p         : 安装${TARGET} supervisor服务prod配置文件 \033[0m"
    echo -e "\033[35m        -s         : 启动supervisor服务 \033[0m"
    echo -e "\033[35m ######################### 帮助 ######################### \033[0m"
}

function main() {
    case $1 in
        "-d")
            install_all dev
        ;;
        "-t")
            install_all test
        ;;
        "-p")
            install_all prod
        ;;
        "-s")
            start_supervisor
        ;;
        *)
            print_help
        ;;
        esac
}

# shellcheck disable=SC2068
main $@