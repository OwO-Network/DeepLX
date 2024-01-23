<!--
 * @Author: Vincent Young
 * @Date: 2022-10-18 07:32:29
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-11-28 00:24:20
 * @FilePath: /DeepLX/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2022 by Vincent, All Rights Reserved. 
-->

[![GitHub Workflow][1]](https://github.com/OwO-Network/DeepLX/actions)
[![Go Version][2]](https://github.com/OwO-Network/DeepLX/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/OwO-Network/DeepLX)
[![Maintainability][4]](https://codeclimate.com/github/OwO-Network/DeepLX/maintainability)
[![GitHub License][5]](https://github.com/OwO-Network/DeepLX/blob/main/LICENSE)
[![Docker Pulls][6]](https://hub.docker.com/r/missuo/deeplx)
[![Releases][7]](https://github.com/OwO-Network/DeepLX/releases)

[1]: https://img.shields.io/github/actions/workflow/status/OwO-Network/DeepLX/release.yaml?logo=github
[2]: https://img.shields.io/github/go-mod/go-version/OwO-Network/DeepLX?logo=go
[3]: https://goreportcard.com/badge/github.com/OwO-Network/DeepLX
[4]: https://api.codeclimate.com/v1/badges/b5b30239174fc6603aca/maintainability
[5]: https://img.shields.io/github/license/OwO-Network/DeepLX
[6]: https://img.shields.io/docker/pulls/missuo/deeplx?logo=docker
[7]: https://img.shields.io/github/v/release/OwO-Network/DeepLX?logo=smartthings

## **Related Projects**
| Link                                                                | Description                         | Maintainer                                                      |
| ------------------------------------------------------------------- | ----------------------------------- | --------------------------------------------------------------- |
| [OwO-Network/PyDeepLX](https://github.com/OwO-Network/PyDeepLX)     | Python Package for DeepLX           | [Vincent Yang](https://github.com/missuo)                       |
| [OwO-Network/gdeeplx](https://github.com/OwO-Network/gdeeplx)       | Golang Package for DeepLX           | [Vincent Yang](https://github.com/missuo)                       |
| [ifyour/deeplx](https://github.com/ifyour/deeplx)                   | JS Package for DeepLX (JavaScript)  | [ifyour](https://github.com/ifyour)                             |
| [deeplx-serverless](https://github.com/LegendLeo/deeplx-serverless) | Serverless Package for DeepLX       | [LegendLeo](https://github.com/LegendLeo/deeplx-serverless)     |

## Discussion Group
[Telegram Group](https://t.me/+8KDGHKJCxEVkNzll)

## Description
- `DeepLX` is listening to `0.0.0.0:1188` by default. You can modify the listening port by yourself.
- `DeepLX` is using `DeepL` Free API.
- `DeepLX` is unlimited to the number of requests.

## Usage
### For Developer
#### Request Parameters
- text: string
- source_lang: string
- target_lang: string

#### Response
```json
{
  "alternatives": [
    "Did you hear about this?",
    "You've heard about this?",
    "You've heard of this?"
  ],
  "code": 200,
  "data": "Have you heard about this?",
  "id": 8356681003,
  "method": "Free",
  "source_lang": "ZH",
  "target_lang": "EN"
}
```
### Custom Options
**The following settings are optional and not required.**
- `-port` or `-p` : Listening port. Default is `1188`.
- `-token` : Access token. If you have set it up, each request will need to include an `Authorization` header.
- `-authkey` : DeepL Official `AuthKey`. If you have set it up, after the 429 response, the official AuthKey will be used for the request. If multiple authKeys are used simultaneously, they need to be separated by commas.

#### Requesting a token-protected **DeepLX API** instance using the `curl`
```
curl -X POST http://localhost:1188/translate \
-H "Content-Type: application/json" \
-H "Authorization: Bearer your_access_token" \
-d '{
    "text": "Hello, world!",
    "source_lang": "EN",
    "target_lang": "DE"
}'
```

### Run with Docker
```bash
# ghcr.io
docker run -itd -p 1188:1188 ghcr.io/owo-network/deeplx:latest

# custom environment variables
docker run -itd -p 1188:1188 -e "TOKEN=helloxxx" -e "AUTHKEY=xxxx:fx" ghcr.io/owo-network/deeplx:latest

# dockerhub
docker run -itd -p 1188:1188 missuo/deeplx:latest

# custom environment variables
docker run -itd -p 1188:1188 -e "TOKEN=helloxxx" -e "AUTHKEY=xxxx:fx" missuo/deeplx:latest
```

### Run with Docker Compose
```bash
mkdir deeplx
cd deeplx
wget https://raw.githubusercontent.com/OwO-Network/DeepLX/main/compose.yaml
# modify environment variables
# environment:
  # - TOKEN=helloxxx
  # - AUTHKEY=xxxxxxx:fx
# docker compose
docker compose up -d
```

### Run on Linux Server
```bash
bash <(curl -Ls https://raw.githubusercontent.com/OwO-Network/DeepLX/main/install.sh)
# or
bash <(curl -Ls https://qwq.mx/deeplx)
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

### Run on Sealos

[![](https://raw.githubusercontent.com/labring-actions/templates/main/Deploy-on-Sealos.svg)](https://cloud.sealos.io/?openapp=system-template%3FtemplateName%3Ddeeplx)

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

## Acknowledgements

### Contributors

<a href="https://github.com/OwO-Network/DeepLX/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=OwO-Network/DeepLX&anon=0" />
</a>

## Activity
![Alt](https://repobeats.axiom.co/api/embed/5f473f85db27cb30028a2f3db7a560f3577a4860.svg "Repobeats analytics image")

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FOwO-Network%2FDeepLX.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FOwO-Network%2FDeepLX?ref=badge_large)
