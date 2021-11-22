package scheduler

import (
	"fmt"
	clientmodel "github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-client/subcommand"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libutils "github.com/filswan/go-swan-lib/utils"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	"github.com/robfig/cron"
	"time"
)

func SendDealScheduler() {
	confDeal := &clientmodel.ConfDeal{
		SwanApiUrl:                   config.GetUserConfig().SwanAddress,
		SwanToken:                    config.GetUserConfig().SwanToken,
		LotusClientApiUrl:            config.GetUserConfig().LotusClientApiUrl,
		LotusClientAccessToken:       config.GetUserConfig().LotusClientAccessToken,
		SenderWallet:                 config.GetUserConfig().Fs3WalletAddress,
		OutputDir:                    config.GetUserConfig().VolumeBackupAddress,
		Duration:                     1051200,
		RelativeEpochFromMainNetwork: -858481,
	}
	confDeal.DealSourceIds = append(confDeal.DealSourceIds, libconstants.TASK_SOURCE_ID_SWAN_FS3)

	c := cron.New()
	err := c.AddFunc("0 */3 * * * ?", func() {
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
	csvFilepaths, _, err := subcommand.SendAutoBidDeals(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	for _, csvFilepath := range csvFilepaths {
		logs.GetLogger().Info(csvFilepath, " is generated")
	}
	return err
}
