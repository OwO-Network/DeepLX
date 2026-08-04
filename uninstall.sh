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

uninstall_dlx() {
    echo -e "${green}Starting DLX uninstallation...${plain}"

    # 1. Stop and disable the deeplx service (historical unit name)
    if systemctl is-active --quiet deeplx; then
        echo -e "${yellow}Stopping deeplx service...${plain}"
        systemctl stop deeplx
    else
        echo -e "${yellow}deeplx service is not running or not found.${plain}"
    fi

    if systemctl is-enabled --quiet deeplx; then
        echo -e "${yellow}Disabling deeplx service from starting on boot...${plain}"
        systemctl disable deeplx
    else
        echo -e "${yellow}deeplx service is not enabled.${plain}"
    fi

    # 2. Remove the systemd service file
    if [ -f /etc/systemd/system/deeplx.service ]; then
        echo -e "${yellow}Removing deeplx systemd service file (/etc/systemd/system/deeplx.service)...${plain}"
        rm -f /etc/systemd/system/deeplx.service
        systemctl daemon-reload
        echo -e "${green}Systemd daemon reloaded.${plain}"
    else
        echo -e "${yellow}deeplx systemd service file not found, skipping removal.${plain}"
    fi

    # 3. Remove the executable (release binary name remains deeplx)
    if [ -f /usr/bin/deeplx ]; then
        echo -e "${yellow}Removing deeplx executable (/usr/bin/deeplx)...${plain}"
        rm -f /usr/bin/deeplx
    else
        echo -e "${yellow}deeplx executable not found, skipping removal.${plain}"
    fi

    echo -e "${green}DLX uninstallation complete.${plain}"
    echo -e "${green}If you wish to reinstall, please run the install script again.${plain}"
}

uninstall_dlx
