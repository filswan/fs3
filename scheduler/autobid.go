package scheduler

import (
	"encoding/json"
	"fmt"
	clientmodel "github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/client/lotus"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	libutils "github.com/filswan/go-swan-lib/utils"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	oshomedir "github.com/mitchellh/go-homedir"
	"github.com/robfig/cron"
	"github.com/syndtr/goleveldb/leveldb"
	"path/filepath"
	"strconv"
	"time"
)

const (
	TableVolumeBackupTask     = "volume_backup_task"
	TableVolumeBackupPlan     = "volume_backup_plan"
	StatusBackupPlanRunning   = "Running"
	StatusBackupTaskRunning   = "Running"
	StatusStorageDealActive   = "StorageDealActive"
	StatusBackupTaskCompleted = "Completed"
)

func SendDealScheduler() {
	volumeBackUpAddress := config.GetUserConfig().VolumeBackupAddress
	expandedVolumeBackUpAddresss, err := oshomedir.Expand(volumeBackUpAddress)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	confDeal := &clientmodel.ConfDeal{
		SwanApiUrl:                   config.GetUserConfig().SwanAddress,
		SwanToken:                    config.GetUserConfig().SwanToken,
		LotusClientApiUrl:            config.GetUserConfig().LotusClientApiUrl,
		LotusClientAccessToken:       config.GetUserConfig().LotusClientAccessToken,
		SenderWallet:                 config.GetUserConfig().Fs3WalletAddress,
		OutputDir:                    expandedVolumeBackUpAddresss,
		RelativeEpochFromMainNetwork: -858481,
	}
	confDeal.DealSourceIds = append(confDeal.DealSourceIds, libconstants.TASK_SOURCE_ID_SWAN_FS3)

	c := cron.New()
	// autobid scheduler
	err = c.AddFunc("0 */3 * * * ?", func() {
		logs.GetLogger().Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ send deal scheduler is running at " + time.Now().Format("2006-01-02 15:04:05"))
		err := SendAutobidDealScheduler(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	})
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	c.Start()
}

func SendAutobidDealScheduler(confDeal *clientmodel.ConfDeal) error {
	startEpoch := libutils.GetCurrentEpoch() + (96+1)*libconstants.EPOCH_PER_HOUR
	fmt.Println(startEpoch)
	csvFilepaths, tasks, err := subcommand.SendAutoBidDeals(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	for _, csvFilepath := range csvFilepaths {
		logs.GetLogger().Info(csvFilepath, " is generated")
	}
	err = UpdateSentBackupTasksInDb(tasks)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = UpdateActiveBackupTasksInDb()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func UpdateActiveBackupTasksInDb() error {
	//open backup db
	expandedDir, err := LevelDbBackupPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
	}
	defer db.Close()

	//get backuptasks
	backupTasksKey := TableVolumeBackupTask
	backupTasks, err := db.Get([]byte(backupTasksKey), nil)
	data := VolumeBackupTasks{}
	err = json.Unmarshal(backupTasks, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	//update backuptasks
	for i, value := range data.VolumeBackupPlans {
		for j, values := range value.BackupPlanTasks {
			if values.Status == StatusBackupTaskRunning {
				status, err := CheckDealStatus(values.Data.DealInfo[0].DealCid)
				if err != nil {
					logs.GetLogger().Error(err)
					return err
				}
				if status == StatusStorageDealActive {
					data.VolumeBackupPlans[i].BackupPlanTasks[j].Data.DealInfo[0].DealCid = StatusBackupTaskCompleted
					timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
					data.VolumeBackupPlans[i].BackupPlanTasks[j].UpdatedOn = timestamp
					data.InProcessVolumeBackupTasksCounts = data.InProcessVolumeBackupTasksCounts - 1
					data.CompletedVolumeBackupTasksCounts = data.CompletedVolumeBackupTasksCounts + 1
				}
			}
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(backupTasksKey), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func CheckDealStatus(dealCid string) (string, error) {
	lotusClient, err := lotus.LotusGetClient(config.GetUserConfig().LotusClientApiUrl, config.GetUserConfig().LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	dealCost, err := lotusClient.LotusClientGetDealInfo(dealCid)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return dealCost.Status, err
}

func UpdateSentBackupTasksInDb(tasks [][]*libmodel.FileDesc) error {
	//open backup db
	expandedDir, err := LevelDbBackupPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
	}
	defer db.Close()

	//get backuptasks
	backupTasksKey := TableVolumeBackupTask
	backupTasks, err := db.Get([]byte(backupTasksKey), nil)
	if err != nil || backupTasks == nil {
		logs.GetLogger().Error(err)
		return err
	}
	data := VolumeBackupTasks{}
	err = json.Unmarshal(backupTasks, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	//update backuptasks
	for _, v := range tasks {
		for j, value := range data.VolumeBackupPlans {
			for k, values := range value.BackupPlanTasks {
				if values.Data.DealInfo[0].Uuid == v[0].Uuid {
					timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Status = StatusBackupTaskRunning
					data.VolumeBackupPlans[j].BackupPlanTasks[k].UpdatedOn = timestamp
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Data.DealInfo[0].MinerId = v[0].MinerFid
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Data.DealInfo[0].DealCid = v[0].DealCid
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Data.DealInfo[0].Cost = v[0].Cost
					break
				}
			}
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(backupTasksKey), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

type ClientStartDealParam struct {
	Data              ClientStartDealParamData
	Wallet            string
	Miner             string
	EpochPrice        string
	MinBlocksDuration int
	DealStartEpoch    int
	FastRetrieval     bool
	VerifiedDeal      bool
}
type ClientStartDealParamData struct {
	TransferType string
	Root         Cid
	PieceCid     Cid
	PieceSize    int
}
type Cid struct {
	Cid string `json:"/"`
}
type ClientStartDeal struct {
	LotusJsonRpcResult
	Result Cid `json:"result"`
}
type LotusJsonRpcResult struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Error   *JsonRpcError `json:"error"`
}
type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type LotusJsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

func LevelDbBackupPath() (string, error) {
	volumeBackUpAddress := config.GetUserConfig().VolumeBackupAddress
	levelDbName := ".leveldb.db"
	levelDbPath := filepath.Join(volumeBackUpAddress, levelDbName)
	expandedDir, err := oshomedir.Expand(levelDbPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedDir, nil
}

type VolumeBackupTasks struct {
	VolumeBackupPlans                []VolumeBackupPlan `json:"volumeBackupPlans"`
	VolumeBackupTasksCounts          int                `json:"backupTasksCounts"`
	VolumeBackupPlansCounts          int                `json:"backupPlansCounts"`
	CompletedVolumeBackupTasksCounts int                `json:"completedVolumeBackupTasksCounts"`
	InProcessVolumeBackupTasksCounts int                `json:"inProcessVolumeBackupTasksCounts"`
	FailedVolumeBackupTasksCounts    int                `json:"failedVolumeBackupTasksCounts"`
}

type VolumeBackupPlan struct {
	BackupPlanName        string                 `json:"backupPlanName"`
	BackupPlanId          int                    `json:"backupPlanId"`
	BackupPlanTasks       []VolumeBackupPlanTask `json:"backupPlanTasks"`
	BackupPlanTasksCounts int                    `json:"backupPlanTasksCounts"`
}

type BackupPlanTaskInfo struct {
	DealInfo []subcommand.Deal `json:"dealInfo"`
	Duration string            `json:"duration"`
}
type VolumeBackupPlanTask struct {
	Data         BackupPlanTaskInfo `json:"data"`
	CreatedOn    string             `json:"createdOn"`
	UpdatedOn    string             `json:"updatedOn"`
	BackupTaskId int                `json:"backupTaskId"`
	Status       string             `json:"status"`
}
