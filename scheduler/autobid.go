package scheduler

import (
	"bufio"
	"encoding/json"
	"fmt"
	clientmodel "github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	libutils "github.com/filswan/go-swan-lib/utils"
	"github.com/minio/minio/cmd"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	oshomedir "github.com/mitchellh/go-homedir"
	"github.com/robfig/cron"
	"github.com/shopspring/decimal"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"time"
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
	UpdateBackupTasksInDb(tasks)
	return err
}

func UpdateBackupTasksInDb(tasks [][]*libmodel.FileDesc) {
	//open backup db
	expandedDir, err := cmd.LevelDbBackupPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
	}
	defer db.Close()

	//get backuptasks
	backupTasksKey := cmd.TableVolumeBackupTask
	backupTasks, err := db.Get([]byte(backupTasksKey), nil)
	if err != nil || backupTasks == nil {
		logs.GetLogger().Error(err)
		return
	}
	data := cmd.VolumeBackupTasks{}
	err = json.Unmarshal(backupTasks, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	//update backuptasks
	for _, v := range tasks {
		for j, value := range data.VolumeBackupPlans {
			for k, values := range value.BackupPlanTasks {
				if values.Data[0].Uuid == v[0].Uuid {
					timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Status = cmd.StatusBackupTaskRunning
					data.VolumeBackupPlans[j].BackupPlanTasks[k].UpdatedOn = timestamp
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Data[0].MinerId = v[0].MinerFid
					data.VolumeBackupPlans[j].BackupPlanTasks[k].Data[0].DealCid = v[0].DealCid
					break
				}
			}
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	err = db.Put([]byte(backupTasksKey), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
}

func SendAutobidDealSchedulers(confDeal *clientmodel.ConfDeal) error {
	startEpoch := libutils.GetCurrentEpoch() + (96+1)*libconstants.EPOCH_PER_HOUR
	fmt.Println(startEpoch)
	csvFilepaths, _, err := SendAutoBidDeals(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	for _, csvFilepath := range csvFilepaths {
		logs.GetLogger().Info(csvFilepath, " is generated")
	}
	return err
}

func SendAutoBidDeals(confDeal *clientmodel.ConfDeal) ([]string, [][]*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	assignedTasks, err := swanClient.SwanGetAssignedTasks()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}
	logs.GetLogger().Info("autobid Swan task count:", len(assignedTasks))
	if len(assignedTasks) == 0 {
		logs.GetLogger().Info("no autobid task to be dealt with")
		return nil, nil, nil
	}

	var tasksDeals [][]*libmodel.FileDesc
	csvFilepaths := []string{}
	for _, assignedTask := range assignedTasks {
		if !IsTaskSourceRight(confDeal, assignedTask) {
			continue
		}

		_, csvFilePath, carFiles, err := SendAutoBidDealsByTaskUuid(confDeal, assignedTask.Uuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		tasksDeals = append(tasksDeals, carFiles)
		csvFilepaths = append(csvFilepaths, csvFilePath)
	}

	return csvFilepaths, tasksDeals, nil
}

func CreateOutputDir(outputDir string) error {
	if len(outputDir) == 0 {
		err := fmt.Errorf("output dir is not provided")
		logs.GetLogger().Info(err)
		return err
	}

	if libutils.IsDirExists(outputDir) {
		return nil
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		err := fmt.Errorf("%s, failed to create output dir:%s", err.Error(), outputDir)
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info(outputDir, " created")
	return nil
}

func IsTaskSourceRight(confDeal *clientmodel.ConfDeal, task libmodel.Task) bool {
	if confDeal == nil {
		return false
	}

	if confDeal.DealSourceIds == nil || len(confDeal.DealSourceIds) == 0 {
		return false
	}

	for _, sourceId := range confDeal.DealSourceIds {
		if task.SourceId == sourceId {
			return true
		}
	}

	return false
}

func SendAutoBidDealsByTaskUuid(confDeal *clientmodel.ConfDeal, taskUuid string) (int, string, []*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	assignedTaskInfo, err := swanClient.SwanGetOfflineDealsByTaskUuid(taskUuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	deals := assignedTaskInfo.Data.Deal
	task := assignedTaskInfo.Data.Task

	if task.Type == libconstants.TASK_TYPE_VERIFIED {
		isWalletVerified, err := swanClient.CheckDatacap(confDeal.SenderWallet)
		if err != nil {
			logs.GetLogger().Error(err)
			return 0, "", nil, err
		}

		if !isWalletVerified {
			err := fmt.Errorf("task:%s is verified, but your wallet:%s is not verified", task.TaskName, confDeal.SenderWallet)
			logs.GetLogger().Error(err)
			return 0, "", nil, err
		}
	}

	dealSentNum, csvFilePath, carFiles, err := SendAutobidDeals4Task(confDeal, deals, task, confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	msg := fmt.Sprintf("%d deal(s) sent to:%s for task:%s", dealSentNum, confDeal.MinerFid, task.TaskName)
	logs.GetLogger().Info(msg)

	if dealSentNum == 0 {
		err := fmt.Errorf("no deal sent for task:%s", task.TaskName)
		logs.GetLogger().Info(err)
		return 0, "", nil, err
	}

	status := libconstants.TASK_STATUS_DEAL_SENT
	if dealSentNum != len(deals) {
		status = libconstants.TASK_STATUS_PROGRESS_WITH_FAILURE
	}

	response, err := swanClient.SwanUpdateAssignedTask(taskUuid, status, csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	logs.GetLogger().Info(response.Message)

	return dealSentNum, csvFilePath, carFiles, nil
}

func SendAutobidDeals4Task(confDeal *clientmodel.ConfDeal, deals []libmodel.OfflineDeal, task libmodel.Task, outputDir string) (int, string, []*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	if !IsTaskSourceRight(confDeal, task) {
		err := fmt.Errorf("you cannot send deal from this kind of source")
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	carFiles := []*libmodel.FileDesc{}

	dealSentNum := 0
	for _, deal := range deals {
		deal.DealCid = strings.Trim(deal.DealCid, " ")
		if len(deal.DealCid) != 0 {
			dealSentNum = dealSentNum + 1
			continue
		}

		err := clientmodel.SetDealConfig4Autobid(confDeal, task, deal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		err = CheckDealConfig(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		fileSizeInt := libutils.GetInt64FromStr(deal.FileSize)
		if fileSizeInt <= 0 {
			logs.GetLogger().Error("file is too small")
			continue
		}
		pieceSize, sectorSize := libutils.CalculatePieceSize(fileSizeInt)
		logs.GetLogger().Info("dealConfig.MinerPrice:", confDeal.MinerPrice)
		cost := libutils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
		carFile := libmodel.FileDesc{
			Uuid:        task.Uuid,
			MinerFid:    task.MinerFid,
			CarFileUrl:  deal.FileSourceUrl,
			CarFileMd5:  deal.Md5Origin,
			PieceCid:    deal.PieceCid,
			DataCid:     deal.PayloadCid,
			CarFileSize: libutils.GetInt64FromStr(deal.FileSize),
		}
		if carFile.MinerFid != "" {
			logs.GetLogger().Info("MinerFid:", carFile.MinerFid)
		}

		logs.GetLogger().Info("FileSourceUrl:", carFile.CarFileUrl)
		carFiles = append(carFiles, &carFile)
		for i := 0; i < 60; i++ {
			msg := fmt.Sprintf("send deal for task:%s, deal:%d", task.TaskName, deal.Id)
			logs.GetLogger().Info(msg)
			dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, confDeal.StartEpoch, confDeal.Duration, confDeal.MinerFid, confDeal.SenderWallet)

			lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
			if err != nil {
				logs.GetLogger().Error(err)
				return 0, "", nil, err
			}

			dealCid, startEpoch, err := LotusClientStartDeal(lotusClient, carFile, cost, pieceSize, *dealConfig, i)
			if err != nil {
				logs.GetLogger().Error("tried ", i, " times,", err)

				if strings.Contains(err.Error(), "already tracking identifier") {
					continue
				} else {
					break
				}
			}
			if dealCid == nil {
				logs.GetLogger().Info("no deal CID returned")
				continue
			}

			carFile.DealCid = *dealCid
			carFile.StartEpoch = startEpoch

			dealCost, err := lotusClient.LotusClientGetDealInfo(carFile.DealCid)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}

			//logs.GetLogger().Info(*dealCost)
			carFile.Cost = dealCost.CostComputed
			dealSentNum = dealSentNum + 1

			logs.GetLogger().Info("task:", task.TaskName, ", deal CID:", carFile.DealCid, ", start epoch:", *carFile.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")
			break
		}
	}

	return dealSentNum, "csvFilepath", carFiles, nil
}

func CheckDealConfig(confDeal *clientmodel.ConfDeal) error {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return err
	}

	lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	minerPrice, minerVerifiedPrice, _, _ := lotusClient.LotusGetMinerConfig(confDeal.MinerFid)

	if confDeal.SenderWallet == "" {
		err := fmt.Errorf("wallet should be set")
		logs.GetLogger().Error(err)
		return err
	}

	if confDeal.VerifiedDeal {
		if minerVerifiedPrice == nil {
			err := fmt.Errorf("miner:%s,cannot get miner verified price for verified deal", confDeal.MinerFid)
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("miner:", confDeal.MinerFid, ",price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			err := fmt.Errorf("miner:%s,cannot get miner price for non-verified deal", confDeal.MinerFid)
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerPrice
		//logs.GetLogger().Info("miner:", confDeal.MinerFid, ",price is:", *minerPrice)
	}

	priceCmp := confDeal.MaxPrice.Cmp(confDeal.MinerPrice)
	//logs.GetLogger().Info("priceCmp:", priceCmp)
	if priceCmp < 0 {
		logs.GetLogger().Info("Miner price is:", confDeal.MinerPrice, " MaxPrice:", confDeal.MaxPrice, " VerifiedDeal:", confDeal.VerifiedDeal)
		err := fmt.Errorf("miner price is higher than deal max price")
		logs.GetLogger().Error(err)
		return err
	}

	//logs.GetLogger().Info("Deal check passed.")

	if confDeal.Duration == 0 {
		confDeal.Duration = DURATION
	}

	err = CheckDuration(confDeal.Duration, confDeal.StartEpoch, confDeal.RelativeEpochFromMainNetwork)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func CheckDuration(duration int, startEpoch int, relativeEpochFromMainNetwork int) error {
	if duration == 0 {
		return nil
	}

	if duration < DURATION_MIN || duration > DURATION_MAX {
		err := fmt.Errorf("deal duration out of bounds (min, max, provided): %d, %d, %d", DURATION_MIN, DURATION_MAX, duration)
		logs.GetLogger().Error(err)
		return err
	}

	currentEpoch := libutils.GetCurrentEpoch() + relativeEpochFromMainNetwork
	endEpoch := startEpoch + duration

	epoch2EndfromNow := endEpoch - currentEpoch
	if epoch2EndfromNow >= DURATION_MAX {
		err := fmt.Errorf("invalid deal end epoch %d: cannot be more than %d past current epoch %d", endEpoch, DURATION_MAX, currentEpoch)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

const (
	DURATION               = 1512000
	DURATION_MIN           = 518400
	DURATION_MAX           = 1540000
	LOTUS_JSON_RPC_ID      = 7878
	LOTUS_JSON_RPC_VERSION = "2.0"
)

func LotusClientStartDeal(lotusClient *lotus.LotusClient, carFile libmodel.FileDesc, cost decimal.Decimal, pieceSize int64, dealConfig libmodel.DealConfig, relativeEpoch int) (*string, *int, error) {
	epochPrice := cost.Mul(decimal.NewFromFloat(libconstants.LOTUS_PRICE_MULTIPLE))
	epochPrice = epochPrice.Ceil()
	startEpoch := dealConfig.StartEpoch - relativeEpoch

	logs.GetLogger().Info("wallet:", dealConfig.SenderWallet)
	logs.GetLogger().Info("miner:", dealConfig.MinerFid)
	logs.GetLogger().Info("start epoch:", startEpoch)
	logs.GetLogger().Info("price:", dealConfig.MinerPrice)
	logs.GetLogger().Info("cost per epoch(filecoin):", cost.String())
	logs.GetLogger().Info("fast-retrieval:", dealConfig.FastRetrieval)
	logs.GetLogger().Info("verified-deal:", dealConfig.VerifiedDeal)
	logs.GetLogger().Info("duration:", dealConfig.Duration)
	logs.GetLogger().Info("data CID:", carFile.DataCid)
	logs.GetLogger().Info("piece CID:", carFile.PieceCid)
	logs.GetLogger().Info("piece size:", pieceSize)

	if !dealConfig.SkipConfirmation {
		logs.GetLogger().Info("Do you confirm to submit the deal?")
		logs.GetLogger().Info("Press Y/y to continue, other key to quit")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		response = strings.TrimRight(response, "\n")

		if !strings.EqualFold(response, "Y") {
			logs.GetLogger().Info("Your input is ", response, ". Now give up submit the deal.")
			return nil, nil, nil
		}
	}

	clientStartDealParamData := ClientStartDealParamData{
		TransferType: "string value",
		Root: Cid{
			Cid: carFile.DataCid,
		},
		PieceCid: Cid{
			Cid: carFile.PieceCid,
		},
		PieceSize: int(pieceSize),
	}

	clientStartDealParam := ClientStartDealParam{
		Data:              clientStartDealParamData,
		Wallet:            dealConfig.SenderWallet,
		Miner:             dealConfig.MinerFid,
		EpochPrice:        epochPrice.String(),
		MinBlocksDuration: dealConfig.Duration,
		DealStartEpoch:    startEpoch,
		FastRetrieval:     dealConfig.FastRetrieval,
		VerifiedDeal:      dealConfig.VerifiedDeal,
	}

	var params []interface{}
	params = append(params, clientStartDealParam)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  "Filecoin.ClientStartDeal",
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}
	res2B, _ := json.Marshal(jsonRpcParams)
	fmt.Println(string(res2B))
	response := client.HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if response == "" {
		err := fmt.Errorf("failed to send deal for %s, no response", carFile.CarFileName)
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	clientStartDeal := &ClientStartDeal{}
	err := json.Unmarshal([]byte(response), clientStartDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	if clientStartDeal.Error != nil {
		err := fmt.Errorf("error, code:%d, message:%s", clientStartDeal.Error.Code, clientStartDeal.Error.Message)
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	logs.GetLogger().Info("deal CID:", clientStartDeal.Result.Cid)
	return &clientStartDeal.Result.Cid, &startEpoch, nil
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
