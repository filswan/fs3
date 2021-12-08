package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/codingsince1985/checksum"
	clientmodel "github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/client/lotus"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	libutils "github.com/filswan/go-swan-lib/utils"
	files "github.com/ipfs/go-ipfs-files"
	ipfsClient "github.com/ipfs/go-ipfs-http-client"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	oshomedir "github.com/mitchellh/go-homedir"
	"github.com/robfig/cron"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	ioioutil "io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	MicroSecondPerDay                 = 86400000000
	MicroSecondPerMinute              = 60000000
	FS3SourceId                       = 3
	TableVolumeBackupDealsMetadataCsv = "volume_backup_deals_metadata_csv"
	TableVolumeBackupDealsCarCsv      = "volume_backup_deals_car_csv"
	StatusBackupTaskCreated           = "Created"
	StatusRebuildTaskCompleted        = "Completed"
	StatusBackupPlanEnabled           = "Enabled"
	LOTUS_JSON_RPC_ID                 = 7878
	LOTUS_JSON_RPC_VERSION            = "2.0"
	LOTUS_CLIENT_IMPORT_CAR           = "Filecoin.ClientImport"
	LOTUS_CLIENT_Retrieve_DEAL        = "Filecoin.ClientRetrieve"
)

func BackupScheduler() {
	c := cron.New()
	//backup scheduler

	interval := "@every 2m"
	err := c.AddFunc(interval, func() {
		logs.GetLogger().Println("++++++++++ backup volume scheduler is running at " + time.Now().Format("2006-01-02 15:04:05") + " ++++++++++")
		err := BackupVolumeScheduler()
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

func BackupVolumeScheduler() error {
	//get all the running backup plans from db
	runningBackupPlans, err := GetRunningBackupPlans()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if runningBackupPlans == nil {
		return err
	} else {
		if len(runningBackupPlans) == 0 {
			logs.GetLogger().Info("No backup plan running now.")
			return err
		}
	}
	//get executable backup plans Id in this scheduler turn
	ExecuteBackupPlansId := []int{}
	for _, v := range runningBackupPlans {
		timestamp := time.Now().UTC().UnixNano() / 1000
		var LastBackupOn int
		if v.LastBackupOn != "" {
			LastBackupOn, _ = strconv.Atoi(v.LastBackupOn)
		} else {
			LastBackupOn, _ = strconv.Atoi(v.CreatedOn)
		}
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		backupInterval, err := strconv.Atoi(v.Interval)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		if timestamp >= int64(LastBackupOn)+int64(MicroSecondPerDay)*int64(backupInterval) {
			ExecuteBackupPlansId = append(ExecuteBackupPlansId, v.ID)
		}
	}

	if len(ExecuteBackupPlansId) == 0 {
		logs.GetLogger().Info("No backup plan needs to backup in this turn.")
		return err
	}
	//updata backup plans LastBackupTime
	err = UpdateLastBackupTime(ExecuteBackupPlansId)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	// add backup jobs
	volumeBackupRequests, err := AddBackupVolumeJobs(ExecuteBackupPlansId)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	//backup volume
	err = BackupVolumeJobs(volumeBackupRequests)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func BackupVolumeJobs(volumeBackupRequests []VolumeBackupRequest) error {
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

	for _, v := range volumeBackupRequests {
		backupPlanName, backupPlanId, backupTaskId := v.BackupPlanName, v.BackupPlanId, v.BackupTaskId
		// get volume path
		volumePath, err := VolumePath()
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		// create backup folder if not exist
		volumeBackupFolderPath, err := VolumeBackUpPath()
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		if _, err := os.Stat(volumeBackupFolderPath); os.IsNotExist(err) {
			err := os.Mkdir(volumeBackupFolderPath, 0775)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
		}
		// generate car file using ipfs
		// generate datacid for volume folder
		ipfsApiAddress := config.GetUserConfig().IpfsApiAddress
		hash, err := IpfsAddFolder(volumePath, ipfsApiAddress)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("FS3 volume backup payload cid: ", hash)

		// generate car file for the volume folder
		confCar := &clientmodel.ConfCar{
			LotusClientApiUrl:      config.GetUserConfig().LotusClientApiUrl,
			LotusClientAccessToken: config.GetUserConfig().LotusClientAccessToken,
			OutputDir:              volumeBackupFolderPath,
			InputDir:               volumePath,
		}

		volumeCarPath, err := generateCarFileWithIpfs(ipfsApiAddress, hash, volumeBackupFolderPath)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("FS3 volume backup car file generation succeed")

		// lotus import car file
		err = LotusRpcClientImportCar(volumeCarPath)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		//generate car.csv
		carCsvStructList, err := generateCarInfo(hash, volumeCarPath, confCar)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("car files created in ", confCar.OutputDir)

		for _, v := range carCsvStructList {
			fileDesc := PsqlVolumeBackupCarCsv{
				Uuid:           v.Uuid,
				SourceFileName: v.SourceFileName,
				SourceFilePath: v.SourceFilePath,
				SourceFileMd5:  v.SourceFileMd5,
				SourceFileSize: v.SourceFileSize,
				CarFileName:    v.CarFileName,
				CarFilePath:    v.CarFilePath,
				CarFileMd5:     v.CarFileMd5,
				CarFileUrl:     v.CarFileUrl,
				CarFileSize:    v.CarFileSize,
				DealCid:        v.DealCid,
				DataCid:        v.DataCid,
				PieceCid:       v.PieceCid,
				MinerFid:       v.MinerFid,
				Cost:           v.Cost,
			}
			result := db.Create(&fileDesc)
			if result.Error != nil {
				logs.GetLogger().Error(result.Error)
				return result.Error
			}
		}

		//upload to ipfs
		confUpload := &clientmodel.ConfUpload{
			StorageServerType:           libconstants.STORAGE_SERVER_TYPE_IPFS_SERVER,
			IpfsServerDownloadUrlPrefix: config.GetUserConfig().IpfsGateway,
			IpfsServerUploadUrl:         config.GetUserConfig().IpfsApiAddress,
			OutputDir:                   confCar.OutputDir,
			InputDir:                    confCar.OutputDir,
		}
		uploadedCarCsvStructList, err := subcommand.UploadCarFiles(confUpload)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("car files uploaded")

		var uploadedFileDesc PsqlVolumeBackupCarCsv
		if err := db.Last(&uploadedFileDesc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.GetLogger().Info("No record found in database")
				return nil
			} else {
				logs.GetLogger().Error(err)
				return err
			}
		}

		for _, v := range uploadedCarCsvStructList {
			uploadedFileDesc.Uuid = v.Uuid
			uploadedFileDesc.SourceFileName = v.SourceFileName
			uploadedFileDesc.SourceFilePath = v.SourceFilePath
			uploadedFileDesc.SourceFileMd5 = v.SourceFileMd5
			uploadedFileDesc.SourceFileSize = v.SourceFileSize
			uploadedFileDesc.CarFileName = v.CarFileName
			uploadedFileDesc.CarFilePath = v.CarFilePath
			uploadedFileDesc.CarFileMd5 = v.CarFileMd5
			uploadedFileDesc.CarFileUrl = v.CarFileUrl
			uploadedFileDesc.CarFileSize = v.CarFileSize
			uploadedFileDesc.DealCid = v.DealCid
			uploadedFileDesc.DataCid = v.DataCid
			uploadedFileDesc.PieceCid = v.PieceCid
			uploadedFileDesc.MinerFid = v.MinerFid
			uploadedFileDesc.Cost = v.Cost
			if err := db.Save(&uploadedFileDesc).Error; err != nil {
				logs.GetLogger().Error(err)
				continue
			}
		}

		backupPlanInfo, err := GetBackupPlanInfo(db, backupPlanId)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		// create public task on swan
		startEpochIntervalHours := 96
		startEpoch := libutils.GetCurrentEpoch() + (startEpochIntervalHours+1)*libconstants.EPOCH_PER_HOUR
		maxPrice, err := decimal.NewFromString(backupPlanInfo.Price)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		duration, err := strconv.Atoi(backupPlanInfo.Duration)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		confTask := &clientmodel.ConfTask{
			SwanApiUrl:                 config.GetUserConfig().SwanAddress,
			SwanToken:                  config.GetUserConfig().SwanToken,
			PublicDeal:                 true,
			BidMode:                    libconstants.TASK_BID_MODE_AUTO,
			VerifiedDeal:               backupPlanInfo.VerifiedDeal,
			OfflineMode:                false,
			FastRetrieval:              backupPlanInfo.FastRetrieval,
			MaxPrice:                   maxPrice,
			StorageServerType:          libconstants.STORAGE_SERVER_TYPE_IPFS_SERVER,
			WebServerDownloadUrlPrefix: confUpload.IpfsServerDownloadUrlPrefix,
			ExpireDays:                 4,
			Duration:                   duration,
			OutputDir:                  confCar.OutputDir,
			InputDir:                   confCar.OutputDir,
			TaskName:                   backupPlanName,
			StartEpochIntervalHours:    startEpochIntervalHours,
			StartEpoch:                 startEpoch,
			SourceId:                   FS3SourceId,
		}

		_, metadataCsvStructList, taskCsvStructList, err := subcommand.CreateTask(confTask, nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		// save metadata
		for _, v := range metadataCsvStructList {
			metadata := PsqlVolumeBackupMetadataCsv{
				Uuid:           v.Uuid,
				SourceFileName: v.SourceFileName,
				SourceFilePath: v.SourceFilePath,
				SourceFileMd5:  v.SourceFileMd5,
				SourceFileSize: v.SourceFileSize,
				CarFileName:    v.CarFileName,
				CarFilePath:    v.CarFilePath,
				CarFileMd5:     v.CarFileMd5,
				CarFileUrl:     v.CarFileUrl,
				CarFileSize:    v.CarFileSize,
				DealCid:        v.DealCid,
				DataCid:        v.DataCid,
				PieceCid:       v.PieceCid,
				MinerFid:       v.MinerFid,
				StartEpoch:     *v.StartEpoch,
				SourceId:       *v.SourceId,
				Cost:           v.Cost,
			}
			result := db.Create(&metadata)
			if result.Error != nil {
				logs.GetLogger().Error(result.Error)
				return result.Error
			}
		}
		//save backup task to db
		err = SaveBackupTaskToDb(taskCsvStructList, backupPlanId, backupTaskId, db)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("task created")
	}
	return nil
}

func AddBackupVolumeJobs(plansId []int) ([]VolumeBackupRequest, error) {
	//open backup db
	db, err := GetPsqlDb()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	//close db
	sqlDB, err := db.DB()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	defer sqlDB.Close()

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	var volumeBackupRequests []VolumeBackupRequest
	for _, v := range plansId {
		var backupPlan PsqlVolumeBackupPlan
		if err := db.First(&backupPlan, v).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.GetLogger().Info("No record found in database")
				continue
			} else {
				logs.GetLogger().Error(err)
				continue
			}
		}
		backupJob := PsqlVolumeBackupJob{
			Name:               backupPlan.Name,
			VolumeBackupPlanID: backupPlan.ID,
			Duration:           backupPlan.Duration,
			CreatedOn:          timestamp,
			UpdatedOn:          timestamp,
			Status:             StatusBackupTaskCreated,
		}
		if err := db.Create(&backupJob).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.GetLogger().Info("No record found in database")
				continue
			} else {
				logs.GetLogger().Error(err)
				continue
			}
		}
		var job PsqlVolumeBackupJob
		if err := db.Last(&job).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logs.GetLogger().Info("No record found in database")
				continue
			} else {
				logs.GetLogger().Error(err)
				continue
			}
		}
		volumeBackupRequest := VolumeBackupRequest{
			BackupTaskId:   job.ID,
			BackupPlanId:   backupPlan.ID,
			BackupPlanName: backupPlan.Name,
		}
		volumeBackupRequests = append(volumeBackupRequests, volumeBackupRequest)
	}
	return volumeBackupRequests, err
}

