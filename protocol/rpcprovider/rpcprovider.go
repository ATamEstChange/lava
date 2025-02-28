package rpcprovider

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lavanet/lava/app"
	"github.com/lavanet/lava/protocol/chainlib"
	"github.com/lavanet/lava/protocol/chainlib/chainproxy"
	"github.com/lavanet/lava/protocol/chaintracker"
	"github.com/lavanet/lava/protocol/common"
	"github.com/lavanet/lava/protocol/lavasession"
	"github.com/lavanet/lava/protocol/metrics"
	"github.com/lavanet/lava/protocol/performance"
	"github.com/lavanet/lava/protocol/rpcprovider/reliabilitymanager"
	"github.com/lavanet/lava/protocol/rpcprovider/rewardserver"
	"github.com/lavanet/lava/protocol/statetracker"
	"github.com/lavanet/lava/protocol/statetracker/updaters"
	"github.com/lavanet/lava/protocol/upgrade"
	"github.com/lavanet/lava/utils"
	"github.com/lavanet/lava/utils/rand"
	"github.com/lavanet/lava/utils/sigs"
	epochstorage "github.com/lavanet/lava/x/epochstorage/types"
	pairingtypes "github.com/lavanet/lava/x/pairing/types"
	protocoltypes "github.com/lavanet/lava/x/protocol/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ChainTrackerDefaultMemory  = 100
	DEFAULT_ALLOWED_MISSING_CU = 0.2

	ShardIDFlagName           = "shard-id"
	StickinessHeaderName      = "sticky-header"
	DefaultShardID       uint = 0
)

var (
	Yaml_config_properties     = []string{"network-address.address", "chain-id", "api-interface", "node-urls.url"}
	DefaultRPCProviderFileName = "rpcprovider.yml"

	RelaysHealthEnableFlagDefault  = true
	RelayHealthIntervalFlagDefault = 5 * time.Minute
)

// used to call SetPolicy in base chain parser so we are allowed to run verifications on the addons and extensions
type ProviderPolicy struct {
	addons     []string
	extensions []epochstorage.EndpointService
}

func (pp *ProviderPolicy) GetSupportedAddons(specID string) (addons []string, err error) {
	return pp.addons, nil
}

func (pp *ProviderPolicy) GetSupportedExtensions(specID string) (extensions []epochstorage.EndpointService, err error) {
	return pp.extensions, nil
}

type ProviderStateTrackerInf interface {
	RegisterForVersionUpdates(ctx context.Context, version *protocoltypes.Version, versionValidator updaters.VersionValidationInf)
	RegisterForSpecUpdates(ctx context.Context, specUpdatable updaters.SpecUpdatable, endpoint lavasession.RPCEndpoint) error
	RegisterForSpecVerifications(ctx context.Context, specVerifier updaters.SpecVerifier, chainId string) error
	RegisterReliabilityManagerForVoteUpdates(ctx context.Context, voteUpdatable updaters.VoteUpdatable, endpointP *lavasession.RPCProviderEndpoint)
	RegisterForEpochUpdates(ctx context.Context, epochUpdatable updaters.EpochUpdatable)
	RegisterForDowntimeParamsUpdates(ctx context.Context, downtimeParamsUpdatable updaters.DowntimeParamsUpdatable) error
	TxRelayPayment(ctx context.Context, relayRequests []*pairingtypes.RelaySession, description string, latestBlocks []*pairingtypes.LatestBlockReport) error
	SendVoteReveal(voteID string, vote *reliabilitymanager.VoteData) error
	SendVoteCommitment(voteID string, vote *reliabilitymanager.VoteData) error
	LatestBlock() int64
	GetMaxCuForUser(ctx context.Context, consumerAddress, chainID string, epocu uint64) (maxCu uint64, err error)
	VerifyPairing(ctx context.Context, consumerAddress, providerAddress string, epoch uint64, chainID string) (valid bool, total int64, projectId string, err error)
	GetEpochSize(ctx context.Context) (uint64, error)
	EarliestBlockInMemory(ctx context.Context) (uint64, error)
	RegisterPaymentUpdatableForPayments(ctx context.Context, paymentUpdatable updaters.PaymentUpdatable)
	GetRecommendedEpochNumToCollectPayment(ctx context.Context) (uint64, error)
	GetEpochSizeMultipliedByRecommendedEpochNumToCollectPayment(ctx context.Context) (uint64, error)
	GetProtocolVersion(ctx context.Context) (*updaters.ProtocolVersionResponse, error)
	GetVirtualEpoch(epoch uint64) uint64
	GetAverageBlockTime() time.Duration
}

type rpcProviderStartOptions struct {
	ctx                       context.Context
	txFactory                 tx.Factory
	clientCtx                 client.Context
	rpcProviderEndpoints      []*lavasession.RPCProviderEndpoint
	cache                     *performance.Cache
	parallelConnections       uint
	metricsListenAddress      string
	rewardStoragePath         string
	rewardTTL                 time.Duration
	shardID                   uint
	rewardsSnapshotThreshold  uint
	rewardsSnapshotTimeoutSec uint
	healthCheckMetricsOptions *rpcProviderHealthCheckMetricsOptions
}

