### 使用说明

#### 实例地址
https://blog.csdn.net/LOVETEDA/article/details/98881072
#### 证书生成
```
openssl req \
    -newkey rsa:2048 \
    -nodes \
    -days 3650 \
    -x509 \
    -keyout ca.key \
    -out ca.crt \
    -subj "/CN=*"
openssl req \
    -newkey rsa:2048 \
    -nodes \
    -keyout server.key \
    -out server.csr \
    -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=*"
openssl x509 \
    -req \
    -days 365 \
    -sha256 \
    -in server.csr \
    -CA ca.crt \
    -CAkey ca.key \
    -CAcreateserial \
    -out server.crt \
    -extfile <(echo subjectAltName = IP:127.0.0.1)
最后一行把IP地址写入SAN，生成证书
```

##### 安装参考build/install.sh