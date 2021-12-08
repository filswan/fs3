package scheduler

import (
	"errors"
	clientmodel "github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/client/lotus"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	oshomedir "github.com/mitchellh/go-homedir"
	"github.com/robfig/cron"
	"gorm.io/gorm"
	"path/filepath"
	"strconv"
	"time"
)

const (
	TableVolumeBackupTask     = "volume_backup_task"
	TableVolumeBackupPlan     = "volume_backup_plan"
	TableVolumeRebuildTask    = "volume_rebuild_task"
	StatusBackupPlanRunning   = "Running"
	StatusBackupTaskRunning   = "Running"
	StatusStorageDealActive   = "StorageDealActive"
	StatusBackupTaskCompleted = "Completed"
	StatusRebuildTaskCreated  = "Created"
	StatusRebuildTaskRunning  = "Running"
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
		logs.GetLogger().Println("^^^^^^^^^^ send deal scheduler is running at " + time.Now().Format("2006-01-02 15:04:05") + " ^^^^^^^^^^")
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
	db, err := GetPsqlDb()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	//close db
	sqlDB, err := db.DB()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer sqlDB.Close()

	var backupJobs []PsqlVolumeBackupJob
	if err := db.Find(&backupJobs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.GetLogger().Info("No record found in database")
			return nil
		} else {
			logs.GetLogger().Error(err)
			return err
		}
	}

	for _, values := range backupJobs {
		if values.Status == StatusBackupTaskRunning {
			if values.DealCid == "" {
				logs.GetLogger().Info("Backup job missing dealcid, ID: ", values.ID)
				continue
			}
			status, err := CheckDealStatus(values.DealCid)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			if status == StatusStorageDealActive {
				values.Status = StatusBackupTaskCompleted
				timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
				values.UpdatedOn = timestamp
				logs.GetLogger().Info("Backup job done, ID: ", values.ID, ", UUID: ", values.Uuid)
				if err := db.Save(&values).Error; err != nil {
					logs.GetLogger().Error(err)
				}
			}
		}
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
	db, err := GetPsqlDb()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	//close db
	sqlDB, err := db.DB()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer sqlDB.Close()

	//update backuptasks
	for _, v := range tasks {
		var backupJob PsqlVolumeBackupJob
		if err := db.Where("uuid=?", v[0].Uuid).First(&backupJob).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.GetLogger().Info("No record found in database")
			} else {
				logs.GetLogger().Error(err)
				continue
			}
		}
		timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
		backupJob.Status = StatusBackupTaskRunning
		backupJob.UpdatedOn = timestamp
		backupJob.MinerId = v[0].MinerFid
		backupJob.DealCid = v[0].DealCid
		backupJob.Cost = v[0].Cost
		if err := db.Save(backupJob).Error; err != nil {
			logs.GetLogger().Error(err)
			continue
		}
		logs.GetLogger().Info("Backup job sent to miner, ID: ", backupJob.ID, ", UUID: ", v[0].Uuid)
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
