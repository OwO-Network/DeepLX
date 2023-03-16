<!--
 * @Author: Vincent Young
 * @Date: 2022-10-18 07:32:29
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-03-16 20:24:59
 * @FilePath: /DeepLX/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2022 by Vincent, All Rights Reserved. 
-->
# DeepL X
Permanently free DeepL API written in Golang

## Description
- `deeplx` in only run in port `1188`, later versions will do the specified port.
- `deeplx` is listening to `0.0.0.0:1188` by default.
- `deeplx` is using `DeepL` Free API.
- `deeplx` is unlimited to the number of requests.

## Usage
### Request Parameters
- text: string
- source_lang: string
- target_lang: string

### Response
```json
{
  "alternatives": [
    "Undisputed",
    "Unquestionable",
    "Unquestionably"
  ],
  "code": 200,
  "data": "Undoubtedly",
  "id": 8300079001
}
```

### Run with Docker
```bash
# ghcr.io
docker run -itd -p 1188:1188 ghcr.io/owo-network/deeplx:latest

# dockerhub
docker run -itd -p 1188:1188 missuo/deeplx:latest
```

### Run with Docker Compose
```bash
mkdir deeplx
cd deeplx
wget https://raw.githubusercontent.com/OwO-Network/DeepLX/main/docker-compose.yaml
docker-compose up -d
```

### Run on Linux Server
```bash
bash <(curl -Ls https://cpp.li/deeplx)
```

### Run on Mac
#### Homebrew (Recommended)
```bash
brew tap owo-network/brew
brew install deeplx
brew services start owo-network/brew/deeplx

# Update to the latest version
brew update
brew upgrade deeplx
brew services restart owo-network/brew/deeplx

# View the currently installed version
brew list --versions deeplx
```

#### Manual
1. Download  the latest release of DeepL X.
```bash
sudo mv deeplx_darwin_amd64 /usr/local/bin/deeplx
sudo chmod +x /usr/local/bin/deeplx
```

2. Download the `me.missuo.deeplx.plist` to `~/Library/LaunchAgents`.
```bash
wget https://raw.githubusercontent.com/OwO-Network/DeepLX/main/me.missuo.deeplx.plist -O ~/Library/LaunchAgents/me.missuo.deeplx.plist
```
3. Run following command.
```bash
launchctl load ~/Library/LaunchAgents/me.missuo.deeplx.plist
launchctl start ~/Library/LaunchAgents/me.missuo.deeplx.plist
```

### Install from AUR
```bash
paru -S deeplx-bin
```

After installation, start the daemon with the following command.

```bash
systemctl daemon-reload
systemctl enable deeplx

```
## Setup on [Bob App](https://bobtranslate.com/)
1. Install [bob-plugin-deeplx](https://github.com/missuo/bob-plugin-deeplx) on Bob.

2. Setup the API. (If you use Brew to install locally you can skip this step)
![c5c19dd89df6fae1a256d](https://missuo.ru/file/c5c19dd89df6fae1a256d.png)

## Setup on [immersive-translate](https://github.com/immersive-translate/immersive-translate)
**It is not recommended, because the `immersive-translate` will send many requests in a short time, which will cause the `DeepL API` to block your IP.**

1. Install Latest [immersive-translate ](https://github.com/immersive-translate/immersive-translate/releases) on your browser.

2. Click on **Developer Settings** in the bottom left corner. **Enable Beta experimental features**.

3. Set the URL. (If you are not deploying locally, you need to change 127.0.0.1 to the IP of your server)

![6a48ba28621f2465028f0](https://missuo.ru/file/6a48ba28621f2465028f0.png)

## Backup the Docker Image of zu1k
```shell
docker run -itd -p 1188:80 missuo/deeplx-bk
```
**This docker image is not related to this project, as the original author deleted the image, it is only for backup.**

## Author
**DeepL X** © [DeepL X Contributors](https://github.com/OwO-Network/DeepLX/graphs/contributors), Released under the [MIT](./LICENSE) License.<br>
