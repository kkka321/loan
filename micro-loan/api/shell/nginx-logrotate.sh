#!/usr/bin/env bash
# @describe: 切割nginx日志,linux
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x
#nginx_prefix="/home/service/openresty/nginx"
nginx_prefix="/usr/local/openresty/nginx"
dest_base_path="/home/nginx/logsdata"

yesterday=`date -d "yesterday" "+%s"`
day=`date -d @"$yesterday" "+%d"`
year=`date -d @"$yesterday" "+%Y"`
month=`date -d @"$yesterday" "+%m"`
date_str="$year$month$day"
#echo $date_str

# 创建目标目录,以 YYYY/MM 来分开
dest_path="$dest_base_path/$year/$month"
if [ ! -d "$dest_path" ]; then
    mkdir -p "$dest_path"
fi

all_logs=`find $nginx_prefix/logs/ -name "*.log"`
for log in $all_logs; do
    log_name=`echo $log | awk -F "/" '{print $NF}'`
    #echo $log
    #echo $log_name
    rotate_name="$dest_path/$log_name.$date_str"
    mv "$log" "$rotate_name"
done

$nginx_prefix/sbin/nginx -s reopen
# vim:set ts=4 sw=4 et fdm=marker:

