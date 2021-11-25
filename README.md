# FS3 Quickstart Guide
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)

- Join us on our [public Slack channel](https://filswan.slack.com) for news, discussions, and status updates. 
- [Check out our medium](https://filswan.medium.com) for the latest posts and announcements.

## How to use
### Prerequisite
- Golang 1.15+.
- Node Js 14.0+.

# Install from Source
## Checkout source code
```
git clone https://github.com/filswan/fs3
cd fs3
git checkout <release_branch>
```

## Build the Source Code
#### Build UI
```bash
cd browser
npm install
npm run release
```
#### Install Filecoin dependency
```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```
#### Install go module dependency
``` bash 
cd ..
git submodule update --init --recursive
make ffi
```

#### Set up FS3 configuration
Set up and customize FS3 configuration by making modifications on `.env` file, which stores your information as environment variables. An example config is given as `.env.example` for reference. 
``` bash
vim .env
```

Modify the `.env` file based on your use cases:

* __SWAN_ADDRESS__ : The address of filswan platform, default as `https://api.filswan.com`.
* __FS3_VOLUME_ADDRESS__ : The address of FS3 VOLUME, default as `~/minio-data`. If changed, the FS3 server start command has to be changed accordingly.
* __FS3_WALLET_ADDRESS__ : A wallet address is a must for sending deals to miner. 
* __CAR_FILE_SIZE__ : A fixed car file size in bytes need to be predefined before generating car files for trunk via variable `CarFileSize`, such as `8589934592` for 8Gb as default.
* __IPFS_API_ADDRESS__ :  An available ipfs address with port need to be set up. For example, `https://MyIpfsUrl:Port`.
* __IPFS_GATEWAY__ :  An available ipfs address with port need to be set up for file downloading. For example, `https://MyIpfsGatewayUrl:Port`.
* __SWAN_TOKEN__ : A valid swan token is required for posting task on swan platform. It can be received after creating an account on [Filswan](https://www.filswan.com). Check [Filswan APIs](https://documenter.getpostman.com/view/13140808/TWDZJbzV) for more details on how to get authorization token.

If the configuration is changed in the future, build up the FS3 server again to make the changes take effect.

#### Build up FS3 server
``` bash
make
```

## Run a Standalone FS3 Server
``` bash
 ./minio server ~/minio-data
```

The default FS3 volume address `Fs3VolumeAddress` is set as `~/minio-data`, which can be changed in `.env`. If the volume address is changed in the future, build up the FS3 server again to make the changes take effect.



The FS3 deployment starts using default root credentials `minioadmin:minioadmin`. You can test the deployment using the FS3 Browser, an embedded
web-based object browser built into FS3 Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, send deals, retrieve data and browse the contents of the FS3 server.

You can also connect using any S3-compatible tool, such as the FS3 `mc` commandline tool.


## FS3 API
### Get FS3 API Token
FS3 APIs are designed to do verification before performing any actions for safety consideration. An FS3 API token is generated from FS3 login API.

POST `minio/webrpc`
#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

# Body
{
    "id": 1,
    "jsonrpc": "2.0",
    "method": "web.Login",
    "params":{
        "username": "minioadmin", 
        "password": "minioadmin"
    }
}
```
Response from POSTMAN
```bash
{
    "jsonrpc": "2.0",
    "result": {
        "token": "eyJhbGc5cCI6IkpXVCJ9.eyJhY2Nlc3NLIiwiZXhwIjoxMJEuksJBALDYXbw9K",
        "uiVersion": "2021-07-31T03:07:17Z"
    },
    "id": 1
}
```

### Send Online Deals (single file)
POST `minio/deal/{bucket}/{object}`

#### Example: 

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
    "VerifiedDeal":"false",
    "FastRetrieval":"true",
    "MinerId":"t00000",
    "Price": "0.000005",
    "Duration":"1036800"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "filename": "~/minio-data/test/waymo.zip",
        "walletAddress": "wabkhtadjzfydxxda2vzyasg7cimd3jie6ermpw",
        "verifiedDeal": "false",
        "fastRetrieval": "true",
        "dataCid": "bafykbzaceb5cfdrbg45khvhk4mza6",
        "minerId": "t03354",
        "price": "0.000005",
        "duration": "1036700",                    //epochs
        "dealCid": "bafyreicmqtttadqdksrqvunxhcgvfvb47m",
        "timeStamp": "1628025191856290"           //miliseconds
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Online Deals (single file)
GET `minio/retrieve/{bucket}/{object}`

#### Example: 

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN
```

Response from POSTMAN
```bash
{
    "data": {
        "file_name": "waymo.zip",
        "deals": [
            {
                "data": {
                    "filename": "~/minio-data/test/waymo.zip",
                    "walletAddress": "5wabkhtadjzfydxxdq66j4dubbhwpnojqd3jmpw",
                    "verifiedDeal": "false",
                    "fastRetrieval": "true",
                    "dataCid": "bafykbzaceb5cfdpdupjd4mza6",
                    "minerId": "t03354",
                    "price": "0.000005",
                    "duration": "1036700",
                    "dealCid": "bafyreicmm2g654",
                    "timeStamp": "1628025191856290"
                },
                "status": "success",
                "message": "success"
            },
            {
                "data": {
                    "filename": "~/minio-data/testre/waymo.zip",
                    "walletAddress": "5wabkhtadjzfydxxda2vzyasg7cimkcphswrq66j4dubbhwpnoj",
                    "verifiedDeal": "false",
                    "fastRetrieval": "true",
                    "dataCid": "bafykbzaceb5cfdpdupjd4mza6",
                    "minerId": "t03354",
                    "price": "0.000005",
                    "duration": "1036700",
                    "dealCid": "bafyreijg227wlo4bge76bcxk7cw",
                    "timeStamp": "1628026238100552"
                },
                "status": "success",
                "message": "success"
            }
        ]
    },
    "status": "success",
    "message": "success"
}
```

### Send Bucket Online Deals (bucket zip file)
POST `minio/deals/{bucket}`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
    "VerifiedDeal":"false",
    "FastRetrieval":"true",
    "MinerId":"t00000",
    "Price": "0.000005",
    "Duration":"1036800"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "filename": "~/minio-data/test_deals.zip",
        "walletAddress": "h376xbytsd3jie6ermpw",
        "verifiedDeal": "false",
        "fastRetrieval": "true",
        "dataCid": "bafk2bza5dgw6pubjodkscqpg",
        "minerId": "t03354",
        "price": "0.000005",
        "duration": "518800",
        "dealCid": "bafyreicvqh7krdhdnpkqwokze",
        "timeStamp": "1629835134146540"
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Bucket Online Deals (bucket zip file)
GET `minio/bucket/retrieve/{bucket}`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN
```

Response from POSTMAN
```bash
{
    "data": {
        "bucket_name": "20210824",
        "deals": [
            {
                "data": {
                    "filename": "~/minio-data/test_deals.zip",
                    "walletAddress": "t3u7pum2vzyasg7cimkpnojqd3jie6erm",
                    "verifiedDeal": "false",
                    "fastRetrieval": "true",
                    "dataCid": "bafykbvgxdpej7neeoqsnvuzppme",
                    "minerId": "t03354",
                    "price": "0.000005",
                    "duration": "518700",
                    "dealCid": "bafyrekm3lmusljgmvyriqid6kcaoed5kni",
                    "timeStamp": "1629816006709676"
                },
                "status": "success",
                "message": "success"
            },
            {
                "data": {
                    "filename": "~/minio-data/test_deals.zip",
                    "walletAddress": "t3u7khtadjzfydxxdanojqd3jie6ermpw",
                    "verifiedDeal": "false",
                    "fastRetrieval": "true",
                    "dataCid": "bafykbnvz5rgs7obwbfztqrr4ahwjue",
                    "minerId": "t03354",
                    "price": "0.000005",
                    "duration": "518800",
                    "dealCid": "bafyrgigdm4ppqzwt4vufm4m3pmuvolnfe",
                    "timeStamp": "1629833844752891"
                },
                "status": "success",
                "message": "success"
            }
        ]
    },
    "status": "success",
    "message": "success"
}
```

### Send Bucket Offline Deals 
POST `minio/offlinedeals/{bucket}`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
    "Task_Name":"test_name",
    "Curated_Dataset":"test_dataset",
	"Description":"test_description",
	"Is_Public": "1",             // public: "1", private: "0"
	"Type": "regular",            // "verified" if deal is verified else "regular"
	"Miner_Id" : "test_miner",    // miner id is ignored if <Is_Public> is set to "1"    
	"Min_Price" : "0.000005",
	"Max_Price" : "0.00005",
	"Tags" : "test_tag",
	"Expire_Days" : "10"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "bucket_name": "test",
        "deals": {
            "data": {
                "taskname": "test-name",
                "filename": "be450523-52ed-44f9-9828-8e382c0d15c8.csv",
                "uuid": "d2d79d42-6f79-46fe-97bd-cd6f69c25116"
            },
            "status": "success",
            "message": "Task created successfully."
        }
    },
    "status": "success",
    "message": "success"
}
```

### Add Volume Backup Plan
POST `/minio/backup/add/plan`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
   "BackupPlanName":"daily",
   "BackupInterval":"1",      //unit in day
   "Price":"0.0005",          //unit in FIL
   "Duration":"518400",       //unit in epoch
   "VerifiedDeal":false,
   "FastRetrieval":true
}
```
Response from POSTMAN
```bash
{
    "data": {
        "backupPlanId": 2,
        "backupPlanName": "daily",
        "backupInterval": "1",
        "minerRegion": "",
        "price": "0.0005",
        "duration": "518400",
        "verifiedDeal": false,
        "fastRetrieval": true,
        "status": "Running",
        "lastBackupOn":""
        "createdOn": "1637790990711492",
        "updatedOn": "1637790990711492"
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Backup Plans
GET `/minio/backup/retrieve/plan`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN
```
Response from POSTMAN
```bash
{
    "data": {
        "volumeBackupJobPlans": [
            {
                "backupPlanId": 1,
                "backupPlanName": "monthly",
                "backupInterval": "1",
                "minerRegion": "",
                "price": "0.0005",
                "duration": "518400",
                "verifiedDeal": false,
                "fastRetrieval": true,
                "status": "Running",
                "lastBackupOn":""
                "createdOn": "1637790861901038",
                "updatedOn": "1637790861901038"
            },
            {
                "backupPlanId": 2,
                "backupPlanName": "daily",
                "backupInterval": "1",
                "minerRegion": "",
                "price": "0.0005",
                "duration": "518400",
                "verifiedDeal": false,
                "fastRetrieval": true,
                "status": "Running",
                "lastBackupOn":"1637790990711494"
                "createdOn": "1637790990711492",
                "updatedOn": "1637790990711492"
            }
        ],
        "backupPlansCounts": 2
    },
    "status": "success",
    "message": "success"
}
```

### Add Volume Backup Job
POST `/minio/backup/add/job`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
   "BackupPlanId":1
}
```
Response from POSTMAN
```bash
{
    "data": {
        "data": null,         
        "backupTaskId": 2,
        "status": "Created"
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Backup Jobs
GET `/minio/backup/retrieve/volume`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN
```
Response from POSTMAN
```bash
{
    "data": {
        "volumeBackupPlans": [
            {
                "backupPlanName": "monthly",
                "backupPlanId": 1,
                "backupPlanTasks": [
                    {
                        "data": [
                            {
                                "uuid": "56f-63-42-b3-b48",
                                "source_file_name": "minio-data",
                                "miner_id": "t00000",
                                "deal_cid": "bafy",
                                "payload_cid": "QmU759Bk5ZT",
                                "file_source_url": "https://ipfs.io/ipfs/QmXo3ZKSnR",
                                "md5": "",
                                "start_epoch": 1328131,
                                "piece_cid": "baga",
                                "file_size": 2050927,
                                "cost": ""
                            }
                        ],
                        "backupTaskId": 1,
                        "status": "Created"
                    }
                ],
                "backupPlanTasksCounts": 1
            }
        ],
        "backupTasksCounts": 1,
        "backupPlansCounts": 1,
        "completedVolumeBackupTasksCounts": 0,
        "inProcessVolumeBackupTasksCounts": 1,
        "failedVolumeBackupTasksCounts": 0
    },
    "status": "success",
    "message": "success"
}
```

### Backup Volume
POST `/minio/backup/volume`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{
   "BackupTaskId": 1,
   "BackupPlanId": 1,
   "BackupPlanName":"test"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "data": [
            {
                "uuid": "56-6e-43-b4-b4",
                "source_file_name": "minio-data",
                "miner_id": "",
                "deal_cid": "",
                "payload_cid": "QmU75",
                "file_source_url": "https://ipfs.io/ipfs/QmXo3",
                "md5": "",
                "start_epoch": 1328131,
                "piece_cid": "baga",
                "file_size": 2050927,
                "cost": ""
            }
        ],
        "backupTaskId": 3,
        "status": "Created"
    },
    "status": "success",
    "message": "success"
}
```

### Add Volume Rebuild Job
POST `/minio/rebuild/add/job`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{ 
    "BackupTaskId": 1,
    "BackupPlanId": 1
}
```
Response from POSTMAN
```bash
{
    "data": {
        "rebuildTaskID": 4,
        "createdOn": "1637801932084525",
        "updatedOn": "1637801932084525",
        "miner_id": "t0000",
        "deal_cid": "bafy",
        "payload_cid": "Qmag",
        "backupTaskId": 1,
        "status": "Created"
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Rebuild Jobs
GET `minio/rebuild/retrieve/volume`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN
```
Response from POSTMAN
```bash
{
    "data": {
        "volumeRebuildTasks": [
            {
                "rebuildTaskID": 1,
                "createdOn": "1637801932084525",
                "updatedOn": "1637801932084525",
                "miner_id": "t00000",
                "deal_cid": "bafy",
                "payload_cid": "Qmag",
                "backupTaskId": 1,
                "status": "Created"
            }
        ],
        "volumeRebuildTasksCounts": 1,
        "completedVolumeRebuildTasksCounts": 0,
        "inProcessVolumeRebuildTasksCounts": 1,
        "failedVolumeRebuildTasksCounts": 0
    },
    "status": "success",
    "message": "success"
}
```

### Rebuild Volume
POST `/minio/rebuild/volume`

#### Example:

Send request using POSTMAN

``` bash
# Headers
## Use a new User-Agent instead of the default User-Agent in Postman
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36

#Authorization
Bearer Token = MY_FS3_TOKEN

# Body
{ 
    "VolumeRebuildTaskId":1,
    "MinerId": "t024557",
    "PayloadCid": "QmXbeefGNYavf6R3WBpSNtadN6m2mtLaAhKJJFtE9kfkHn",
    "DealCid": "bafyreifo2pp5d4se44xu32p5ikm3qjzmfv7ihbmdsilz7j5wii7h7ne3gm"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "volume_rebuild_address": "/home/test/minio-data",
        "volume_rebuild_name": "minio-data",
        "miner_id": "t00000",
        "deal_cid": "bafy",
        "payload_cid": "QmXb",
        "timeStamp": "1637802544000230"
    },
    "status": "success",
    "message": "success"
}
```

# Deployment Recommendations

## Allow port access for Firewalls

By default FS3 uses the port 9000 to listen for incoming connections. If your platform blocks the port by default, you may need to enable access to the port.

### ufw

For hosts with ufw enabled (Debian based distros), you can use `ufw` command to allow traffic to specific ports. Use below command to allow access to port 9000

```sh
ufw allow 9000
```

Below command enables all incoming traffic to ports ranging from 9000 to 9010.

```sh
ufw allow 9000:9010/tcp
```

### firewall-cmd

For hosts with firewall-cmd enabled (CentOS), you can use `firewall-cmd` command to allow traffic to specific ports. Use below commands to allow access to port 9000

```sh
firewall-cmd --get-active-zones
```

This command gets the active zone(s). Now, apply port rules to the relevant zones returned above. For example if the zone is `public`, use

```sh
firewall-cmd --zone=public --add-port=9000/tcp --permanent
```

Note that `permanent` makes sure the rules are persistent across firewall start, restart or reload. Finally reload the firewall for changes to take effect.

```sh
firewall-cmd --reload
```

### iptables

For hosts with iptables enabled (RHEL, CentOS, etc), you can use `iptables` command to enable all traffic coming to specific ports. Use below command to allow
access to port 9000

```sh
iptables -A INPUT -p tcp --dport 9000 -j ACCEPT
service iptables restart
```

Below command enables all incoming traffic to ports ranging from 9000 to 9010.

```sh
iptables -A INPUT -p tcp --dport 9000:9010 -j ACCEPT
service iptables restart
```

## Pre-existing data
When deployed on a single drive, FS3 lets clients access any pre-existing data in the data directory. For example, if FS3 is started with the command  `minio server /mnt/data`, any pre-existing data in the `/mnt/data` directory would be accessible to the clients.

The above statement is also valid for all gateway backends.

# Test FS3 Connectivity

## Test using FS3 Browser
FS3 Server comes with an embedded web based object browser. Point your web browser to http://127.0.0.1:9000 to ensure your server has started successfully.

## License

[AGPL](https://github.com/filswan/fs3/blob/master/LICENSE)
