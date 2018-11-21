#!/usr/bin/env bash
# @describe: 切换服务区域
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x

function usage() {
    echo "$0 region"
    echo "  region: india|indonesia"
    exit 0
}

argc=$#
if ((argc < 1))
then
    usage
fi

cd conf

region=$1
origin_cnf="app-$region.conf"
des_cnf="app.conf"

if [ ! -f "$origin_cnf" ]
then
    echo "config file does not exist, please check it out: $origin_cnf"
    exit 1
fi

rm -vrf "$des_cnf"
ln -s "$origin_cnf" "$des_cnf"

echo "dev switch region success: $des_cnf"
# vim:set ts=4 sw=4 et fdm=marker:

