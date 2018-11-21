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
api_path="$gopath/api"
cp -Rfp "$api_path/conf" conf
cp -Rfp "$api_path/shell" shell
cp -Rfp "$gopath/favicon.ico" favicon.ico
echo "$git_hash" > "$WORKSPACE/conf/git-rev-hash"
echo "init completed"

cd $api_path
echo `pwd`

godep get

go build -o ${WORKSPACE}/api main.go
