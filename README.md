<!--
 * @Author: Vincent Young
 * @Date: 2022-10-18 07:32:29
 * @LastEditors: Vincent Young
 * @LastEditTime: 2023-02-12 09:29:40
 * @FilePath: /DeepLX/README.md
 * @Telegram: https://t.me/missuo
 * 
 * Copyright © 2022 by Vincent, All Rights Reserved. 
-->
# DeepLX
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
  "code": 200,
  "data": "Hello world",
  "id": 8305092005
}
```

### Run on Mac
1. Download  the latest release of DeepLX.
```bash
sudo mv deeplx_darwin_amd64 /usr/local/bin/deeplx
```

2. Download the `me.missuo.deeplx.plist` to `/Users/YOUR_USERNAME/Library/LaunchAgents`.
```bash
wget https://raw.githubusercontent.com/OwO-Network/DeepLX/main/me.missuo.deeplx.plist -O /Users/YOUR_USERNAME/Library/LaunchAgents/me.missuo.deeplx.plist
```
3. Run following command.
```bash
launchctl load /Library/LaunchAgents/me.missuo.deeplx.plist
launchctl start /Library/LaunchAgents/me.missuo.deeplx.plist
```

## Setup on [Bob App](https://bobtranslate.com/)
1. Install [bob-plugin-deeplx](https://github.com/clubxdev/bob-plugin-deeplx) on Bob.

2. Setup the API.
![9a75c26ad6e8bd9b7582c](https://telegraph.eowo.us/file/9a75c26ad6e8bd9b7582c.png)

## Docker Backup for zu1k
```shell
docker run -itd -p 7373:80 missuo/deeplx-bk
```
## Author
**DeepL X** © [Vincent Young](https://github.com/missuo) & [Leo Shen](https://github.com/sjlleo), Released under the [MIT](./LICENSE) License.<br>
