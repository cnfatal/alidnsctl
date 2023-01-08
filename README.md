# alidnsctl

A lite tool to operate with Aliyun DNS Records.

## Quick Start

An exampl in ppp `/etc/ppp/ip-up.d/20-ddns.sh` script:

```sh
#!/bin/sh -e

export ACCESS_KEY_ID="<access key id>"
export ACCESS_KEY_SECRET="<access key secret>"

alidnsctl set router.example.com ${IPLOCAL}
```

## Install

Download prebuild binaries from [github release](https://github.com/cnfatal/alidnsctl/releases/latest).

```sh
wget https://github.com/cnfatal/alidnsctl/releases/latest/download/alidnsctl-linux-amd64 -O /usr/bin/alidnsctl
chmod +x /usr/bin/alidnsctl
```

Install from source code:

```sh
go install github.com/cnfatal/alidnsctl@latest
```

## Usage

Set access key via enviroment:

```sh
export ACCESS_KEY_ID=<access key id>
export ACCESS_KEY_SECRET=<access key id>
```

Example update records:

```sh
alidnsctl set api.example.com 127.0.0.1
alidnsctl set --type AAAA api.example.com fe80::b0bb:26ff:fe2b:cb20
alidnsctl set --type CNAME @.example.com www.example.com
alidnsctl set --type TXT txt.example.com hello_world hello_world2
```

For more infomations see help:

```sh
alidnsctl --help
```

List records:

```sh
$ alidnsctl list abc.example.com
[
  {
    "DomainName": "example.com",
    "Line": "default",
    "Locked": false,
    "RR": "abc",
    "RecordId": "xxxxxxxxx",
    "Status": "ENABLE",
    "TTL": 600,
    "Type": "CNAME",
    "Value": "www.example.com",
    "Weight": 1
  }
]
```

Remove Record:

```sh
$ alidnsctl del router.example.com 127.0.0.1 fe80::b0bb:26ff:fe2b:cb20
[] #show left records on rr
```

List Domains:

```sh
$ alidnsctl domains list
[
  {
    "AliDomain": true,
    "CreateTime": "2017-09-07T01:21Z",
    "CreateTimestamp": 1504747278000,
    "DnsServers": {
      "DnsServer": [
        "dns29.hichina.com",
        "dns30.hichina.com"
      ]
    },
    "DomainId": "xxxx-xxxx-xxxx-xxxx-675c9c69e149",
    "DomainName": "example.com",
    "PunyCode": "example.com",
    "RecordCount": 8,
    "ResourceGroupId": "group",
    "Starmark": false,
    "Tags": {},
    "VersionCode": "mianfei",
    "VersionName": "Alibaba Cloud DNS"
  }
]
```
