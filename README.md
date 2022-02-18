# FS3 Quickstart Guide
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)

- Join us on our public [Slack channel](https://filswan.slack.com) for news, discussions, and status updates. 
- Check out our [medium](https://filswan.medium.com) for the latest posts and announcements.

## How to use
### Prerequisite
- Golang 1.15+.
- Node Js 14.0+.
- PostgreSQL 10.19+.
- IPFS 0.8.0+.
- Lotus node 1.13+.

__Note__: A Lotus full node is not a must for FS3 if a lite node is already configured to have connection with a Lotus full node. More information on how to config a lite node can be found at [Lotus Lite node](https://lotus.filecoin.io/docs/set-up/lotus-lite/).

## Functions
* Upload files to FS3 as a local container for storage service.
* Backup a single file or an entire bucket to FIL by using online deals service.
* Backup the whole volume with customized schedulers(daily/weekly) using offline deals. 
* Send volume backup task to the assigned storage provider automatically using Autobid module.
* Rebuild the volume content by retrieving previous volume backup content specified by users.
* Save all the deals information into PostgreSQL database.
* List all the deals history and status.

__Note__: Please check [FS3 User Guide](https://docs.filswan.com/fs3/fs3-user-guide) on how to use FS3 service.

## Install Prerequisite Dependencies
### IPFS node
A running IPFS node is needed for CAR file generation and files uploading and storage. You can refer [IPFS Command-line Docs](https://docs.ipfs.io/install/command-line/#official-distributions) for installation instructions and configuration.
### Lotus node
A running lotus node is needed for CAR file information generation, deals sending and deals retrieval. You can refer [Lotus Docs](https://lotus.filecoin.io/docs/set-up/install/) for installation instructions and configuration. As mentioned before, a Lotus full node is not a must for FS3 server but if you do not have a lotus full node, a lite node which configured to connect to a full node is required. More information on how to use a [Lotus Lite node](https://lotus.filecoin.io/docs/set-up/lotus-lite/).
### PostgreSQL
A PostgreSQL database is required to be pre-built for FS3 server usage. Check [PostgreSQL Tutorial](https://www.postgresqltutorial.com/) on installation and connection instructions. The required database schema and tables schema are listed below.

#### Install PostgreSQL on Ubuntu
First, execute the following command to create the file repository configuration:
```sh
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
```
Second, import the repository signing key:
```sh
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
```
Third, update the package list:
```sh
sudo apt-get update
```
Finally, install the latest version of PostgreSQL:
```sh
sudo apt-get install postgresql
```
If you want to install a specific version, you can use `postgresql-version` instead of `postgresql`. For example, to install PostgreSQL version 12, you use the following command:
```sh
sudo apt-get install postgresql-12
```

#### Connect to the PostgreSQL database server via psql
In PostgreSQL, a user account is referred as a role. When you installed PostgreSQL, the installation process created a user account called `postgres` associated with the default `postgres` role. To connect to PostgreSQL using the `postgres` role, you switch over to the `postgres` account on your server by typing:
```sh
sudo -i -u postgres
```
It’ll prompt for the password of the current user. You need to provide the password and hit the `Enter/Return` keyboard.
Then, you can access the PostgreSQL using the psql by typing the following command:
```sh
psql
```
You’ll access the postgres prompt like this:
```sh
postgres=#
```
To quit the PostgreSQL prompt, you run the following command:
```sh
postgres=# \q
```
This above command will bring you back to the postgres Linux command prompt.
```sh
postgres@ubuntu-dev:~$
```
To return to your regular system user, you execute the `exit` command like this:
```sh
postgres@ubuntu-dev:~$ exit
```

#### Create PostgreSQL Database
```sh
sudo -u postgres psql
postgres=# create database fs3;                                                //create 'fs3' PostgreSQL database                                    
postgres=# create user root with encrypted password 'root';                    //create USER and PASSWORD as 'root'
postgres=# grant all privileges on database fs3 to root;                       //grant privileges
postgres=# \q                                                                  //logout postgres
```

# Install from Source
## 1. Checkout source code
```
git clone https://github.com/filswan/fs3
cd fs3
git checkout <release_branch>
```

## 2. Build the Source Code
#### 2.1 Build UI
```bash
cd browser

npm install

# install cross-env if not exists
$ npm i cross-env --save

# Build production projects
$ npm run release:prod
```

#### 2.2 Install Filecoin dependency
```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```
#### 2.3 Install go module dependency
``` bash 
cd ..
git submodule update --init --recursive
make ffi
```

#### 2.4 Build up Tables in Postgresql Database
Create tables and sequences
```sh
bash db_setup.sh             
```
It’ll prompt for the password of the current user. You need to provide the password and hit the `Enter/Return` keyboard.Then it'll prompt for the password of the database user. Type `root` and hit the `Enter/Return` keyboard.

#### 2.5 Set up FS3 configuration
Set up and customize FS3 configuration by making modifications on `.env` file, which stores your information as environment variables. An example config is given as `.env.example` for reference. 
``` bash
vim .env
```

Modify the `.env` file based on your use cases:

* __SWAN_ADDRESS__ : The address of filswan platform, default as `https://api.filswan.com`.
* __FS3_VOLUME_ADDRESS__ : The address of FS3 VOLUME where the uploaded data stored, default as `~/fs3-data`. If changed, the FS3 server start command has to be changed accordingly. For example, if the `FS3_VOLUME_ADDRESS` is changed to `/PATH/TO/MY_FS3_VOLUME_ADDRESS`, the server running command line will change from `./minio server ~/fs3-data` to `./minio server /PATH/TO/MY_FS3_VOLUME_ADDRESS`.
* __FS3_WALLET_ADDRESS__ : A wallet address is a must for sending deals to miner. Firstly,follow the instruction on [Manage FIL](https://lotus.filecoin.io/docs/set-up/manage-fil/#creating-a-wallet) to create a wallet. Secondly, after creating a wallet, check [obtaining FIL](https://lotus.filecoin.io/docs/set-up/manage-fil/#obtaining-fil) on how to receive FIL.
* __CAR_FILE_SIZE__ : A fixed car file size in bytes need to be predefined before generating car files for trunk via variable `CarFileSize`, such as `8589934592` for 8Gb as default.
* __IPFS_API_ADDRESS__ :  An available IPFS address with port need to be set up as the format of `https://MyIPFSUrl:Port`. Firstly, find the configuration file of IPFS node, which is located at `~/.ipfs/config` by default. Secondly, look up `API` in `Addresses` section of the `config` file. Thirdly, transform the `API` url into form of `https://MyIPFSUrl:Port`. For example, transform `/ip4/192.168.88.41/tcp/5001` into `http://127.0.0.1:5001`. Check [IPFS Docs](https://docs.ipfs.io/how-to/configure-node/#profiles) for more information of IPFS node configuration.
* __IPFS_GATEWAY__ :  An available IPFS address with port need to be set up for file downloading as the format of `https://MyIPFSGatewayUrl:Port`. Firstly, find the configuration file of IPFS node, which is located at `~/.ipfs/config` by default. Secondly, look up `Gateway` in `Addresses` section of the `config` file. Thirdly, transform the `Gateway` url into form of `https://MyIPFSUrl:Port`. For example, transform `/ip4/192.168.88.41/tcp/5001` into `http://127.0.0.1:5001`. The process is similar to that of `IPFS_API_ADDRESS`. Check [IPFS Docs](https://docs.ipfs.io/how-to/configure-node/#profiles) for more information of IPFS node configuration.
* __SWAN_TOKEN__ : A valid swan token is required for posting task on swan platform. It can be received after creating an account on [Filswan](https://console.filswan.com/). Check [Filswan APIs](https://documenter.getpostman.com/view/13140808/TWDZJbzV) for more details on how to get an authorization token.
* __LOTUS_CLIENT_API_URL__ : A valid lotus endpoint is required to connect to a Lotus node as the format of `http://[api:port]/rpc/v0`.The Lotus node comes with its own local API endpoint. Firstly, find the lotus configuration file which is located at `~/.lotus/config.toml` by default. Secondly, look up `ListenAddress` in the `[API]` section of the `config.toml` file. Thirdly,transform the lotus api address found into format as `http://[api:port]/rpc/v0`. For example, transform the endpoint `/ip4/0.0.0.0/tcp/1234/http` found in config file into `http://127.0.0.1:1234/rpc/v0`. Check  [Lotus Configuration](https://lotus.filecoin.io/docs/set-up/configuration/) for more information of Lotus configuration.
* __LOTUS_CLIENT_ACCESS_TOKEN__ :An `admin` permission token is required to talk to the lotus API endpoints. It can be generated by following the [Obtaining tokens](https://lotus.filecoin.io/docs/developers/api-access/#obtaining-tokens) steps. Check [Lotus Docs](https://lotus.filecoin.io/) for more details.
* __PSQL_HOST__ : The host name of the machine on which the server is running. Default as:`127.0.0.1`
* __PSQL_USER__ : The user name to connect to the database(You must have the permission to do so,of course). Default as `root`.
* __PSQL_PASSWORD__ : The password for connecting to a database if password authentification is required. Default as `root`.
* __PSQL_DBNAME__ : The name of the database you want to connect to. Default as `fs3`.
* __PSQL_PORT__ : The database server port to which you want to connect. Default as `5432`.
  
__Note__:If the configuration is changed in the future, build up the FS3 server again to make the changes take effect.

#### Examples:
*  __FS3_WALLET_ADDRESS__
    1. Creata a wallet
          ```shell
          # Create a BLS wallet
          lotus wallet new bls
          ```
          Or
          ```shell
          # Create a secp256k1 wallet
          lotus wallet new
          ```
   
*  __IPFS_API_ADDRESS__ and __IPFS_GATEWAY__:
    1. Open the IPFS configuration file:
          ```shell
          vim ~/.ipfs/config               //default configuration location
          ```
    2. IPFS Configuration file example (truncated):
          ```shell
          {
              "API": {
                "HTTPHeaders": {
                  "Access-Control-Allow-Methods": [
                    "PUT",
                    "GET",
                    "POST"
                  ],
                  "Access-Control-Allow-Origin": [
                    "http://localhost:3000",
                    "http://127.0.0.1:5001",
                    "https://webui.ipfs.io"
                  ]
                }
              },
              "Addresses": {
                "API": "/ip4/127.0.0.1/tcp/5001",             // IPFS API ADDRESS
                "Announce": [],
                "Gateway": "/ip4/127.0.0.1/tcp/5050",         // IPFS GATEWAY
                ...       
              }
              ...
          }
          ```
    3. Transform format and fill in the `.env` file:
          ```shell
          IPFS_API_URL = http://127.0.0.1:5001               //transformed from "API": "/ip4/127.0.0.1/tcp/5001"
          IPFS_GATEWAY = http://127.0.0.1:8080               //transformed from "Gateway": "/ip4/127.0.0.1/tcp/5050"
          ```
* __SWAN_TOKEN__:
    1. Create an account on [Filswan](https://console.filswan.com/) with your Email and Password.
    2. Send request to [Filswan APIs](https://documenter.getpostman.com/view/13140808/TWDZJbzV):
       
       Example Request:
          ```shell
          curl --location --request POST 'https://api.filswan.com//auth/login' \
          --data-raw '{
          "email":"my-email",
          "password":"my-password"
          }'
          ```
       Example Response
          ```shell
          {
          "auth_token": "eyJXXX.eyJXXX9.nqVRXXX",
          "message": "Successfully logged in.",
          "status": "success"
          }
          ```
    3. Fill in the `.env` file:
          ```shell
          SWAN_TOKEN = eyJXXX.eyJXXX9.nqVRXXX
          ```
       
* __LOTUS_CLIENT_API_URL__:
    1. Open the Lotus configuration file:
          ```shell
          vim ~/.lotus/config.toml               //default configuration location
          ```
    2. Lotus Configuration file example (truncated):
          ```shell
          # Default config:
          [API]
          ListenAddress = /ip4/127.0.0.1/tcp/1234/http  // LOTUS CLIENT API URL
          RemoteListenAddress = ""
          #  Timeout = "30s"
          ...
          ```
    3. Transform format and fill in the `.env` file:
          ```shell
          IPFS_API_URL = http://127.0.0.1/rpc/v0        //transformed from "ListenAddress": /ip4/127.0.0.1/tcp/1234/http
          ```

* __LOTUS_CLIENT_ACCESS_TOKEN__:
    1. Create lotus node access token with `admin` permission:
          ```shell
          lotus auth create-token --perm admin             //admin permission 
          ```
    2. `admin` token example :
          ```shell
          eyJXXXX.eyJXXXX._J7XXXX
          ```
    3. Fill in the `.env` file:
          ```shell
          LOTUS_CLIENT_ACCESS_TOKEN = eyJXXXX.eyJXXXX._J7XXXX
          ```


       
#### 2.6 Build up FS3 server
``` bash
make
```

## Run a Standalone FS3 Server
``` bash
./minio server ~/fs3-data
```

The default FS3 volume address `Fs3VolumeAddress` is set as `~/fs3-data`, which can be changed in `.env`. If the volume address is changed in the future, build up the FS3 server again to make the changes take effect.

#### Access Key and Secret Key
The FS3 deployment starts using default root credentials `minioadmin:minioadmin` but you can change it with your own credentials.

#### Change your Access Key and Secret Key
``` bash
export MINIO_ROOT_USER= MY_FS3_ACCESS_KEY
export MINIO_ROOT_PASSWORD=MY_FS3_SECRET_KEY
```

If you change the credential, build up FS3 server again to make it take effect. Then re-run the fs3 server.
``` bash
make

./minio server ~/fs3-data
```

## Open FS3 Browser
You can test the deployment using the FS3 Browser, an embedded
web-based object browser built into FS3 Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, send deals, retrieve data and browse the contents on the FS3 server.

You can also connect using any S3-compatible tool, such as the [FS3-mc](https://github.com/filswan/fs3-mc) commandline tool.

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
        "filename": "~/fs3-data/test/waymo.zip",
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
                    "filename": "~/fs3-data/test/waymo.zip",
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
                    "filename": "~/fs3-data/testre/waymo.zip",
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
        "filename": "~/fs3-data/test_deals.zip",
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
                    "filename": "~/fs3-data/test_deals.zip",
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
                    "filename": "~/fs3-data/test_deals.zip",
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
        "ID": 1,
        "Name": "daily",
        "Interval": "1",
        "MinerRegion": "",
        "Price": "0.0005",
        "Duration": "518400",
        "VerifiedDeal": false,
        "FastRetrieval": true,
        "Status": "Enabled",            // plan is set to "Enabled" as default when created frist time
        "LastBackupOn": "",
        "CreatedOn": "1638396058992883",
        "UpdatedOn": "1638396058992883"
    },
    "status": "success",
    "message": "success"
}
```

### Update Volume Backup Plan
POST `/minio/backup/update/plan`

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
   "BackupPlanId":"2",
   "Status":"Disabled"
}
```
Response from POSTMAN
```bash
{
    "data": {
        "ID": 2,
        "Name": "weekly",
        "Interval": "1",
        "MinerRegion": "",
        "Price": "0.0005",
        "Duration": "518400",
        "VerifiedDeal": false,
        "FastRetrieval": true,
        "Status": "Disabled",
        "LastBackupOn": "",
        "CreatedOn": "1638396058992883",
        "UpdatedOn": "1638396058992883"
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Backup Plans
POST `/minio/backup/retrieve/plan`

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
   "Offset":0,   //default as 0 
   "Limit":10"   //default as 10
   "Status": ["Enabled","Disabled"]  //default as all status
}
```
Response from POSTMAN
```bash
{
    "data": {
        "backupPlans": [
            {
                "ID": 1,
                "Name": "daily",
                "Interval": "1",
                "MinerRegion": "",
                "Price": "0.0005",
                "Duration": "518400",
                "VerifiedDeal": false,
                "FastRetrieval": true,
                "Status": "Enabled",
                "LastBackupOn": "1638397353014173",
                "CreatedOn": "1638396058992883",
                "UpdatedOn": "1638396058992883"
            }
        ],
        "TotalVolumeBackupPlan": 1
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Backup Jobs
POST `/minio/backup/retrieve/volume`

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
   "Offset":0,   //default as 0 
   "Limit":10"   //default as 10
}
```
Response from POSTMAN
```bash
{
    "data": {
        "VolumeBackupJobs ": [
            {
                "ID": 1,
                "Name": "daily",
                "Uuid": "90-da-45-a0-14",
                "SourceFileName": "fs3-data",
                "MinerId": "t000000",
                "DealCid": "bafy",
                "PayloadCid": "QmXv",
                "FileSourceUrl": "https://ipfs.io/ipfs/QmXh",
                "Md5": "",
                "StartEpoch": 1347978,
                "PieceCid": "baga",
                "FileSize": 849400,
                "Cost": "254331901032",
                "Duration": "518400",
                "Status": "Running",
                "CreatedOn": "1638396635020859",
                "UpdatedOn": "1638396721214519",
                "VolumeBackupPlanID": 1,
                "VolumeBackupPlan": {
                    "ID": 0,
                    "Name": "",
                    "Interval": "",
                    "MinerRegion": "",
                    "Price": "",
                    "Duration": "",
                    "VerifiedDeal": false,
                    "FastRetrieval": false,
                    "Status": "",
                    "LastBackupOn": "",
                    "CreatedOn": "",
                    "UpdatedOn": ""
                }
            },
        ],
        "totalVolumeBackupTasksCounts": 1,
        "completedVolumeBackupTasksCounts": 0,
        "inProcessVolumeBackupTasksCounts": 1,
        "failedVolumeBackupTasksCounts": 0
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
    "BackupTaskId": 1
}
```
Response from POSTMAN
```bash
{
    "data": {
        "ID": 1,
        "MinerId": "t000000",
        "DealCid": "bafy",
        "PayloadCid": "QmXv",
        "Status": "Created",
        "CreatedOn": "1638398434961766",
        "UpdatedOn": "1638398434961766",
        "BackupJobId": 1,
        "BackupJob": {
            "ID": 0,
            "Name": "",
            "Uuid": "",
            "SourceFileName": "",
            "MinerId": "",
            "DealCid": "",
            "PayloadCid": "",
            "FileSourceUrl": "",
            "Md5": "",
            "StartEpoch": 0,
            "PieceCid": "",
            "FileSize": 0,
            "Cost": "",
            "Duration": "",
            "Status": "",
            "CreatedOn": "",
            "UpdatedOn": "",
            "VolumeBackupPlanID": 0,
            "VolumeBackupPlan": {
                "ID": 0,
                "Name": "",
                "Interval": "",
                "MinerRegion": "",
                "Price": "",
                "Duration": "",
                "VerifiedDeal": false,
                "FastRetrieval": false,
                "Status": "",
                "LastBackupOn": "",
                "CreatedOn": "",
                "UpdatedOn": ""
            }
        }
    },
    "status": "success",
    "message": "success"
}
```

### Retrieve Volume Rebuild Jobs
POST `minio/rebuild/retrieve/volume`

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
   "Offset":0,   //default as 0 
   "Limit":10"   //default as 10
}
```
Response from POSTMAN
```bash
{
    "data": {
        "volumeRebuildJobs": [
            {
                "ID": 1,
                "MinerId": "t000000",
                "DealCid": "bafy",
                "PayloadCid": "QmXv",
                "Status": "Created",
                "CreatedOn": "1638398434961766",
                "UpdatedOn": "1638398434961766",
                "BackupPlanName": "daily",
                "BackupJobId": 1,
                "BackupJob": {
                    "ID": 0,
                    "Name": "",
                    "Uuid": "",
                    "SourceFileName": "",
                    "MinerId": "",
                    "DealCid": "",
                    "PayloadCid": "",
                    "FileSourceUrl": "",
                    "Md5": "",
                    "StartEpoch": 0,
                    "PieceCid": "",
                    "FileSize": 0,
                    "Cost": "",
                    "Duration": "",
                    "Status": "",
                    "CreatedOn": "",
                    "UpdatedOn": "",
                    "VolumeBackupPlanID": 0,
                    "VolumeBackupPlan": {
                        "ID": 0,
                        "Name": "",
                        "Interval": "",
                        "MinerRegion": "",
                        "Price": "",
                        "Duration": "",
                        "VerifiedDeal": false,
                        "FastRetrieval": false,
                        "Status": "",
                        "LastBackupOn": "",
                        "CreatedOn": "",
                        "UpdatedOn": ""
                    }
                }
            }
        ],
        "totalVolumeRebuildTasksCounts": 1,
        "completedVolumeRebuildTasksCounts": 0,
        "inProcessVolumeRebuildTasksCounts": 1,
        "failedVolumeRebuildTasksCounts": 0
    },
    "status": "success",
    "message": "success"
}
```

#### Database Schema
```sh
                          List of relations
 Schema |                  Name                   |   Type   | Owner 
--------+-----------------------------------------+----------+-------
 public | psql_volume_backup_car_csvs             | table    | root
 public | psql_volume_backup_car_csvs_id_seq      | sequence | root
 public | psql_volume_backup_jobs                 | table    | root
 public | psql_volume_backup_jobs_id_seq          | sequence | root
 public | psql_volume_backup_metadata_csvs        | table    | root
 public | psql_volume_backup_metadata_csvs_id_seq | sequence | root
 public | psql_volume_backup_plans                | table    | root
 public | psql_volume_backup_plans_id_seq         | sequence | root
 public | psql_volume_backup_task_csvs            | table    | root
 public | psql_volume_backup_task_csvs_id_seq     | sequence | root
 public | psql_volume_rebuild_jobs                | table    | root
 public | psql_volume_rebuild_jobs_id_seq         | sequence | root
(12 rows)
```

### Tables schema
##### Table psql_volume_backup_car_csvs
```sh
                                          Table "public.psql_volume_backup_car_csvs"
      Column      |           Type           | Collation | Nullable |                         Default                         
------------------+--------------------------+-----------+----------+---------------------------------------------------------
 id               | bigint                   |           | not null | nextval('psql_volume_backup_car_csvs_id_seq'::regclass)
 created_at       | timestamp with time zone |           |          | 
 updated_at       | timestamp with time zone |           |          | 
 deleted_at       | timestamp with time zone |           |          | 
 uuid             | text                     |           |          | 
 source_file_name | text                     |           |          | 
 source_file_path | text                     |           |          | 
 source_file_md5  | text                     |           |          | 
 source_file_size | bigint                   |           |          | 
 car_file_name    | text                     |           |          | 
 car_file_path    | text                     |           |          | 
 car_file_md5     | text                     |           |          | 
 car_file_url     | text                     |           |          | 
 car_file_size    | bigint                   |           |          | 
 deal_cid         | text                     |           |          | 
 data_cid         | text                     |           |          | 
 piece_cid        | text                     |           |          | 
 miner_fid        | text                     |           |          | 
 start_epoch      | bigint                   |           |          | 
 source_id        | bigint                   |           |          | 
 cost             | text                     |           |          | 
Indexes:
    "psql_volume_backup_car_csvs_pkey" PRIMARY KEY, btree (id)
    "idx_psql_volume_backup_car_csvs_deleted_at" btree (deleted_at)
```

##### Table psql_volume_backup_jobs
```sh
                                   Table "public.psql_volume_backup_jobs"
        Column         |  Type  | Collation | Nullable |                       Default                       
-----------------------+--------+-----------+----------+-----------------------------------------------------
 id                    | bigint |           | not null | nextval('psql_volume_backup_jobs_id_seq'::regclass)
 name                  | text   |           |          | 
 uuid                  | text   |           |          | 
 source_file_name      | text   |           |          | 
 miner_id              | text   |           |          | 
 deal_cid              | text   |           |          | 
 payload_cid           | text   |           |          | 
 file_source_url       | text   |           |          | 
 md5                   | text   |           |          | 
 start_epoch           | bigint |           |          | 
 piece_cid             | text   |           |          | 
 file_size             | bigint |           |          | 
 cost                  | text   |           |          | 
 duration              | text   |           |          | 
 status                | text   |           |          | 
 created_on            | text   |           |          | 
 updated_on            | text   |           |          | 
 volume_backup_plan_id | bigint |           |          | 
Indexes:
    "psql_volume_backup_jobs_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "fk_psql_volume_backup_jobs_volume_backup_plan" FOREIGN KEY (volume_backup_plan_id) REFERENCES psql_volume_backup_plans(id)
Referenced by:
    TABLE "psql_volume_rebuild_jobs" CONSTRAINT "fk_psql_volume_rebuild_jobs_backup_job" FOREIGN KEY (backup_job_id) REFERENCES psql_volume_backup_jobs(id)
```

##### Table psql_volume_backup_metadata_csvs
```sh
                                          Table "public.psql_volume_backup_metadata_csvs"
      Column      |           Type           | Collation | Nullable |                           Default                            
------------------+--------------------------+-----------+----------+--------------------------------------------------------------
 id               | bigint                   |           | not null | nextval('psql_volume_backup_metadata_csvs_id_seq'::regclass)
 created_at       | timestamp with time zone |           |          | 
 updated_at       | timestamp with time zone |           |          | 
 deleted_at       | timestamp with time zone |           |          | 
 uuid             | text                     |           |          | 
 source_file_name | text                     |           |          | 
 source_file_path | text                     |           |          | 
 source_file_md5  | text                     |           |          | 
 source_file_size | bigint                   |           |          | 
 car_file_name    | text                     |           |          | 
 car_file_path    | text                     |           |          | 
 car_file_md5     | text                     |           |          | 
 car_file_url     | text                     |           |          | 
 car_file_size    | bigint                   |           |          | 
 deal_cid         | text                     |           |          | 
 data_cid         | text                     |           |          | 
 piece_cid        | text                     |           |          | 
 miner_fid        | text                     |           |          | 
 start_epoch      | bigint                   |           |          | 
 source_id        | bigint                   |           |          | 
 cost             | text                     |           |          | 
Indexes:
    "psql_volume_backup_metadata_csvs_pkey" PRIMARY KEY, btree (id)
    "idx_psql_volume_backup_metadata_csvs_deleted_at" btree (deleted_at)
```

##### Table psql_volume_backup_plans
```sh
                                Table "public.psql_volume_backup_plans"
     Column     |  Type   | Collation | Nullable |                       Default                        
----------------+---------+-----------+----------+------------------------------------------------------
 id             | bigint  |           | not null | nextval('psql_volume_backup_plans_id_seq'::regclass)
 name           | text    |           |          | 
 interval       | text    |           |          | 
 miner_region   | text    |           |          | 
 price          | text    |           |          | 
 duration       | text    |           |          | 
 verified_deal  | boolean |           |          | 
 fast_retrieval | boolean |           |          | 
 status         | text    |           |          | 
 last_backup_on | text    |           |          | 
 created_on     | text    |           |          | 
 updated_on     | text    |           |          | 
Indexes:
    "psql_volume_backup_plans_pkey" PRIMARY KEY, btree (id)
Referenced by:
    TABLE "psql_volume_backup_jobs" CONSTRAINT "fk_psql_volume_backup_jobs_volume_backup_plan" FOREIGN KEY (volume_backup_plan_id) REFERENCES psql_volume_backup_plans(id)
```
##### Table psql_volume_backup_task_csvs
```sh
                                          Table "public.psql_volume_backup_task_csvs"
      Column      |           Type           | Collation | Nullable |                         Default                          
------------------+--------------------------+-----------+----------+----------------------------------------------------------
 id               | bigint                   |           | not null | nextval('psql_volume_backup_task_csvs_id_seq'::regclass)
 created_at       | timestamp with time zone |           |          | 
 updated_at       | timestamp with time zone |           |          | 
 deleted_at       | timestamp with time zone |           |          | 
 uuid             | text                     |           |          | 
 source_file_name | text                     |           |          | 
 miner_id         | text                     |           |          | 
 deal_cid         | text                     |           |          | 
 payload_cid      | text                     |           |          | 
 file_source_url  | text                     |           |          | 
 md5              | text                     |           |          | 
 start_epoch      | bigint                   |           |          | 
 piece_cid        | text                     |           |          | 
 file_size        | bigint                   |           |          | 
 cost             | text                     |           |          | 
Indexes:
    "psql_volume_backup_task_csvs_pkey" PRIMARY KEY, btree (id)
    "idx_psql_volume_backup_task_csvs_deleted_at" btree (deleted_at)
```
##### Table psql_volume_rebuild_jobs
```sh
                               Table "public.psql_volume_rebuild_jobs"
    Column     |  Type  | Collation | Nullable |                       Default                        
---------------+--------+-----------+----------+------------------------------------------------------
 id            | bigint |           | not null | nextval('psql_volume_rebuild_jobs_id_seq'::regclass)
 miner_id      | text   |           |          | 
 deal_cid      | text   |           |          | 
 payload_cid   | text   |           |          | 
 status        | text   |           |          | 
 created_on    | text   |           |          | 
 updated_on    | text   |           |          | 
 backup_job_id | bigint |           |          | 
Indexes:
    "psql_volume_rebuild_jobs_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "fk_psql_volume_rebuild_jobs_backup_job" FOREIGN KEY (backup_job_id) REFERENCES psql_volume_backup_jobs(id)
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
