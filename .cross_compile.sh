
###
 # @Author: Vincent Young
 # @Date: 2022-10-20 02:19:06
 # @LastEditors: Vincent Yang
 # @LastEditTime: 2024-03-20 16:52:40
 # @FilePath: /DeepLX/.cross_compile.sh
 # @Telegram: https://t.me/missuo
 # 
 # Copyright © 2022 by Vincent, All Rights Reserved. 
### 
set -e

DIST_PREFIX="deeplx"
DEBUG_MODE=${2}
TARGET_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm64 linux/mips openbsd/amd64 openbsd/arm64 freebsd/amd64 freebsd/arm64 windows/386 windows/amd64"

rm -rf ${TARGET_DIR}
mkdir ${TARGET_DIR}

for pl in ${PLATFORMS}; do
    export GOOS=$(echo ${pl} | cut -d'/' -f1)
    export GOARCH=$(echo ${pl} | cut -d'/' -f2)
    export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}
    if [ "${GOOS}" == "windows" ]; then
        export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}.exe
    fi

    echo "build => ${TARGET}"
    if [ "${DEBUG_MODE}" == "debug" ]; then
        CGO_ENABLED=0 go build -trimpath -gcflags "all=-N -l" -o ${TARGET} \
            -ldflags "-w -s" .
    else
        CGO_ENABLED=0 go build -trimpath -o ${TARGET} \
            -ldflags "-w -s" .
    fi
done
