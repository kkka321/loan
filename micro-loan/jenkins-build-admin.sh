# Go相关配置项
export GOROOT=/www/go1.10.1
export GOPATH=/www/gopath1.10.1
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

gopath=$GOPATH/src/micro-loan

[ ! -d $gopath ] && mkdir -p $gopath
[ ! -d $gopath ] && exit 1

echo "rm -rf  $gopath/*"
rm -rf  $gopath/*

# 生成最后一次git提交hash
cd "${WORKSPACE}"
git ls-files | while read file; do touch -d $(git log -1 --format="@%ct" "$file") "$file"; done
git_hash=`git rev-parse HEAD`
echo "cp -Rfp ${WORKSPACE}/* $gopath"
cp -Rfp ${WORKSPACE}/* $gopath/
echo "clean up target directory: $WORKSPACE"
rm -rf *
# 初始化
echo "init target directory..."
admin_path="$gopath/admin"
mkdir -p app
cp -Rfp "$admin_path/app/views" app
cp -Rfp "$admin_path/conf" conf
cp -Rfp "$admin_path/ctl.sh" ctl.sh
cp -Rfp "$admin_path/shell" shell
cp -Rfp "$admin_path/static" static
cp -Rfp "$gopath/favicon.ico" favicon.ico
echo "$git_hash" > "$WORKSPACE/conf/git-rev-hash"
echo "init completed"

cd $admin_path
echo `pwd`

godep get

go build -o ${WORKSPACE}/admin main.go
go build -o ${WORKSPACE}/task  cli-task/task.go
go build -o ${WORKSPACE}/schema  cli-task/schema.go

#go build -o ${WORKSPACE}/fix_ticket_add_repay_date hot-fix/fix_ticket_add_repay_date/main.go
#go build -o ${WORKSPACE}/export-tool  cmd/export-tool/main.go
#go build -o ${WORKSPACE}/fix_xendit_va hot-fix/fix_xendit_va/fix_xendit_va.go
#go build -o ${WORKSPACE}/fix-rm_1-order-find hot-fix/fix-rm_1-order-find.go
#go build -o ${WORKSPACE}/fix_coupon_date hot-fix/fix_coupon_date.go