func UpdateLastBackupTime(plansId []int) error {
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

	for _, v := range plansId {
		var plan PsqlVolumeBackupPlan
		db.Where("ID = ?", v).First(&plan)
		timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
		plan.LastBackupOn = timestamp
		db.Save(plan)
	}
	return err
}

func GetRunningBackupPlans() ([]PsqlVolumeBackupPlan, error) {
	//get backupplans
	//open backup db
	db, err := GetPsqlDb()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	//close db
	sqlDB, err := db.DB()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	defer sqlDB.Close()

	var plans []PsqlVolumeBackupPlan
	result := db.Where("status = ?", StatusBackupPlanEnabled).Find(&plans)

	return plans, result.Error
}

type VolumeBackupJobPlans struct {
	VolumeBackupJobPlans       []VolumeBackupJobPlan `json:"volumeBackupJobPlans"`
	VolumeBackupJobPlansCounts int                   `json:"backupPlansCounts"`
}

type VolumeBackupJobPlan struct {
	BackupPlanId   int    `json:"backupPlanId"`
	BackupPlanName string `json:"backupPlanName"`
	BackupInterval string `json:"backupInterval"`
	MinerRegion    string `json:"minerRegion"`
	Price          string `json:"price"`
	Duration       string `json:"duration"`
	VerifiedDeal   bool   `json:"verifiedDeal"`
	FastRetrieval  bool   `json:"fastRetrieval"`
	Status         string `json:"status"`
	LastBackupOn   string `json:"lastBackupOn"`
	CreatedOn      string `json:"createdOn"`
	UpdatedOn      string `json:"updatedOn"`
}

