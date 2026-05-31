###
 # @Author: Vincent Young
 # @Date: 2023-02-12 09:53:21
 # @LastEditors: Vincent Young
 # @LastEditTime: 2023-02-12 10:01:57
 # @FilePath: /DeepLX/install.sh
 # @Telegram: https://t.me/missuo
 # 
 # Copyright © 2023 by Vincent, All Rights Reserved. 
### 

install_deeplx(){
    last_version=$(curl -Ls "https://api.github.com/repos/OwO-Network/DeepLX/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ ! -n "$last_version" ]]; then
        echo -e "${red}Failed to detect DeepLX version, probably due to exceeding Github API limitations.${plain}"
        exit 1
    fi
    echo -e "DeepLX latest version: ${last_version}, Start install..."

    arch=$(uname -m)
    case "${arch}" in
        x86_64 | amd64) file_arch="amd64" ;;
        aarch64 | arm64) file_arch="arm64" ;;
        i386 | i686) file_arch="386" ;;
        mips) file_arch="mips" ;;
        *)
            echo -e "${red}Unsupported architecture: ${arch}${plain}"
            exit 1
            ;;
    esac

    wget -q -N --no-check-certificate -O /usr/bin/deeplx https://github.com/OwO-Network/DeepLX/releases/download/${last_version}/deeplx_linux_${file_arch}

    chmod +x /usr/bin/deeplx
    wget -q -N --no-check-certificate -O /etc/systemd/system/deeplx.service https://raw.githubusercontent.com/OwO-Network/DeepLX/main/deeplx.service
    systemctl daemon-reload
    systemctl enable deeplx
    systemctl start deeplx
    echo -e "Installed successfully, listening at 0.0.0.0:1188"
}
install_deeplx
