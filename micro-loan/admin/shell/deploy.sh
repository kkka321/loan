#!/usr/bin/env bash
# @describe: 切换端口,实现平滑重启服务
# 目前只考虑单机部署情况,后续分布式了,大思想是一致
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x

admin_current_file="/home/nginx/admin-current"
current_work=`cat ~/admin-current`
if [ "x$current_work" != "x8600" ] && [ "x$current_work" != "x8601" ]
then
    echo "当前工作端口不正确,请检查"
    exit 1
fi

if [ "x$current_work" == "x8600" ]
then
    switch_work=8601
else
    switch_work=8600
fi

# 找到未提供服务的服务
prog_title="admin:$switch_work"

function getPid() {
    pid=`ps aux | grep $prog_title | grep -v grep | awk '{print $2}'`
}

getPid

if [ -z "$pid" ]
then
    echo "$prog_title 工作进程pid不存在,请检查"
    exit 2
fi

# 部署和切换项目
echo "部署最新的程序,切换配置文件"
work_space="/data/micro-loan"
origin="$work_space/admin"
dest="$origin-$switch_work"
backup="$dest.bak"
rm -rf "$backup" && mv $dest "$backup"
cp -Rf "$origin" "$dest"
online_conf="$dest/conf/app.prod.$switch_work.conf"
dest_conf="$dest/conf/app.conf"
cp -vf "$online_conf" "$dest_conf"

#exit

echo "将要 kill $prog_title,并休眠5秒"
kill -9 $pid

for ((i = 1; i <= 5; i++))
do
    echo -n "....."
    sleep 1
done

getPid
echo ""

if [ -z "$pid" ]
then
    echo "$prog_title 重启失败,请确认"
    exit 3
fi

echo "$prog_title 重启成功"

nginx="/usr/local/openresty/nginx/sbin/nginx"
nginx_conf="/usr/local/openresty/nginx/conf/conf.d"
work_conf="$nginx_conf/admin.conf.$switch_work"
admin_conf="$nginx_conf/admin.conf"
cp -vf "$work_conf" "$admin_conf"
sudo $nginx -t && sudo $nginx -s reload
if [ $? != 0 ]
then
    echo "重启 nginx 失败"
    exit 4
fi

echo $switch_work > "$admin_current_file"
echo "平滑部署完成"
# vim:set ts=4 sw=4 et fdm=marker:

