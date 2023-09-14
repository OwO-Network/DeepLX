<!--
 * @Author: Vincent Young
 * @Date: 2022-10-18 07:32:29
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-09-14 13:58:57
 * @FilePath: /DeepLX/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2022 by Vincent, All Rights Reserved. 
-->
<h1 align="center">
  <br>DeepL X<br>
</h1>
<h4 align="center">Permanently free DeepL API written in Golang.</h4>
<p align="center">
  <a href="https://goreportcard.com/report/github.com/OwO-Network/DeepLX">
    <img src="https://goreportcard.com/badge/github.com/OwO-Network/DeepLX?style=flat-square">
  </a>
  <a href="https://github.com/OwO-Network/DeepLX/releases">
    <img src="https://img.shields.io/github/release/OwO-Network/DeepLX/all.svg?style=flat-square">
  </a>
</p>

## **Related Projects**
[OwO-Network/PyDeepLX](https://github.com/OwO-Network/PyDeepLX): Python Package for DeepLX.

[OwO-Network/gdeeplx](https://github.com/OwO-Network/gdeeplx): Golang Package for DeepLX.

## Discussion Group
[Telegram Group](https://t.me/+8KDGHKJCxEVkNzll)

## Description
- `DeepLX` is listening to `0.0.0.0:1188` by default. You can modify the listening port by yourself.
- `DeepLX` is using `DeepL` Free API.
- `DeepLX` is unlimited to the number of requests.

## Usage
### Request Parameters
- text: string
- source_lang: string
- target_lang: string

### Response
```json
{
  "alternatives": [
    "no one else",
    "there is no other person (idiom); there is no one else",
    "there is no other person"
  ],
  "code": 200,
  "data": "there is no one else",
  "id": 8352115005,
  "source_lang": "ZH",
  "target_lang": "EN"
}
```
### Specify the port
**Thanks to [cijiugechu](https://github.com/cijiugechu) for [his contribution](https://github.com/OwO-Network/DeepLX/commit/4a0920579ea868b0f05ccdff6bceae316bfd5dc8) to make this feature possible for this project!**
```bash
./deeplx -p 3333
# or
./deeplx -port 3333
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
# docker compose v1
docker-compose up -d
# or docker compose v2
docker compose up -d
```

### Run on Linux Server
```bash
bash <(curl -Ls https://raw.githubusercontent.com/OwO-Network/DeepLX/main/install.sh)
# or
bash <(curl -Ls https://cpp.li/deeplx)
```

### Run on Mac
#### Homebrew (Recommended)
**Homebrew has been fixed in the latest version and works perfectly.**
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

## Use in Python
```python
import httpx, json

deeplx_api = "http://127.0.0.1:1188/translate"

data = {
	"text": "Hello World",
	"source_lang": "EN",
	"target_lang": "ZH"
}

post_data = json.dumps(data)
r = httpx.post(url = deeplx_api, data = post_data).text
print(r)
```

## Backup the Docker Image of zu1k
```shell
docker run -itd -p 1188:80 missuo/deeplx-bk
```
**This docker image is not related to this project, as the original author deleted the image, it is only for backup.**

## Author
**DeepL X** © [DeepL X Contributors](https://github.com/OwO-Network/DeepLX/graphs/contributors), Released under the [MIT](./LICENSE) License.<br>