type rpcProviderHealthCheckMetricsOptions struct {
	relaysHealthEnableFlag   bool          // enables relay health check
	relaysHealthIntervalFlag time.Duration // interval for relay health check
	grpcHealthCheckEndpoint  string
}

type RPCProvider struct {
	providerStateTracker ProviderStateTrackerInf
	rpcProviderListeners map[string]*ProviderListener
	lock                 sync.Mutex
	// all of the following members need to be concurrency proof
	providerMetricsManager    *metrics.ProviderMetricsManager
	rewardServer              *rewardserver.RewardServer
	privKey                   *btcec.PrivateKey
	lavaChainID               string
	addr                      sdk.AccAddress
	blockMemorySize           uint64
	chainMutexes              map[string]*sync.Mutex
	parallelConnections       uint
	cache                     *performance.Cache
	shardID                   uint // shardID is a flag that allows setting up multiple provider databases of the same chain
	chainTrackers             *ChainTrackers
	relaysMonitorAggregator   *metrics.RelaysMonitorAggregator
	relaysHealthCheckEnabled  bool
	relaysHealthCheckInterval time.Duration
	grpcHealthCheckEndpoint   string
}

func (rpcp *RPCProvider) Start(options *rpcProviderStartOptions) (err error) {
	ctx, cancel := context.WithCancel(options.ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	rpcp.chainTrackers = &ChainTrackers{}
	rpcp.parallelConnections = options.parallelConnections
	rpcp.cache = options.cache
	rpcp.providerMetricsManager = metrics.NewProviderMetricsManager(options.metricsListenAddress) // start up prometheus metrics
	rpcp.providerMetricsManager.SetVersion(upgrade.GetCurrentVersion().ProviderVersion)
	rpcp.rpcProviderListeners = make(map[string]*ProviderListener)
	rpcp.shardID = options.shardID
	rpcp.relaysHealthCheckEnabled = options.healthCheckMetricsOptions.relaysHealthEnableFlag
	rpcp.relaysHealthCheckInterval = options.healthCheckMetricsOptions.relaysHealthIntervalFlag
	rpcp.relaysMonitorAggregator = metrics.NewRelaysMonitorAggregator(rpcp.relaysHealthCheckInterval, rpcp.providerMetricsManager)
	rpcp.grpcHealthCheckEndpoint = options.healthCheckMetricsOptions.grpcHealthCheckEndpoint
	// single state tracker
	lavaChainFetcher := chainlib.NewLavaChainFetcher(ctx, options.clientCtx)
	providerStateTracker, err := statetracker.NewProviderStateTracker(ctx, options.txFactory, options.clientCtx, lavaChainFetcher, rpcp.providerMetricsManager)
	if err != nil {
		return err
	}

	rpcp.providerStateTracker = providerStateTracker
	providerStateTracker.RegisterForUpdates(ctx, updaters.NewMetricsUpdater(rpcp.providerMetricsManager))
	// check version
	version, err := rpcp.providerStateTracker.GetProtocolVersion(ctx)
	if err != nil {
		utils.LavaFormatFatal("failed fetching protocol version from node", err)
	}
	rpcp.providerStateTracker.RegisterForVersionUpdates(ctx, version.Version, &upgrade.ProtocolVersion{})

	// single reward server
	rewardDB := rewardserver.NewRewardDBWithTTL(options.rewardTTL)
	rpcp.rewardServer = rewardserver.NewRewardServer(providerStateTracker, rpcp.providerMetricsManager, rewardDB, options.rewardStoragePath, options.rewardsSnapshotThreshold, options.rewardsSnapshotTimeoutSec, rpcp.chainTrackers)
	rpcp.providerStateTracker.RegisterForEpochUpdates(ctx, rpcp.rewardServer)
	rpcp.providerStateTracker.RegisterPaymentUpdatableForPayments(ctx, rpcp.rewardServer)
	keyName, err := sigs.GetKeyName(options.clientCtx)
	if err != nil {
		utils.LavaFormatFatal("failed getting key name from clientCtx", err)
	}
	privKey, err := sigs.GetPrivKey(options.clientCtx, keyName)
	if err != nil {
		utils.LavaFormatFatal("failed getting private key from key name", err, utils.Attribute{Key: "keyName", Value: keyName})
	}
	rpcp.privKey = privKey
	clientKey, _ := options.clientCtx.Keyring.Key(keyName)
	rpcp.lavaChainID = options.clientCtx.ChainID

	pubKey, err := clientKey.GetPubKey()
	if err != nil {
		return err
	}

	err = rpcp.addr.Unmarshal(pubKey.Address())
	if err != nil {
		utils.LavaFormatFatal("failed unmarshaling public address", err, utils.Attribute{Key: "keyName", Value: keyName}, utils.Attribute{Key: "pubkey", Value: pubKey.Address()})
	}
	utils.LavaFormatInfo("RPCProvider pubkey: " + rpcp.addr.String())
	utils.LavaFormatInfo("RPCProvider setting up endpoints", utils.Attribute{Key: "count", Value: strconv.Itoa(len(options.rpcProviderEndpoints))})
	blockMemorySize, err := rpcp.providerStateTracker.GetEpochSizeMultipliedByRecommendedEpochNumToCollectPayment(ctx) // get the number of blocks to keep in PSM.
	if err != nil {
		utils.LavaFormatFatal("Failed fetching GetEpochSizeMultipliedByRecommendedEpochNumToCollectPayment in RPCProvider Start", err)
	}
	rpcp.blockMemorySize = blockMemorySize
	// pre loop to handle synchronous actions
	rpcp.chainMutexes = map[string]*sync.Mutex{}
	for idx, endpoint := range options.rpcProviderEndpoints {
		rpcp.chainMutexes[endpoint.ChainID] = &sync.Mutex{}   // create a mutex per chain for shared resources
		if idx > 0 && endpoint.NetworkAddress.Address == "" { // handle undefined addresses as the previous endpoint for shared listeners
			endpoint.NetworkAddress = options.rpcProviderEndpoints[idx-1].NetworkAddress
		}
	}

	specValidator := NewSpecValidator()
	disabledEndpointsList := rpcp.SetupProviderEndpoints(options.rpcProviderEndpoints, specValidator, true)
	rpcp.relaysMonitorAggregator.StartMonitoring(ctx)
	specValidator.Start(ctx)
	utils.LavaFormatInfo("RPCProvider done setting up endpoints, ready for service")
	if len(disabledEndpointsList) > 0 {
		utils.LavaFormatError(utils.FormatStringerList("[-] RPCProvider running with disabled endpoints:", disabledEndpointsList, "[-]"), nil)
		if len(disabledEndpointsList) == len(options.rpcProviderEndpoints) {
			utils.LavaFormatFatal("all endpoints are disabled", nil)
		}
		activeEndpointsList := getActiveEndpoints(options.rpcProviderEndpoints, disabledEndpointsList)
		utils.LavaFormatInfo(utils.FormatStringerList("[+] active endpoints:", activeEndpointsList, "[+]"))
		// try to save disabled endpoints
		go rpcp.RetryDisabledEndpoints(disabledEndpointsList, specValidator, 1)
	} else {
		utils.LavaFormatInfo("[+] all endpoints up and running")
	}
	// tearing down
	select {
	case <-ctx.Done():
		utils.LavaFormatInfo("Provider Server ctx.Done")
	case <-signalChan:
		utils.LavaFormatInfo("Provider Server signalChan")
	}

	for _, listener := range rpcp.rpcProviderListeners {
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		listener.Shutdown(shutdownCtx)
		defer shutdownRelease()
	}

	// close all reward dbs
	err = rpcp.rewardServer.CloseAllDataBases()
	if err != nil {
		utils.LavaFormatError("failed to close reward db", err)
	}

	return nil
}

func getActiveEndpoints(rpcProviderEndpoints []*lavasession.RPCProviderEndpoint, disabledEndpointsList []*lavasession.RPCProviderEndpoint) []*lavasession.RPCProviderEndpoint {
	activeEndpoints := map[*lavasession.RPCProviderEndpoint]struct{}{}
	for _, endpoint := range rpcProviderEndpoints {
		activeEndpoints[endpoint] = struct{}{}
	}
	for _, disabled := range disabledEndpointsList {
		delete(activeEndpoints, disabled)
	}
	activeEndpointsList := []*lavasession.RPCProviderEndpoint{}
	for endpoint := range activeEndpoints {
		activeEndpointsList = append(activeEndpointsList, endpoint)
	}
	return activeEndpointsList
}

func (rpcp *RPCProvider) RetryDisabledEndpoints(disabledEndpoints []*lavasession.RPCProviderEndpoint, specValidator *SpecValidator, retryCount int) {
	time.Sleep(time.Duration(retryCount) * time.Second)
	parallel := retryCount > 2
	utils.LavaFormatInfo("Retrying disabled endpoints", utils.Attribute{Key: "disabled endpoints list", Value: disabledEndpoints}, utils.Attribute{Key: "parallel", Value: parallel})
	disabledEndpointsAfterRetry := rpcp.SetupProviderEndpoints(disabledEndpoints, specValidator, parallel)
	if len(disabledEndpointsAfterRetry) > 0 {
		utils.LavaFormatError(utils.FormatStringerList("RPCProvider running with disabled endpoints:", disabledEndpointsAfterRetry, "[-]"), nil)
		rpcp.RetryDisabledEndpoints(disabledEndpointsAfterRetry, specValidator, retryCount+1)
		activeEndpointsList := getActiveEndpoints(disabledEndpoints, disabledEndpointsAfterRetry)
		utils.LavaFormatInfo(utils.FormatStringerList("[+] active endpoints:", activeEndpointsList, "[+]"))
	} else {
		utils.LavaFormatInfo("[+] all endpoints up and running")
	}
}

func (rpcp *RPCProvider) SetupProviderEndpoints(rpcProviderEndpoints []*lavasession.RPCProviderEndpoint, specValidator *SpecValidator, parallel bool) (disabledEndpointsRet []*lavasession.RPCProviderEndpoint) {
	var wg sync.WaitGroup
	parallelJobs := len(rpcProviderEndpoints)
	wg.Add(parallelJobs)
	disabledEndpoints := make(chan *lavasession.RPCProviderEndpoint, parallelJobs)
	for _, rpcProviderEndpoint := range rpcProviderEndpoints {
		setupEndpoint := func(rpcProviderEndpoint *lavasession.RPCProviderEndpoint, specValidator *SpecValidator) {
			defer wg.Done()
			err := rpcp.SetupEndpoint(context.Background(), rpcProviderEndpoint, specValidator)
			if err != nil {
				rpcp.providerMetricsManager.SetDisabledChain(rpcProviderEndpoint.ChainID, rpcProviderEndpoint.ApiInterface)
				disabledEndpoints <- rpcProviderEndpoint
			}
		}
		if parallel {
			go setupEndpoint(rpcProviderEndpoint, specValidator)
		} else {
			setupEndpoint(rpcProviderEndpoint, specValidator)
		}
	}
	wg.Wait()
	close(disabledEndpoints)
	disabledEndpointsList := []*lavasession.RPCProviderEndpoint{}
	for disabledEndpoint := range disabledEndpoints {
		disabledEndpointsList = append(disabledEndpointsList, disabledEndpoint)
	}
	return disabledEndpointsList
}

func (rpcp *RPCProvider) getAllAddonsAndExtensionsFromNodeUrlSlice(nodeUrls []common.NodeUrl) *ProviderPolicy {
	policy := &ProviderPolicy{}
	for _, nodeUrl := range nodeUrls {
		policy.addons = append(policy.addons, nodeUrl.Addons...) // addons are added without validation while extensions are. so we add to the addons all.
	}
	return policy
}

func (rpcp *RPCProvider) SetupEndpoint(ctx context.Context, rpcProviderEndpoint *lavasession.RPCProviderEndpoint, specValidator *SpecValidator) error {
	err := rpcProviderEndpoint.Validate()
	if err != nil {
		return utils.LavaFormatError("[PANIC] panic severity critical error, aborting support for chain api due to invalid node url definition, continuing with others", err, utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint.String()})
	}
	chainID := rpcProviderEndpoint.ChainID
	apiInterface := rpcProviderEndpoint.ApiInterface
	providerSessionManager := lavasession.NewProviderSessionManager(rpcProviderEndpoint, rpcp.blockMemorySize)
	rpcp.providerStateTracker.RegisterForEpochUpdates(ctx, providerSessionManager)
	chainParser, err := chainlib.NewChainParser(apiInterface)
	if err != nil {
		return utils.LavaFormatError("[PANIC] panic severity critical error, aborting support for chain api due to invalid chain parser, continuing with others", err, utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint.String()})
	}

	rpcEndpoint := lavasession.RPCEndpoint{ChainID: chainID, ApiInterface: apiInterface}
	err = rpcp.providerStateTracker.RegisterForSpecUpdates(ctx, chainParser, rpcEndpoint)
	if err != nil {
		return utils.LavaFormatError("[PANIC] failed to RegisterForSpecUpdates, panic severity critical error, aborting support for chain api due to invalid chain parser, continuing with others", err, utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint.String()})
	}

	// after registering for spec updates our chain parser contains the spec and we can add our addons and extensions to allow our provider to function properly
	providerPolicy := rpcp.getAllAddonsAndExtensionsFromNodeUrlSlice(rpcProviderEndpoint.NodeUrls)
	utils.LavaFormatDebug("supported services for provider",
		utils.LogAttr("specId", rpcProviderEndpoint.ChainID),
		utils.LogAttr("apiInterface", apiInterface),
		utils.LogAttr("supportedServices", providerPolicy.addons))
	chainParser.SetPolicy(providerPolicy, rpcProviderEndpoint.ChainID, apiInterface)
	chainRouter, err := chainlib.GetChainRouter(ctx, rpcp.parallelConnections, rpcProviderEndpoint, chainParser)
	if err != nil {
		return utils.LavaFormatError("[PANIC] panic severity critical error, failed creating chain proxy, continuing with others endpoints", err, utils.Attribute{Key: "parallelConnections", Value: uint64(rpcp.parallelConnections)}, utils.Attribute{Key: "rpcProviderEndpoint", Value: rpcProviderEndpoint})
	}

	_, averageBlockTime, blocksToFinalization, blocksInFinalizationData := chainParser.ChainBlockStats()
	var chainTracker *chaintracker.ChainTracker
	// chainTracker accepts a callback to be called on new blocks, we use this to call metrics update on a new block
	recordMetricsOnNewBlock := func(blockFrom int64, blockTo int64, hash string) {
		for block := blockFrom + 1; block <= blockTo; block++ {
			rpcp.providerMetricsManager.SetLatestBlock(chainID, uint64(block))
		}
	}
	var chainFetcher chainlib.ChainFetcherIf
	if enabled, _ := chainParser.DataReliabilityParams(); enabled {
		chainFetcher = chainlib.NewChainFetcher(
			ctx,
			&chainlib.ChainFetcherOptions{
				ChainRouter: chainRouter,
				ChainParser: chainParser,
				Endpoint:    rpcProviderEndpoint,
				Cache:       rpcp.cache,
			},
		)
	} else {
		chainFetcher = chainlib.NewVerificationsOnlyChainFetcher(ctx, chainRouter, chainParser, rpcProviderEndpoint)
	}

	// check the chain fetcher verification works, if it doesn't we disable the chain+apiInterface and this triggers a boot retry
	err = chainFetcher.Validate(ctx)
	if err != nil {
		return utils.LavaFormatError("[PANIC] Failed starting due to chain fetcher validation failure", err,
			utils.Attribute{Key: "Chain", Value: rpcProviderEndpoint.ChainID},
			utils.Attribute{Key: "apiInterface", Value: apiInterface})
	}

	// in order to utilize shared resources between chains we need go routines with the same chain to wait for one another here
	chainCommonSetup := func() error {
		rpcp.chainMutexes[chainID].Lock()
		defer rpcp.chainMutexes[chainID].Unlock()
		var found bool
		chainTracker, found = rpcp.chainTrackers.GetTrackerPerChain(chainID)
		if !found {
			consistencyErrorCallback := func(oldBlock, newBlock int64) {
				utils.LavaFormatError("Consistency issue detected", nil,
					utils.Attribute{Key: "oldBlock", Value: oldBlock},
					utils.Attribute{Key: "newBlock", Value: newBlock},
					utils.Attribute{Key: "Chain", Value: rpcProviderEndpoint.ChainID},
					utils.Attribute{Key: "apiInterface", Value: apiInterface},
				)
			}
			blocksToSaveChainTracker := uint64(blocksToFinalization + blocksInFinalizationData)
			chainTrackerConfig := chaintracker.ChainTrackerConfig{
				BlocksToSave:        blocksToSaveChainTracker,
				AverageBlockTime:    averageBlockTime,
				ServerBlockMemory:   ChainTrackerDefaultMemory + blocksToSaveChainTracker,
				NewLatestCallback:   recordMetricsOnNewBlock,
				ConsistencyCallback: consistencyErrorCallback,
				Pmetrics:            rpcp.providerMetricsManager,
			}

			chainTracker, err = chaintracker.NewChainTracker(ctx, chainFetcher, chainTrackerConfig)
			if err != nil {
				return utils.LavaFormatError("panic severity critical error, aborting support for chain api due to node access, continuing with other endpoints", err, utils.Attribute{Key: "chainTrackerConfig", Value: chainTrackerConfig}, utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint})
			}

			utils.LavaFormatDebug("Registering for spec verifications for endpoint", utils.LogAttr("rpcEndpoint", rpcEndpoint))
			// we register for spec verifications only once, and this triggers all chainFetchers of that specId when it triggers
			err = rpcp.providerStateTracker.RegisterForSpecVerifications(ctx, specValidator, rpcEndpoint.ChainID)
			if err != nil {
				return utils.LavaFormatError("failed to RegisterForSpecUpdates, panic severity critical error, aborting support for chain api due to invalid chain parser, continuing with others", err, utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint.String()})
			}

			// Any validation needs to be before we store chain tracker for given chain id
			rpcp.chainTrackers.SetTrackerForChain(rpcProviderEndpoint.ChainID, chainTracker)
		} else {
			utils.LavaFormatDebug("reusing chain tracker", utils.Attribute{Key: "chain", Value: rpcProviderEndpoint.ChainID})
		}
		return nil
	}
	err = chainCommonSetup()
	if err != nil {
		return err
	}

	// Add the chain fetcher to the spec validator
	err = specValidator.AddChainFetcher(ctx, &chainFetcher, chainID)
	if err != nil {
		return utils.LavaFormatError("panic severity critical error, failed validating chain", err, utils.Attribute{Key: "rpcProviderEndpoint", Value: rpcProviderEndpoint})
	}

	providerMetrics := rpcp.providerMetricsManager.AddProviderMetrics(chainID, apiInterface)

	reliabilityManager := reliabilitymanager.NewReliabilityManager(chainTracker, rpcp.providerStateTracker, rpcp.addr.String(), chainRouter, chainParser)
	rpcp.providerStateTracker.RegisterReliabilityManagerForVoteUpdates(ctx, reliabilityManager, rpcProviderEndpoint)

	// add a database for this chainID if does not exist.
	rpcp.rewardServer.AddDataBase(rpcProviderEndpoint.ChainID, rpcp.addr.String(), rpcp.shardID)

	var relaysMonitor *metrics.RelaysMonitor = nil
	if rpcp.relaysHealthCheckEnabled {
		relaysMonitor = metrics.NewRelaysMonitor(rpcp.relaysHealthCheckInterval, chainID, apiInterface)
		rpcp.relaysMonitorAggregator.RegisterRelaysMonitor(rpcEndpoint.Key(), relaysMonitor)
		rpcp.providerMetricsManager.RegisterRelaysMonitor(chainID, apiInterface, relaysMonitor)
	}

	rpcProviderServer := &RPCProviderServer{}
	rpcProviderServer.ServeRPCRequests(ctx, rpcProviderEndpoint, chainParser, rpcp.rewardServer, providerSessionManager, reliabilityManager, rpcp.privKey, rpcp.cache, chainRouter, rpcp.providerStateTracker, rpcp.addr, rpcp.lavaChainID, DEFAULT_ALLOWED_MISSING_CU, providerMetrics, relaysMonitor)
	// set up grpc listener
	var listener *ProviderListener
	func() {
		rpcp.lock.Lock()
		defer rpcp.lock.Unlock()
		var ok bool
		listener, ok = rpcp.rpcProviderListeners[rpcProviderEndpoint.NetworkAddress.Address]
		if !ok {
			utils.LavaFormatDebug("creating new listener", utils.Attribute{Key: "NetworkAddress", Value: rpcProviderEndpoint.NetworkAddress})
			listener = NewProviderListener(ctx, rpcProviderEndpoint.NetworkAddress, rpcp.grpcHealthCheckEndpoint)
			specValidator.AddRPCProviderListener(rpcProviderEndpoint.NetworkAddress.Address, listener)
			rpcp.rpcProviderListeners[rpcProviderEndpoint.NetworkAddress.Address] = listener
		}
	}()
	if listener == nil {
		utils.LavaFormatFatal("listener not defined, cant register RPCProviderServer", nil, utils.Attribute{Key: "RPCProviderEndpoint", Value: rpcProviderEndpoint.String()})
	}
	err = listener.RegisterReceiver(rpcProviderServer, rpcProviderEndpoint)
	if err != nil {
		utils.LavaFormatError("error in register receiver", err)
	}
	utils.LavaFormatDebug("provider finished setting up endpoint", utils.Attribute{Key: "endpoint", Value: rpcProviderEndpoint.Key()})
	// prevents these objects form being overrun later
	chainParser.Activate()
	chainTracker.RegisterForBlockTimeUpdates(chainParser)
	rpcp.providerMetricsManager.SetEnabledChain(rpcProviderEndpoint.ChainID, apiInterface)
	return nil
}

