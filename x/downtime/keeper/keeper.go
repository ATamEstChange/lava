package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogowellknown "github.com/gogo/protobuf/types"
	"github.com/lavanet/lava/x/downtime/types"
	v1 "github.com/lavanet/lava/x/downtime/v1"
)

func NewKeeper(cdc codec.BinaryCodec, sk sdk.StoreKey, ps paramtypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(v1.ParamKeyTable())
	}
	return Keeper{
		storeKey:   sk,
		cdc:        cdc,
		paramstore: ps,
	}
}

type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace
}

// ------ STATE -------

func (k Keeper) GetParams(ctx sdk.Context) (params v1.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return
}

func (k Keeper) SetParams(ctx sdk.Context, params v1.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) (*v1.GenesisState, error) {
	return new(v1.GenesisState), nil
}

func (k Keeper) ImportGenesis(ctx sdk.Context, gs *v1.GenesisState) error {
	return nil
}

func (k Keeper) GetLastBlockTime(ctx sdk.Context) (time.Time, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastBlockTimeKey)
	if bz == nil {
		return time.Time{}, false
	}
	protoTime := &gogowellknown.Timestamp{}
	k.cdc.MustUnmarshal(bz, protoTime)
	stdTime, err := gogowellknown.TimestampFromProto(protoTime)
	if err != nil {
		panic(err)
	}
	return stdTime, true
}

func (k Keeper) SetLastBlockTime(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	protoTime, err := gogowellknown.TimestampProto(ctx.BlockTime())
	if err != nil {
		panic(err)
	}
	bz := k.cdc.MustMarshal(protoTime)
	store.Set(types.LastBlockTimeKey, bz)
}

// ------ STATE END -------

func (k Keeper) BeginBlock(ctx sdk.Context) {
}
