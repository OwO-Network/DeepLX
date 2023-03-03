<!--
 * @Author: Vincent Young
 * @Date: 2022-10-18 07:32:29
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-03-03 02:30:52
 * @FilePath: /DeepLX/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2022 by Vincent, All Rights Reserved. 
-->
# DeepL X
Permanently free DeepL API written in Golang

## Description
- `deeplx` is listening to `0.0.0.0:1188` by default, you can change it to what you want.
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
  "code": 200,
  "data": "Hello world",
  "id": 8305092005
}
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

### Configuration

You can change the default configuration in command line, such as setting the port:
```bash
deeplx_linux_amd64 --port 27001
#or shorthand
deeplx_linux_amd64 -p 27001
```

## Setup on [Bob App](https://bobtranslate.com/)
1. Install [bob-plugin-deeplx](https://github.com/clubxdev/bob-plugin-deeplx) on Bob.

2. Setup the API.
![c5c19dd89df6fae1a256d](https://missuo.ru/file/c5c19dd89df6fae1a256d.png)

## Setup on [immersive-translate](https://github.com/immersive-translate/immersive-translate)
1. Install Latest [immersive-translate ](https://github.com/immersive-translate/immersive-translate/releases) on your browser.

2. Click on **Developer Settings** in the bottom left corner. **Enable Beta experimental features**.

3. Set the URL.
![0779ecf8c7d7d1bee532b](https://missuo.ru/file/0779ecf8c7d7d1bee532b.png)

## Backup the Docker Image of zu1k
```shell
docker run -itd -p 1188:80 missuo/deeplx-bk
```
**This docker image is not related to this project, as the original author deleted the image, it is only for backup.**

## Author
**DeepL X** © [DeepL X Contributors](https://github.com/OwO-Network/DeepLX/graphs/contributors), Released under the [MIT](./LICENSE) License.<br>
