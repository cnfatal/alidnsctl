# alidnsctl

A lite tool to operate with Aliyun DNS Records.

## Quick Start

An exampl in ppp `/etc/ppp/ip-up.d/20-ddns.sh` script:

```sh
#!/bin/sh -e

export ACCESS_KEY_ID=<access key id>
export ACCESS_KEY_SECRET=<access key id>

alidnsctl records set --domain example.com --type=A --rr=router --value=${IPLOCAL}
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

List records:

```sh
$ alidnsctl records list --domain example.com
[
  {
    "DomainName": "example.com",
    "Line": "default",
    "Locked": false,
    "RR": "blog",
    "RecordId": "xxxxxxxxx",
    "Status": "ENABLE",
    "TTL": 600,
    "Type": "CNAME",
    "Value": "www.example.com",
    "Weight": 1
  }
]
```

Apply Record:

```sh
$ alidnsctl records set --domain example.com --type A --rr local --value 127.0.0.1
{
  "DomainName": "example.com",
  "Line": "default",
  "Locked": false,
  "RR": "local",
  "RecordId": "xxxxxx",
  "RequestId": "xxxxxxxx-xxxx-xxxx-xxxx-1AAC99FBDC9E",
  "Status": "ENABLE",
  "TTL": 600,
  "Type": "A",
  "Value": "127.0.0.1"
}
```

Remove Record:

```sh
$ alidnsctl records remove <RecordId>
{
  "RecordId": "<RecordId>",
  "RequestId": "xxxxxxxx-3265-55D6-AFD7-1C082412AF3F"
}
```
