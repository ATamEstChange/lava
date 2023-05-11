package spec

import (
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramkeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/lavanet/lava/utils"
	epochstoragetypes "github.com/lavanet/lava/x/epochstorage/types"
	"github.com/lavanet/lava/x/spec/keeper"
	"github.com/lavanet/lava/x/spec/types"
)

// overwriting the params handler so we can add events and callbacks on specific params
// NewParamChangeProposalHandler creates a new governance Handler for a ParamChangeProposal
func NewParamChangeProposalHandler(k paramkeeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *paramproposal.ParameterChangeProposal:
			return HandleParameterChangeProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized param proposal content type: %T", c)
		}
	}
}

func HandleParameterChangeProposal(ctx sdk.Context, k paramkeeper.Keeper, p *paramproposal.ParameterChangeProposal) error {
	for _, c := range p.Changes {
		ss, ok := k.GetSubspace(c.Subspace)
		if !ok {
			return sdkerrors.Wrap(paramproposal.ErrUnknownSubspace, c.Subspace)
		}

		logger := k.Logger(ctx)
		details := []utils.Attribute{
			{Key: "param", Value: c.Key},
			{Key: "value", Value: c.Value},
		}
		if c.Key == string(epochstoragetypes.KeyLatestParamChange) {
			return utils.LavaFormatError("Gov Proposal Param Change Error", fmt.Errorf("tried to modify "+string(epochstoragetypes.KeyLatestParamChange)),
				details...,
			)
		}
		if err := ss.Update(ctx, []byte(c.Key), []byte(c.Value)); err != nil {
			return utils.LavaFormatError("Gov Proposal Param Change Error", fmt.Errorf("tried to modify "+string(epochstoragetypes.KeyLatestParamChange)),
				details...,
			)
		}

		details = append(details, utils.Attribute{Key: epochstoragetypes.ModuleName, Value: ctx.BlockHeight()})

		detailsMap := map[string]string{}
		for _, atr := range details {
			detailsMap[atr.Key] = fmt.Sprint(atr.Value)
		}
		utils.LogLavaEvent(ctx, logger, types.ParamChangeEventName, detailsMap, "Gov Proposal Accepted Param Changed")
	}

	ss, ok := k.GetSubspace(epochstoragetypes.ModuleName)
	if !ok {
		return sdkerrors.Wrap(paramproposal.ErrUnknownSubspace, epochstoragetypes.ModuleName)
	}
	ss.Set(ctx, epochstoragetypes.KeyLatestParamChange, uint64(ctx.BlockHeight())) // set the LatestParamChange

	return nil
}

// NewSpecProposalsHandler creates a new governance Handler for a Spec
func NewSpecProposalsHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SpecAddProposal:
			return handleSpecProposal(ctx, k, c)

		default:
			log.Println("unrecognized spec proposal content")
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized spec proposal content type: %T", c)
		}
	}
}

func handleSpecProposal(ctx sdk.Context, k keeper.Keeper, p *types.SpecAddProposal) error {
	logger := k.Logger(ctx)

	for _, spec := range p.Specs {
		_, found := k.GetSpec(ctx, spec.Index)

		details, err := k.ValidateSpec(ctx, spec)
		detailsList := []utils.Attribute{}
		for key, val := range details {
			detailsList = append(detailsList, utils.Attribute{Key: key, Value: val})
		}
		if err != nil {
			return utils.LavaFormatWarning("invalid spec", err,
				detailsList...,
			)
		}

		spec.BlockLastUpdated = uint64(ctx.BlockHeight())
		k.SetSpec(ctx, spec)

		name := types.SpecAddEventName

		if found {
			name = types.SpecModifyEventName
		}

		utils.LogLavaEvent(ctx, logger, name, details, "Gov Proposal Accepted Spec")
		// TODO: add api types once its implemented to the event
	}

	// re-validate all the specs, in case the modified spec is imported by
	// other specs and the new version creates a conflict.
	for _, spec := range k.GetAllSpec(ctx) {
		if details, err := k.ValidateSpec(ctx, spec); err != nil {
			details["invalidates"] = spec.Index
			detailsList := []utils.Attribute{}
			for key, val := range details {
				detailsList = append(detailsList, utils.Attribute{Key: key, Value: val})
			}
			return utils.LavaFormatError("invalidated_spec", err,
				detailsList...,
			)
		}
	}

	return nil
}