type VolumeBackupRequest struct {
	BackupTaskId   int    `json:"backupTaskId"`
	BackupPlanId   int    `json:"backupPlanId"`
	BackupPlanName string `json:"backupPlanName"`
}

func VolumePath() (string, error) {
	fs3VolumeAddress := config.GetUserConfig().Fs3VolumeAddress
	expandedFs3VolumeAddress, err := oshomedir.Expand(fs3VolumeAddress)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedFs3VolumeAddress, nil
}

func VolumeBackUpPath() (string, error) {
	volumeBackUpAddress := config.GetUserConfig().VolumeBackupAddress
	expandedVolumeBackUpAddresss, err := oshomedir.Expand(volumeBackUpAddress)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedVolumeBackUpAddresss, nil
}

func IpfsAddFolder(volumePath string, ipfsApiUrl string) (string, error) {
	ipfsApi := NewApi()
	api, err := ipfsClient.NewURLApiWithClient(ipfsApiUrl, ipfsApi)
	c(err)
	stat, err := os.Stat(volumePath)
	c(err)
	// This walks the filesystem at /tmp/example/ and create a list of the files / directories we have.
	node, err := files.NewSerialFile(volumePath, true, stat)
	c(err)
	// Add the files / directory to IPFS
	path, err := api.Unixfs().Add(context.Background(), node)
	c(err)
	// Output the resulting CID
	return fmt.Sprint(path.Root().String()), nil
}

