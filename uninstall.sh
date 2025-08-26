#!/bin/bash

# Colors
red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
    echo -e "${red}This script must be run as root.${plain}"
    exit 1
fi

uninstall_deeplx() {
    echo -e "${green}Starting DeepLX uninstallation...${plain}"

    # 1. Stop and disable the DeepLX service
    if systemctl is-active --quiet deeplx; then
        echo -e "${yellow}Stopping DeepLX service...${plain}"
        systemctl stop deeplx
    else
        echo -e "${yellow}DeepLX service is not running or not found.${plain}"
    fi

    if systemctl is-enabled --quiet deeplx; then
        echo -e "${yellow}Disabling DeepLX service from starting on boot...${plain}"
        systemctl disable deeplx
    else
        echo -e "${yellow}DeepLX service is not enabled.${plain}"
    fi

    # 2. Remove the systemd service file
    if [ -f /etc/systemd/system/deeplx.service ]; then
        echo -e "${yellow}Removing DeepLX systemd service file (/etc/systemd/system/deeplx.service)...${plain}"
        rm -f /etc/systemd/system/deeplx.service
        systemctl daemon-reload
        echo -e "${green}Systemd daemon reloaded.${plain}"
    else
        echo -e "${yellow}DeepLX systemd service file not found, skipping removal.${plain}"
    fi

    # 3. Remove the DeepLX executable
    if [ -f /usr/bin/deeplx ]; then
        echo -e "${yellow}Removing DeepLX executable (/usr/bin/deeplx)...${plain}"
        rm -f /usr/bin/deeplx
    else
        echo -e "${yellow}DeepLX executable not found, skipping removal.${plain}"
    fi

    echo -e "${green}DeepLX uninstallation complete.${plain}"
    echo -e "${green}If you wish to reinstall, please run the install script again.${plain}"
}

uninstall_deeplx
