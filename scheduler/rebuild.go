package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filswan/go-swan-lib/client"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	"github.com/robfig/cron"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func RebuildScheduler() {
	c := cron.New()
	//backup scheduler
	interval := "@every 1m"
	fmt.Println(interval)
	err := c.AddFunc(interval, func() {
		logs.GetLogger().Println("---------- rebuild volume scheduler is running at " + time.Now().Format("2006-01-02 15:04:05") + " ----------")
		err := RebuildVolumeScheduler()
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

func RebuildVolumeScheduler() error {
	//open backup db
	expandedDir, err := LevelDbBackupPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	var db *leveldb.DB
	for true {
		dB, err := leveldb.OpenFile(expandedDir, nil)
		if err == errors.New("resource temporarily unavailable") {
			db = dB
			continue
		}
		time.Sleep(30 * time.Second)
	}

	if err != nil {
		logs.GetLogger().Error(err)
	}
	defer db.Close()

	//get all the running rebuild jobs from db
	runningRebuildJobs, err := GetOneRunningRebuildJob(db)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if runningRebuildJobs == nil {
		return err
	}
	if len(runningRebuildJobs) == 0 {
		logs.GetLogger().Info("No running rebuild job")
		return err
	}

	//retrieve
	err = RebuildVolumeAndUpdateDb(runningRebuildJobs[0])
	return err
}

func GetOneRunningRebuildJob(db *leveldb.DB) ([]VolumeRebuildTask, error) {
	//get backupplans
	rebuildJobsKey := TableVolumeRebuildTask
	//check if key exists
	has, err := db.Has([]byte(rebuildJobsKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if has == false {
		return nil, err
	}
	backupPlans, err := db.Get([]byte(rebuildJobsKey), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	data := VolumeRebuildJobs{}
	err = json.Unmarshal(backupPlans, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	//get only one rebuild job with the smallest create time
	runningRebuildJobs := []VolumeRebuildTask{}
	for _, v := range data.VolumeRebuildTasks {
		if v.Status == StatusRebuildJobCreated {
			runningRebuildJobs = append(runningRebuildJobs, v)
			break
		}
	}
	return runningRebuildJobs, err
}

type VolumeRebuildTask struct {
	RebuildTaskID int    `json:"rebuildTaskID"`
	CreatedOn     string `json:"createdOn"`
	UpdatedOn     string `json:"updatedOn"`
	MinerId       string `json:"miner_id"`
	DealCid       string `json:"deal_cid"`
	PayloadCid    string `json:"payload_cid"`
	BackupTaskId  int    `json:"backupTaskId"`
	Status        string `json:"status"`
}

type VolumeRebuildJobs struct {
	VolumeRebuildTasks                []VolumeRebuildTask `json:"volumeRebuildTasks"`
	VolumeRebuildTasksCounts          int                 `json:"volumeRebuildTasksCounts"`
	CompletedVolumeRebuildTasksCounts int                 `json:"completedVolumeRebuildTasksCounts"`
	InProcessVolumeRebuildTasksCounts int                 `json:"inProcessVolumeRebuildTasksCounts"`
	FailedVolumeRebuildTasksCounts    int                 `json:"failedVolumeRebuildTasksCounts"`
}

func RebuildVolumeAndUpdateDb(rebuildJob VolumeRebuildTask) error {
	// get volume path
	volumePath, err := VolumePath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	//rename previous version volume
	if _, err := os.Stat(volumePath); !os.IsNotExist(err) {
		dir, file := filepath.Split(volumePath)
		fileBase, fileExt := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)), filepath.Ext(file)
		_, err = exec.Command("mv", volumePath, dir+fileBase+"_"+timestamp+fileExt).Output()
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	//retrieve deal
	err = LotusRpcClientRetrieve(rebuildJob.MinerId, rebuildJob.PayloadCid, volumePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	logs.GetLogger().Info("Rebuild job done, ID: ", rebuildJob.RebuildTaskID)

	//update db
	//open backup db
	expandedDir, err := LevelDbBackupPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	rebuildTimestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	backupTasksKey := TableVolumeRebuildTask
	backupTasks, err := db.Get([]byte(backupTasksKey), nil)
	data := VolumeRebuildJobs{}
	err = json.Unmarshal(backupTasks, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	for i, v := range data.VolumeRebuildTasks {
		if v.RebuildTaskID == rebuildJob.RebuildTaskID {
			data.VolumeRebuildTasks[i].UpdatedOn = rebuildTimestamp
			data.VolumeRebuildTasks[i].Status = StatusRebuildTaskCompleted
			data.InProcessVolumeRebuildTasksCounts = data.InProcessVolumeRebuildTasksCounts - 1
			data.CompletedVolumeRebuildTasksCounts = data.CompletedVolumeRebuildTasksCounts + 1
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(TableVolumeRebuildTask), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func LotusRpcClientRetrieve(minerId string, payloadCid string, outputPath string) error {
	clientRetrieveDealParamDataPartOne := ClientRetrieveDealParamDataPartOne{
		Root: Cid{
			Cid: payloadCid,
		},
		Size:        42,
		Total:       "0",
		UnsealPrice: "0",
		Client:      minerId,
		Miner:       minerId,
	}
	clientRetrieveDealParamDataPartTwo := ClientRetrieveDealParamDataPartTwo{
		Path:  outputPath,
		IsCAR: false,
	}
	var params []interface{}
	params = append(params, clientRetrieveDealParamDataPartOne, clientRetrieveDealParamDataPartTwo)
	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_Retrieve_DEAL,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}
	bodyByte, _ := json.Marshal(jsonRpcParams)
	fmt.Println(string(bodyByte))
	response := client.HttpGet(config.GetUserConfig().LotusClientApiUrl, config.GetUserConfig().LotusClientAccessToken, jsonRpcParams)
	if response == "" {
		err := fmt.Errorf("failed to retrieve data %s from miner %s, no response", payloadCid, minerId)
		logs.GetLogger().Error(err)
		return err
	}
	fmt.Println(response)
	lotusJsonRpcResult := &LotusJsonRpcResult{}
	err := json.Unmarshal([]byte(response), lotusJsonRpcResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	if lotusJsonRpcResult.Error != nil {
		err := fmt.Errorf("failed to retrieve data %s from miner %s, message: %s", payloadCid, minerId, lotusJsonRpcResult.Error.Message)
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

type ClientRetrieveDealParamDataPartOne struct {
	Root        Cid
	Size        int
	Total       string
	UnsealPrice string
	Client      string
	Miner       string
}

type ClientRetrieveDealParamDataPartTwo struct {
	Path  string
	IsCAR bool
}
