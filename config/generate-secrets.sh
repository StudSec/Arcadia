#!/usr/bin/env bash
# This script generates secret files for the setup and CTFd configuration.

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Make sure we are in the right directory
cd "$(dirname "$0")"

if ! command -v openssl &> /dev/null
then
    echo -e "${RED}[!] openssl could not be found, please install it first.${NC}"
    exit 1
fi

echo "[-] Generating secret files..."
if [ ! -f ctfd/.admin-password ]; then
    openssl rand -base64 20 > ctfd/.admin-password
    echo -e "${GREEN}[+] Generated ctfd/.admin-password file.${NC}"
else
    echo -e "${GREEN}[!] ctfd/.admin-password file already exists, skipping...${NC}"
fi

# Generate id_ed25519 if it doesn't exist
if [ ! -f ssh/id_ed25519 ]; then
    ssh-keygen -t ed25519 -f ssh/id_ed25519 -N "" > /dev/null
    echo -e "${GREEN}[+] Generated ssh/id_ed25519 key pair.${NC}"
else
    echo -e "${GREEN}[!] ssh/id_ed25519 key pair already exists, skipping...${NC}"
fi
