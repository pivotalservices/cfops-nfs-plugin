# cfops-nfs-plugin

### Overview

cfops plugin to back up and internal NFS server given a productName and jobName

### Install

Download the latest version here:
https://github.com/pivotalservices/cfops-nfs-plugin/releases/latest

### Contributing

PRs welcome.


### Usage

**backup command**

```
NAME:
   cfops - backup

USAGE:
   ./cfops backup [ standard arguments...] --tile nfs-tile --pluginargs "--productName <productName> --jobName <jobName>"


```

**restore command**

```
NAME:
   cfops - restore

USAGE:
   ./cfops restore [ standard arguments...] --tile nfs-tile --pluginargs "--productName <productName> --jobName <jobName>"


```
