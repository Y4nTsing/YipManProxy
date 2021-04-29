# YipManProxy
适用于整合多个云函数的反向socks5代理工具

具体工作原理：

![image](https://raw.githubusercontent.com/Y4nTsing/YipManProxy/main/%E8%AF%B4%E6%98%8E%E5%9B%BE.png)

使用说明：
```
yipmanproxy -m server -p port1 -ccp port2 -key xxxx(max len 32)
-m server：服务端模式
-p port1：指服务器开放的socks5端口
-ccp port2：指用于客户服务端连接的端口
-key xxxx：指用于认证客户服务端接入的密码，最长32位，需要和客户端保持一致


yipmanproxy -m client -r x.x.x.x:port2 -u foo -pwd bar -key xxxx(max len 32)
-m client：客户端模式
-r x.x.x.x:port2：指服务端开放的ccp地址
-u foo：socks5用户名
-pwd：socks5密码
-key xxxx：指用于认证客户服务端接入的密码，最长32位，需要和服务端保持一致
```
之后可以通过访问服务端的port1使用代理，如：
```
curl --proxy "socks5://foo:bar@x.x.x.x:port1" target.com
```
将会使用随机一个客户服务端作为出口，如果没有客户服务端的话，就无法正常使用。

使用场景是比如将多个区域的云函数或者多个厂商的云函数整合到一起形成代理池，省去配置负载均衡和认证的环节。