func ParseEndpoints(viper_endpoints *viper.Viper, geolocation uint64) (endpoints []*lavasession.RPCProviderEndpoint, err error) {
	err = viper_endpoints.UnmarshalKey(common.EndpointsConfigName, &endpoints)
	if err != nil {
		utils.LavaFormatFatal("could not unmarshal endpoints", err, utils.Attribute{Key: "viper_endpoints", Value: viper_endpoints.AllSettings()})
	}
	for _, endpoint := range endpoints {
		endpoint.Geolocation = geolocation
	}
	return
}

func CreateRPCProviderCobraCommand() *cobra.Command {
	cmdRPCProvider := &cobra.Command{
		Use:   `rpcprovider [config-file] | { {listen-ip:listen-port spec-chain-id api-interface "comma-separated-node-urls"} ... } --gas-adjustment "1.5" --gas "auto" --gas-prices $GASPRICE`,
		Short: `rpcprovider sets up a server to listen for rpc-consumers requests from the lava protocol send them to a configured node and respond with the reply`,
		Long: `rpcprovider sets up a server to listen for rpc-consumers requests from the lava protocol send them to a configured node and respond with the reply
		all configs should be located in` + app.DefaultNodeHome + "/config or the local running directory" + ` 
		if no arguments are passed, assumes default config file: ` + DefaultRPCProviderFileName + `
		if one argument is passed, its assumed the config file name
		--gas-adjustment "1.5" --gas "auto" --gas-prices $GASPRICE are necessary to send reward transactions, according to the current lava gas price set in validators
		`,
		Example: `required flags: --geolocation 1 --from alice
optional: --save-conf
rpcprovider <flags>
rpcprovider rpcprovider_conf.yml <flags>
rpcprovider 127.0.0.1:3333 ETH1 jsonrpc wss://www.eth-node.com:80 <flags>
rpcprovider 127.0.0.1:3333 COS3 tendermintrpc "wss://www.node-path.com:80,https://www.node-path.com:80" 127.0.0.1:3333 COS3 rest https://www.node-path.com:1317 <flags>`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Optionally run one of the validators provided by cobra
			if err := cobra.RangeArgs(0, 1)(cmd, args); err == nil {
				// zero or one argument is allowed
				return nil
			}
			if len(args)%len(Yaml_config_properties) != 0 {
				return fmt.Errorf("invalid number of arguments, either its a single config file or repeated groups of 4 HOST:PORT chain-id api-interface [node_url,node_url_2], arg count: %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.LavaFormatInfo(common.ProcessStartLogText)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			config_name := DefaultRPCProviderFileName
			if len(args) == 1 {
				config_name = args[0] // name of config file (without extension)
			}
			viper.SetConfigName(config_name)
			viper.SetConfigType("yml")
			viper.AddConfigPath(".")
			viper.AddConfigPath("./config")

			// set log format
			logFormat := viper.GetString(flags.FlagLogFormat)
			utils.JsonFormat = logFormat == "json"
			// set rolling log.
			closeLoggerOnFinish := common.SetupRollingLogger()
			defer closeLoggerOnFinish()

			utils.LavaFormatInfo("RPCProvider started")
			var rpcProviderEndpoints []*lavasession.RPCProviderEndpoint
			var endpoints_strings []string
			var viper_endpoints *viper.Viper
			if len(args) > 1 {
				viper_endpoints, err = common.ParseEndpointArgs(args, Yaml_config_properties, common.EndpointsConfigName)
				if err != nil {
					return utils.LavaFormatError("invalid endpoints arguments", err, utils.Attribute{Key: "endpoint_strings", Value: strings.Join(args, "")})
				}
				save_config, err := cmd.Flags().GetBool(common.SaveConfigFlagName)
				if err != nil {
					return utils.LavaFormatError("failed reading flag", err, utils.Attribute{Key: "flag_name", Value: common.SaveConfigFlagName})
				}
				viper.MergeConfigMap(viper_endpoints.AllSettings())
				if save_config {
					err := viper.SafeWriteConfigAs(DefaultRPCProviderFileName)
					if err != nil {
						utils.LavaFormatInfo("did not create new config file, if it's desired remove the config file", utils.Attribute{Key: "file_name", Value: DefaultRPCProviderFileName}, utils.Attribute{Key: "error", Value: err})
					} else {
						utils.LavaFormatInfo("created new config file", utils.Attribute{Key: "file_name", Value: DefaultRPCProviderFileName})
					}
				}
			} else {
				err = viper.ReadInConfig()
				if err != nil {
					utils.LavaFormatFatal("could not load config file", err, utils.Attribute{Key: "expected_config_name", Value: viper.ConfigFileUsed()})
				}
				utils.LavaFormatInfo("read config file successfully", utils.Attribute{Key: "expected_config_name", Value: viper.ConfigFileUsed()})
			}
			geolocation, err := cmd.Flags().GetUint64(lavasession.GeolocationFlag)
			if err != nil {
				utils.LavaFormatFatal("failed to read geolocation flag, required flag", err)
			}
			rpcProviderEndpoints, err = ParseEndpoints(viper.GetViper(), geolocation)
			if err != nil || len(rpcProviderEndpoints) == 0 {
				return utils.LavaFormatError("invalid endpoints definition", err, utils.Attribute{Key: "endpoint_strings", Value: strings.Join(endpoints_strings, "")})
			}
			for _, endpoint := range rpcProviderEndpoints {
				for _, nodeUrl := range endpoint.NodeUrls {
					if nodeUrl.Url == "" {
						utils.LavaFormatError("invalid endpoint definition, empty url in nodeUrl", err, utils.Attribute{Key: "endpoint", Value: endpoint})
					}
				}
			}
			// handle flags, pass necessary fields
			ctx := context.Background()

			networkChainId := viper.GetString(flags.FlagChainID)
			if networkChainId == app.Name {
				clientTomlConfig, err := config.ReadFromClientConfig(clientCtx)
				if err == nil {
					if clientTomlConfig.ChainID != "" {
						networkChainId = clientTomlConfig.ChainID
					}
				}
			}
			utils.LavaFormatInfo("Running with chain-id:" + networkChainId)

			clientCtx = clientCtx.WithChainID(networkChainId)
			err = common.VerifyAndHandleUnsupportedFlags(cmd.Flags())
			if err != nil {
				utils.LavaFormatFatal("failed to verify cmd flags", err)
			}

			txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				utils.LavaFormatFatal("failed to create tx factory", err)
			}
			txFactory = txFactory.WithGasAdjustment(viper.GetFloat64(flags.FlagGasAdjustment))

			logLevel, err := cmd.Flags().GetString(flags.FlagLogLevel)
			if err != nil {
				utils.LavaFormatFatal("failed to read log level flag", err)
			}
			utils.SetGlobalLoggingLevel(logLevel)

			// check if the command includes --pprof-address
			pprofAddressFlagUsed := cmd.Flags().Lookup("pprof-address").Changed
			if pprofAddressFlagUsed {
				// get pprof server ip address (default value: "")
				pprofServerAddress, err := cmd.Flags().GetString("pprof-address")
				if err != nil {
					utils.LavaFormatFatal("failed to read pprof address flag", err)
				}

				// start pprof HTTP server
				err = performance.StartPprofServer(pprofServerAddress)
				if err != nil {
					return utils.LavaFormatError("failed to start pprof HTTP server", err)
				}
			}

			utils.LavaFormatInfo("lavap Binary Version: " + upgrade.GetCurrentVersion().ProviderVersion)
			rand.InitRandomSeed()
			var cache *performance.Cache = nil
			cacheAddr := viper.GetString(performance.CacheFlagName)
			if cacheAddr != "" {
				cache, err = performance.InitCache(ctx, cacheAddr)
				if err != nil {
					utils.LavaFormatError("Failed To Connect to cache at address", err, utils.Attribute{Key: "address", Value: cacheAddr})
				} else {
					utils.LavaFormatInfo("cache service connected", utils.Attribute{Key: "address", Value: cacheAddr})
				}
			}
			numberOfNodeParallelConnections, err := cmd.Flags().GetUint(chainproxy.ParallelConnectionsFlag)
			if err != nil {
				utils.LavaFormatFatal("error fetching chainproxy.ParallelConnectionsFlag", err)
			}
			for _, endpoint := range rpcProviderEndpoints {
				utils.LavaFormatDebug("endpoint description", utils.Attribute{Key: "endpoint", Value: endpoint})
			}
			stickinessHeaderName := viper.GetString(StickinessHeaderName)
			if stickinessHeaderName != "" {
				RPCProviderStickinessHeaderName = stickinessHeaderName
			}
			prometheusListenAddr := viper.GetString(metrics.MetricsListenFlagName)
			rewardStoragePath := viper.GetString(rewardserver.RewardServerStorageFlagName)
			rewardTTL := viper.GetDuration(rewardserver.RewardTTLFlagName)
			shardID := viper.GetUint(ShardIDFlagName)
			rewardsSnapshotThreshold := viper.GetUint(rewardserver.RewardsSnapshotThresholdFlagName)
			rewardsSnapshotTimeoutSec := viper.GetUint(rewardserver.RewardsSnapshotTimeoutSecFlagName)
			enableRelaysHealth := viper.GetBool(common.RelaysHealthEnableFlag)
			relaysHealthInterval := viper.GetDuration(common.RelayHealthIntervalFlag)
			healthCheckURLPath := viper.GetString(HealthCheckURLPathFlagName)

			rpcProviderHealthCheckMetricsOptions := rpcProviderHealthCheckMetricsOptions{
				enableRelaysHealth,
				relaysHealthInterval,
				healthCheckURLPath,
			}

			rpcProviderStartOptions := rpcProviderStartOptions{
				ctx,
				txFactory,
				clientCtx,
				rpcProviderEndpoints,
				cache,
				numberOfNodeParallelConnections,
				prometheusListenAddr,
				rewardStoragePath,
				rewardTTL,
				shardID,
				rewardsSnapshotThreshold,
				rewardsSnapshotTimeoutSec,
				&rpcProviderHealthCheckMetricsOptions,
			}

			rpcProvider := RPCProvider{}
			err = rpcProvider.Start(&rpcProviderStartOptions)
			return err
		},
	}

	// RPCProvider command flags
	flags.AddTxFlagsToCmd(cmdRPCProvider)
	cmdRPCProvider.MarkFlagRequired(flags.FlagFrom)
	cmdRPCProvider.Flags().Bool(common.SaveConfigFlagName, false, "save cmd args to a config file")
	cmdRPCProvider.Flags().Uint64(common.GeolocationFlag, 0, "geolocation to run from")
	cmdRPCProvider.MarkFlagRequired(common.GeolocationFlag)
	cmdRPCProvider.Flags().String(performance.PprofAddressFlagName, "", "pprof server address, used for code profiling")
	cmdRPCProvider.Flags().String(performance.CacheFlagName, "", "address for a cache server to improve performance")
	cmdRPCProvider.Flags().Uint(chainproxy.ParallelConnectionsFlag, chainproxy.NumberOfParallelConnections, "parallel connections")
	cmdRPCProvider.Flags().String(flags.FlagLogLevel, "debug", "log level")
	cmdRPCProvider.Flags().String(metrics.MetricsListenFlagName, metrics.DisabledFlagOption, "the address to expose prometheus metrics (such as localhost:7779)")
	cmdRPCProvider.Flags().String(rewardserver.RewardServerStorageFlagName, rewardserver.DefaultRewardServerStorage, "the path to store reward server data")
	cmdRPCProvider.Flags().Duration(rewardserver.RewardTTLFlagName, rewardserver.DefaultRewardTTL, "reward time to live")
	cmdRPCProvider.Flags().Uint(ShardIDFlagName, DefaultShardID, "shard id")
	cmdRPCProvider.Flags().Uint(rewardserver.RewardsSnapshotThresholdFlagName, rewardserver.DefaultRewardsSnapshotThreshold, "the number of rewards to wait until making snapshot of the rewards memory")
	cmdRPCProvider.Flags().Uint(rewardserver.RewardsSnapshotTimeoutSecFlagName, rewardserver.DefaultRewardsSnapshotTimeoutSec, "the seconds to wait until making snapshot of the rewards memory")
	cmdRPCProvider.Flags().String(StickinessHeaderName, RPCProviderStickinessHeaderName, "the name of the header to be attacked to requests for stickiness by consumer, used for consistency")
	cmdRPCProvider.Flags().Uint64Var(&chaintracker.PollingMultiplier, chaintracker.PollingMultiplierFlagName, 1, "when set, forces the chain tracker to poll more often, improving the sync at the cost of more queries")
	cmdRPCProvider.Flags().DurationVar(&SpecValidationInterval, SpecValidationIntervalFlagName, SpecValidationInterval, "determines the interval of which to run validation on the spec for all connected chains")
	cmdRPCProvider.Flags().DurationVar(&SpecValidationIntervalDisabledChains, SpecValidationIntervalDisabledChainsFlagName, SpecValidationIntervalDisabledChains, "determines the interval of which to run validation on the spec for all disabled chains, determines recovery time")
	cmdRPCProvider.Flags().Bool(common.RelaysHealthEnableFlag, true, "enables relays health check")
	cmdRPCProvider.Flags().Duration(common.RelayHealthIntervalFlag, RelayHealthIntervalFlagDefault, "interval between relay health checks")
	cmdRPCProvider.Flags().String(HealthCheckURLPathFlagName, HealthCheckURLPathFlagDefault, "the url path for the provider's grpc health check")
	cmdRPCProvider.Flags().DurationVar(&updaters.TimeOutForFetchingLavaBlocks, common.TimeOutForFetchingLavaBlocksFlag, time.Second*5, "setting the timeout for fetching lava blocks")

	common.AddRollingLogConfig(cmdRPCProvider)
	return cmdRPCProvider
}