func NewApi() *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	}
	return c
}

func c(err error) {
	if err != nil {
		logs.GetLogger().Error(err)
	}
}

func generateCarFileWithIpfs(ipfsApiAddress string, hash string, volumeBackupPath string) (string, error) {
	logs.GetLogger().Info("volume backup car file generation begins")
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
	volumeCarName := "volume_" + timestamp + ".car"
	volumeCarPath := filepath.Join(volumeBackupPath, volumeCarName)

	commandLine := "curl -X POST \"" + ipfsApiAddress + "/api/v0/dag/export?arg=" + hash + "&progress=true\" >" + volumeCarPath
	_, err := ExecCommand(commandLine)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	if _, err := os.Stat(volumeCarPath); errors.Is(err, os.ErrNotExist) {
		logs.GetLogger().Error("volume backup car file generation failed")
	}
	logs.GetLogger().Info("volume backup car file generation success. Car file path:  ", volumeCarPath)
	return volumeCarPath, nil
}

func ExecCommand(strCommand string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", strCommand)
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		logs.GetLogger().Error("Execute failed when Start:" + err.Error())
		return "", err
	}
	out_bytes, _ := ioioutil.ReadAll(stdout)
	if err := stdout.Close(); err != nil {
		logs.GetLogger().Error("Execute failed when close stdout:" + err.Error())
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		logs.GetLogger().Error("Execute failed when Wait:" + err.Error())
		return "", err
	}
	return string(out_bytes), nil
}

