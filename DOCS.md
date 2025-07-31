# Docs

## `config.yaml` 配置说明

### 16图床

若启用需配置`token`，其值随意，仅作为删除图片的密钥。

## 美团图床

若启用需配置`token`，如何获取见下图：

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1752412071/vqnzbnhwre01xe5napk9.png)

## Cloudinary

若启用需配置`cloud_name` `api_key` `api_secret`。首先前往[官网](https://cloudinary.com)注册，然后进入设置复制以上字段：

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1752412601/dwwwmgiiffldhshftfnv.png)

## NodeSeek图床

若启用需配置`token` 。你需要拥有[NodeSeek](https://www.nodeseek.com)账号，然后前往[NodeImage](https://www.nodeimage.com)登录账号，在API页面复制key即可。

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1752412870/akay7qukatenrhh4dljr.png)

## EroLabs

若启用需配置`token`。访问[EroLabs官网](https://game.ero-labs.cool)，注册/登录账号，随机抓取请求获取Cookie：

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1752415543/pnqy0dpj2ojlih1vm2mn.png)

## DeepSider

若启用需配置`token`。安装[DeepSider](https://microsoftedge.microsoft.com/addons/detail/minfmdkpoboejckenbchpjbjjkbdebdm)浏览器扩展，打开开发者工具，按如图所示抓取`Authorization`即可，注意不带`Bearer`字段。

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1753019709/srlttk17euh6p5uzm2wh.png)

## 极客侧边栏图床

若启用需配置`token`。安装[极客侧边栏](https://www.geeksidebar.com)浏览器扩展，打开开发者工具，按下图获取`Authorization`，注意不带`Bearer`字段。

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1753937392/m9fnc9rjn0xyrfcfa9vv.png)

## Qwen图床

若启用需配置`token`。访问[Qwen Chat](https://chat.qwen.ai/)，打开开发者工具，上传一张图片，按下图所示在网络请求中抓取`bx-umidtoken`请求头值。

![](https://i0.wp.com/res.cloudinary.com/dyxhgk4ga/image/upload/v1753942573/gdfpjrcfhe3um6piei0h.png)