
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt
```


### 生成完整的证书
创建文件 openssl.conf
```
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localhost

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
```

### 运行命令
```
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
  -keyout server.key -out server.crt -config openssl.cnf -extensions v3_req
```
