package scheduler

import (
	"context"
	"encoding/json"
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
	"github.com/syndtr/goleveldb/leveldb"
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
	LOTUS_JSON_RPC_ID                 = 7878
	LOTUS_JSON_RPC_VERSION            = "2.0"
	LOTUS_CLIENT_IMPORT_CAR           = "Filecoin.ClientImport"
	LOTUS_CLIENT_Retrieve_DEAL        = "Filecoin.ClientRetrieve"
)

func BackupScheduler() {
	c := cron.New()
	//backup scheduler

	interval := "@every 10m"
	fmt.Println(interval)
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

	//get all the running backup plans from db
	runningBackupPlans, err := GetRunningBackupPlans(db)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if runningBackupPlans == nil {
		return err
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
		backupInterval, err := strconv.Atoi(v.BackupInterval)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		if timestamp > int64(LastBackupOn)+3*int64(MicroSecondPerMinute)*int64(backupInterval) {
			ExecuteBackupPlansId = append(ExecuteBackupPlansId, v.BackupPlanId)
		}
	}

	//updata backup plans LastBackupTime
	err = UpdateLastBackupTime(db, ExecuteBackupPlansId)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	// add backup jobs
	volumeBackupRequests, err := AddBackupVolumeJobs(db, ExecuteBackupPlansId)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if volumeBackupRequests == nil {
		return err
	}

	//backup volume
	err = BackupVolumeJobs(db, volumeBackupRequests)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func BackupVolumeJobs(db *leveldb.DB, volumeBackupRequests []VolumeBackupRequest) error {
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

		dataBytes, err := json.Marshal(&carCsvStructList)
		err = db.Put([]byte(TableVolumeBackupDealsCarCsv), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
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

		dataBytes, err = json.Marshal(&uploadedCarCsvStructList)
		err = db.Put([]byte(TableVolumeBackupDealsCarCsv), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
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
		dataBytes, err = json.Marshal(&metadataCsvStructList)
		err = db.Put([]byte(TableVolumeBackupDealsMetadataCsv), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		//save backup task to db
		_, err = SaveBackupTaskToDb(taskCsvStructList, backupPlanId, backupTaskId, db)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info("task created")
	}
	return nil
}

func AddBackupVolumeJobs(db *leveldb.DB, plansId []int) ([]VolumeBackupRequest, error) {

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	backupPlansKey := TableVolumeBackupPlan

	//check if key exists
	has, err := db.Has([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if has == false {
		return nil, err
	}
	backupPlans, err := db.Get([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	data := VolumeBackupJobPlans{}
	err = json.Unmarshal(backupPlans, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	volumeBackupRequests := []VolumeBackupRequest{}
	for _, v := range plansId {
		backupPlan := VolumeBackupJobPlan{}
		for j, value := range data.VolumeBackupJobPlans {
			if value.BackupPlanId == v {
				backupPlan = data.VolumeBackupJobPlans[j]
				break
			}
		}

		dbVolumeBackupTasks := TableVolumeBackupTask
		//check if key exists
		has, err := db.Has([]byte(dbVolumeBackupTasks), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		if has != false {
			volumeBackupTasks, err := db.Get([]byte(dbVolumeBackupTasks), nil)
			taskData := VolumeBackupTasks{}
			err = json.Unmarshal(volumeBackupTasks, &taskData)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			newVolumeBackupPlanTask := VolumeBackupPlanTask{
				CreatedOn:    timestamp,
				UpdatedOn:    timestamp,
				BackupTaskId: taskData.VolumeBackupTasksCounts + 1,
				Status:       StatusBackupTaskCreated,
			}
			newVolumeBackupPlanTask.Data.Duration = backupPlan.Duration
			planIndex := -1
			for i, v := range taskData.VolumeBackupPlans {
				if v.BackupPlanId == backupPlan.BackupPlanId {
					planIndex = i
					taskData.InProcessVolumeBackupTasksCounts = taskData.InProcessVolumeBackupTasksCounts + 1
					taskData.VolumeBackupTasksCounts = taskData.VolumeBackupTasksCounts + 1
					taskData.VolumeBackupPlans[planIndex].BackupPlanTasksCounts = taskData.VolumeBackupPlans[planIndex].BackupPlanTasksCounts + 1
					taskData.VolumeBackupPlans[planIndex].BackupPlanTasks = append(taskData.VolumeBackupPlans[planIndex].BackupPlanTasks, newVolumeBackupPlanTask)
					dataByte, err := json.Marshal(taskData)
					if err != nil {
						logs.GetLogger().Error(err)
						return nil, err
					}
					err = db.Put([]byte(dbVolumeBackupTasks), []byte(dataByte), nil)
					if err != nil {
						logs.GetLogger().Error(err)
						return nil, err
					}
					volumeBackupRequest := VolumeBackupRequest{
						BackupTaskId:   newVolumeBackupPlanTask.BackupTaskId,
						BackupPlanId:   backupPlan.BackupPlanId,
						BackupPlanName: backupPlan.BackupPlanName,
					}
					volumeBackupRequests = append(volumeBackupRequests, volumeBackupRequest)
				}
			}
			if planIndex == -1 {
				taskData.VolumeBackupPlansCounts = taskData.VolumeBackupPlansCounts + 1
				taskData.VolumeBackupTasksCounts = taskData.VolumeBackupTasksCounts + 1
				taskData.InProcessVolumeBackupTasksCounts = taskData.InProcessVolumeBackupTasksCounts + 1
				newVolumeBackupPlanTask := VolumeBackupPlanTask{
					CreatedOn:    timestamp,
					UpdatedOn:    timestamp,
					BackupTaskId: taskData.VolumeBackupTasksCounts,
					Status:       StatusBackupTaskCreated,
				}
				newVolumeBackupPlanTask.Data.Duration = backupPlan.Duration
				newVolumeBackupPlanTasks := []VolumeBackupPlanTask{}
				newVolumeBackupPlanTasks = append(newVolumeBackupPlanTasks, newVolumeBackupPlanTask)
				newVolumeBackupPlan := VolumeBackupPlan{
					BackupPlanName:        backupPlan.BackupPlanName,
					BackupPlanId:          backupPlan.BackupPlanId,
					BackupPlanTasks:       newVolumeBackupPlanTasks,
					BackupPlanTasksCounts: 1,
				}
				taskData.VolumeBackupPlans = append(taskData.VolumeBackupPlans, newVolumeBackupPlan)
				dataByte, err := json.Marshal(taskData)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, err
				}
				err = db.Put([]byte(dbVolumeBackupTasks), []byte(dataByte), nil)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, err
				}
				volumeBackupRequest := VolumeBackupRequest{
					BackupTaskId:   newVolumeBackupPlanTask.BackupTaskId,
					BackupPlanId:   backupPlan.BackupPlanId,
					BackupPlanName: backupPlan.BackupPlanName,
				}
				volumeBackupRequests = append(volumeBackupRequests, volumeBackupRequest)
			}
		} else {
			newVolumeBackupPlanTask := VolumeBackupPlanTask{
				CreatedOn:    timestamp,
				UpdatedOn:    timestamp,
				BackupTaskId: 1,
				Status:       StatusBackupTaskCreated,
			}
			newVolumeBackupPlanTask.Data.Duration = backupPlan.Duration
			newVolumeBackupPlanTasks := []VolumeBackupPlanTask{}
			newVolumeBackupPlanTasks = append(newVolumeBackupPlanTasks, newVolumeBackupPlanTask)
			newVolumeBackupPlan := VolumeBackupPlan{
				BackupPlanName:        backupPlan.BackupPlanName,
				BackupPlanId:          backupPlan.BackupPlanId,
				BackupPlanTasks:       newVolumeBackupPlanTasks,
				BackupPlanTasksCounts: 1,
			}
			newVolumeBackupPlans := []VolumeBackupPlan{}
			newVolumeBackupPlans = append(newVolumeBackupPlans, newVolumeBackupPlan)
			newVolumeBackupTasks := VolumeBackupTasks{
				VolumeBackupPlans:                newVolumeBackupPlans,
				VolumeBackupTasksCounts:          1,
				VolumeBackupPlansCounts:          1,
				CompletedVolumeBackupTasksCounts: 0,
				InProcessVolumeBackupTasksCounts: 1,
				FailedVolumeBackupTasksCounts:    0,
			}

			dataByte, err := json.Marshal(newVolumeBackupTasks)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			err = db.Put([]byte(dbVolumeBackupTasks), []byte(dataByte), nil)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			volumeBackupRequest := VolumeBackupRequest{
				BackupTaskId:   newVolumeBackupPlanTask.BackupTaskId,
				BackupPlanId:   backupPlan.BackupPlanId,
				BackupPlanName: backupPlan.BackupPlanName,
			}
			volumeBackupRequests = append(volumeBackupRequests, volumeBackupRequest)
		}
	}
	return volumeBackupRequests, err
}

func UpdateLastBackupTime(db *leveldb.DB, plansId []int) error {
	//get backupplans
	backupPlansKey := TableVolumeBackupPlan
	//check if key exists
	has, err := db.Has([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if has == false {
		return nil
	}
	backupPlans, err := db.Get([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	data := VolumeBackupJobPlans{}
	err = json.Unmarshal(backupPlans, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, v := range plansId {
		for j, value := range data.VolumeBackupJobPlans {
			if value.BackupPlanId == v {
				timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
				data.VolumeBackupJobPlans[j].LastBackupOn = timestamp
				break
			}
		}
	}
	dataBytes, err := json.Marshal(data)
	err = db.Put([]byte(backupPlansKey), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func GetRunningBackupPlans(db *leveldb.DB) ([]VolumeBackupJobPlan, error) {
	//get backupplans
	backupPlansKey := TableVolumeBackupPlan
	//check if key exists
	has, err := db.Has([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if has == false {
		return nil, err
	}
	backupPlans, err := db.Get([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	data := VolumeBackupJobPlans{}
	err = json.Unmarshal(backupPlans, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	runningBackupPlans := []VolumeBackupJobPlan{}
	for _, v := range data.VolumeBackupJobPlans {
		if v.Status == StatusBackupPlanRunning {
			runningBackupPlans = append(runningBackupPlans, v)
		}
	}
	return runningBackupPlans, err
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
	fmt.Println(path.Root().String())
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
		fmt.Println(err)
	}
}

func generateCarFileWithIpfs(ipfsApiAddress string, hash string, volumeBackupPath string) (string, error) {
	logs.GetLogger().Info("volume backup car file generation begins")
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
	volumeCarName := "volume_" + timestamp + ".car"
	volumeCarPath := filepath.Join(volumeBackupPath, volumeCarName)

	commandLine := "curl -X POST \"" + ipfsApiAddress + "/api/v0/dag/export?arg=" + hash + "&progress=true\" >" + volumeCarPath
	fmt.Println(commandLine)
	_, err := ExecCommand(commandLine)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	if _, err := os.Stat(volumeCarPath); errors.Is(err, os.ErrNotExist) {
		logs.GetLogger().Error("volume backup car file generation failed")
	}
	logs.GetLogger().Info("volume backup car file generation success. Car file path: %s", volumeCarPath)
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

func SaveBackupTaskToDb(task []*subcommand.Deal, backupPlanId int, backupTaskId int, db *leveldb.DB) (VolumeBackupPlanTask, error) {
	tasks := []subcommand.Deal{}
	for _, v := range task {
		tasks = append(tasks, *v)
	}

	dbVolumeBackupTasks := TableVolumeBackupTask
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
	//check if key exists
	has, err := db.Has([]byte(dbVolumeBackupTasks), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupPlanTask{}, err
	}
	if has == false {
		return VolumeBackupPlanTask{}, err
	}
	volumeBackupTasks, _ := db.Get([]byte(dbVolumeBackupTasks), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupPlanTask{}, err
	}

	data := VolumeBackupTasks{}
	err = json.Unmarshal(volumeBackupTasks, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupPlanTask{}, err
	}
	planIndex, taskIndex := -1, -1
	for i, v := range data.VolumeBackupPlans {
		if v.BackupPlanId == backupPlanId {
			planIndex = i
			for j, values := range v.BackupPlanTasks {
				if values.BackupTaskId == backupTaskId {
					taskIndex = j
					data.VolumeBackupPlans[i].BackupPlanTasks[j].Data.DealInfo = tasks
					data.VolumeBackupPlans[i].BackupPlanTasks[j].Status = StatusBackupTaskRunning
					data.VolumeBackupPlans[i].BackupPlanTasks[j].UpdatedOn = timestamp
					break
				}
			}
			break
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupPlanTask{}, err
	}
	err = db.Put([]byte(dbVolumeBackupTasks), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupPlanTask{}, err
	}
	return data.VolumeBackupPlans[planIndex].BackupPlanTasks[taskIndex], err
}

func GetBackupPlanInfo(db *leveldb.DB, backupPlanId int) (VolumeBackupJobPlan, error) {
	backupPlansKey := TableVolumeBackupPlan
	//check if key exists
	has, err := db.Has([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupJobPlan{}, err
	}
	if has == false {
		return VolumeBackupJobPlan{}, errors.New("Key is not in leveldb")
	}
	backupPlans, err := db.Get([]byte(backupPlansKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupJobPlan{}, err
	}
	data := VolumeBackupJobPlans{}
	err = json.Unmarshal(backupPlans, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return VolumeBackupJobPlan{}, err
	}
	for i, v := range data.VolumeBackupJobPlans {
		if v.BackupPlanId == backupPlanId {
			return data.VolumeBackupJobPlans[i], err
		}
	}
	return VolumeBackupJobPlan{}, err
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
	bodyByte, _ := json.Marshal(jsonRpcParams)
	fmt.Println(string(bodyByte))
	response := client.HttpGet(config.GetUserConfig().LotusClientApiUrl, config.GetUserConfig().LotusClientAccessToken, jsonRpcParams)
	fmt.Println(response)
	return nil
}

type ClientImportCar struct {
	Path  string
	IsCAR bool
}