func generateCarInfo(hash string, volumeCarPath string, confCar *clientmodel.ConfCar) ([]*libmodel.FileDesc, error) {
	carFiles := []*libmodel.FileDesc{}
	lotusClient, err := lotus.LotusGetClient(confCar.LotusClientApiUrl, confCar.LotusClientAccessToken)

	carFile := libmodel.FileDesc{}
	carFile.SourceFileName = filepath.Base(confCar.InputDir)
	carFile.SourceFilePath = confCar.InputDir
	carFile.SourceFileSize, _ = DirSize(confCar.InputDir)
	carFile.CarFileName = filepath.Base(volumeCarPath)
	carFile.CarFilePath = filepath.Join(confCar.OutputDir, carFile.CarFileName)
	fmt.Println()

	pieceCid := lotusClient.LotusClientCalcCommP(carFile.CarFilePath)
	if pieceCid == nil {
		err := fmt.Errorf("failed to generate piece cid")
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFile.PieceCid = *pieceCid
	carFile.DataCid = hash
	carFile.CarFileSize = libutils.GetFileSize(carFile.CarFilePath)

	if confCar.GenerateMd5 {
		srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.SourceFileMd5 = srcFileMd5

		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.CarFileMd5 = carFileMd5
	}

	carFiles = append(carFiles, &carFile)

	_, err = subcommand.WriteCarFilesToFiles(carFiles, confCar.OutputDir, libconstants.JSON_FILE_NAME_BY_CAR, libconstants.CSV_FILE_NAME_BY_CAR, subcommand.SUBCOMMAND_CAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(carFiles), " car files info has been created to directory:", confCar.OutputDir)

	return carFiles, nil
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func SaveBackupTaskToDb(task []*subcommand.Deal, backupPlanId int, backupTaskId int, db *gorm.DB) error {
	tasks := []subcommand.Deal{}
	for _, v := range task {
		tasks = append(tasks, *v)
	}
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	for _, v := range tasks {
		task := PsqlVolumeBackupTaskCsv{
			Uuid:           v.Uuid,
			SourceFileName: v.SourceFileName,
			MinerId:        v.MinerId,
			DealCid:        v.DealCid,
			PayloadCid:     v.PayloadCid,
			FileSourceUrl:  v.FileSourceUrl,
			Md5:            v.Md5,
			StartEpoch:     *v.StartEpoch,
			PieceCid:       v.PieceCid,
			FileSize:       v.FileSize,
			Cost:           v.Cost,
		}
		result := db.Create(&task)
		if result.Error != nil {
			logs.GetLogger().Error(result.Error)
			return result.Error
		}

		var job PsqlVolumeBackupJob
		db.Where("id=?", backupTaskId).First(&job)
		job.Status = StatusBackupTaskRunning
		job.UpdatedOn = timestamp
		job.Uuid = task.Uuid
		job.SourceFileName = task.SourceFileName
		job.MinerId = task.MinerId
		job.DealCid = task.DealCid
		job.PayloadCid = task.PayloadCid
		job.FileSourceUrl = task.FileSourceUrl
		job.Md5 = task.Md5
		job.StartEpoch = task.StartEpoch
		job.PieceCid = task.PieceCid
		job.FileSize = task.FileSize
		job.Cost = task.Cost
		db.Save(&job)
	}
	return nil
}

func GetBackupPlanInfo(db *gorm.DB, backupPlanId int) (PsqlVolumeBackupPlan, error) {
	var plan PsqlVolumeBackupPlan
	db.Where("id = ?", backupPlanId).First(&plan)
	return plan, nil
}

func LotusRpcClientImportCar(carPath string) error {
	clientImportCar := ClientImportCar{
		Path:  carPath,
		IsCAR: true,
	}
	var params []interface{}
	params = append(params, clientImportCar)
	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_IMPORT_CAR,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}
	client.HttpGet(config.GetUserConfig().LotusClientApiUrl, config.GetUserConfig().LotusClientAccessToken, jsonRpcParams)
	return nil
}

type ClientImportCar struct {
	Path  string
	IsCAR bool
}

func GetPsqlDb() (*gorm.DB, error) {
	host := config.GetUserConfig().PsqlHost
	user := config.GetUserConfig().PsqlUser
	password := config.GetUserConfig().PsqlPassword
	dbname := config.GetUserConfig().PsqlDbname
	port := config.GetUserConfig().PsqlPort
	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return db, err
}

type PsqlVolumeBackupPlan struct {
	ID            int `gorm:"primary_key"`
	Name          string
	Interval      string
	MinerRegion   string
	Price         string
	Duration      string
	VerifiedDeal  bool
	FastRetrieval bool
	Status        string
	LastBackupOn  string
	CreatedOn     string
	UpdatedOn     string
}

type PsqlVolumeBackupJob struct {
	ID                 int `gorm:"primary_key"`
	Name               string
	Uuid               string
	SourceFileName     string
	MinerId            string
	DealCid            string
	PayloadCid         string
	FileSourceUrl      string
	Md5                string
	StartEpoch         int
	PieceCid           string
	FileSize           int64
	Cost               string
	Duration           string
	Status             string
	CreatedOn          string
	UpdatedOn          string
	VolumeBackupPlanID int
	VolumeBackupPlan   PsqlVolumeBackupPlan `gorm:"foreignKey:VolumeBackupPlanID"`
}

type PsqlVolumeBackupCarCsv struct {
	gorm.Model
	Uuid           string
	SourceFileName string
	SourceFilePath string
	SourceFileMd5  string
	SourceFileSize int64
	CarFileName    string
	CarFilePath    string
	CarFileMd5     string
	CarFileUrl     string
	CarFileSize    int64
	DealCid        string
	DataCid        string
	PieceCid       string
	MinerFid       string
	StartEpoch     int
	SourceId       int `gorm:"SMALLINT"`
	Cost           string
}

type PsqlVolumeBackupMetadataCsv struct {
	gorm.Model
	Uuid           string
	SourceFileName string
	SourceFilePath string
	SourceFileMd5  string
	SourceFileSize int64
	CarFileName    string
	CarFilePath    string
	CarFileMd5     string
	CarFileUrl     string
	CarFileSize    int64
	DealCid        string
	DataCid        string
	PieceCid       string
	MinerFid       string
	StartEpoch     int
	SourceId       int `gorm:"SMALLINT"`
	Cost           string
}

type PsqlVolumeBackupTaskCsv struct {
	gorm.Model
	Uuid           string
	SourceFileName string
	MinerId        string
	DealCid        string
	PayloadCid     string
	FileSourceUrl  string
	Md5            string
	StartEpoch     int
	PieceCid       string
	FileSize       int64
	Cost           string
}
