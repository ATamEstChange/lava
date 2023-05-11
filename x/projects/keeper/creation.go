package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lavanet/lava/utils"
	plantypes "github.com/lavanet/lava/x/plans/types"
	"github.com/lavanet/lava/x/projects/types"
)

// add a default project to a subscription, add the subscription key as
func (k Keeper) CreateAdminProject(ctx sdk.Context, subscriptionAddress string, plan plantypes.Plan, vrfpk string) error {
	projectData := types.ProjectData{
		Name:        types.ADMIN_PROJECT_NAME,
		Description: types.ADMIN_PROJECT_DESCRIPTION,
		Enabled:     true,
		ProjectKeys: []types.ProjectKey{{Key: subscriptionAddress, Types: []types.ProjectKey_KEY_TYPE{types.ProjectKey_DEVELOPER}, Vrfpk: vrfpk}},
		Policy:      nil,
	}
	return k.CreateProject(ctx, subscriptionAddress, projectData, plan)
}

// add a new project to the subscription
func (k Keeper) CreateProject(ctx sdk.Context, subscriptionAddress string, projectData types.ProjectData, plan plantypes.Plan) error {
	project, err := types.NewProject(subscriptionAddress, projectData.GetName(), projectData.GetDescription(),
		projectData.GetEnabled())
	if err != nil {
		return err
	}

	var emptyProject types.Project
	blockHeight := uint64(ctx.BlockHeight())
	if found := k.projectsFS.FindEntry(ctx, project.Index, blockHeight, &emptyProject); found {
		// the project with the same name already exists if no error has returned
		return utils.LavaFormatWarning("project already exist for the current subscription with the same name", fmt.Errorf("could not create project"),
			utils.Attribute{Key: "subscription", Value: subscriptionAddress},
		)
	}

	project.AdminPolicy = projectData.GetPolicy()

	// projects can be created only by the subscription owner. So the subscription policy is equal to the admin policy
	project.SubscriptionPolicy = project.AdminPolicy

	for _, projectKey := range projectData.GetProjectKeys() {
		err = k.RegisterKey(ctx, types.ProjectKey{Key: projectKey.GetKey(), Types: projectKey.GetTypes(), Vrfpk: projectKey.GetVrfpk()}, &project, blockHeight)
		if err != nil {
			return err
		}
	}

	return k.projectsFS.AppendEntry(ctx, project.Index, blockHeight, &project)
}

func (k Keeper) RegisterKey(ctx sdk.Context, key types.ProjectKey, project *types.Project, blockHeight uint64) error {
	if project == nil {
		return utils.LavaFormatError("project is nil", fmt.Errorf("could not register key to project"))
	}

	for _, keyType := range key.GetTypes() {
		switch keyType {
		case types.ProjectKey_ADMIN:
			k.AddAdminKey(project, key.GetKey(), "")
		case types.ProjectKey_DEVELOPER:
			// try to find the developer key
			var developerData types.ProtoDeveloperData
			found := k.developerKeysFS.FindEntry(ctx, key.GetKey(), blockHeight, &developerData)

			// if we find the developer key and it belongs to a different project, return error
			if found && developerData.ProjectID != project.GetIndex() {
				return utils.LavaFormatWarning("key already exists", fmt.Errorf("could not register key to project"),
					utils.Attribute{Key: "key", Value: key.Key},
					utils.Attribute{Key: "keyTypes", Value: key.Types},
				)
			}

			if !found {
				err := k.AddDeveloperKey(ctx, key.GetKey(), project, blockHeight, key.GetVrfpk())
				if err != nil {
					return utils.LavaFormatError("adding developer key to project failed", err,
						utils.Attribute{Key: "developerKey", Value: key.Key},
						utils.Attribute{Key: "projectIndex", Value: project.Index},
						utils.Attribute{Key: "blockHeight", Value: blockHeight},
					)
				}
			}
		}
	}

	return nil
}

func (k Keeper) AddAdminKey(project *types.Project, adminKey string, vrfpk string) {
	project.AppendKey(types.ProjectKey{Key: adminKey, Types: []types.ProjectKey_KEY_TYPE{types.ProjectKey_ADMIN}, Vrfpk: vrfpk})
}

func (k Keeper) AddDeveloperKey(ctx sdk.Context, developerKey string, project *types.Project, blockHeight uint64, vrfpk string) error {
	var developerData types.ProtoDeveloperData
	developerData.ProjectID = project.GetIndex()
	developerData.Vrfpk = vrfpk
	err := k.developerKeysFS.AppendEntry(ctx, developerKey, blockHeight, &developerData)
	if err != nil {
		return err
	}

	project.AppendKey(types.ProjectKey{Key: developerKey, Types: []types.ProjectKey_KEY_TYPE{types.ProjectKey_DEVELOPER}, Vrfpk: vrfpk})

	return nil
}

// Snapshot all projects of a given subscription
func (k Keeper) SnapshotSubscriptionProjects(ctx sdk.Context, subscriptionAddr string) {
	projects := k.projectsFS.GetAllEntryIndicesWithPrefix(ctx, subscriptionAddr)
	for _, projectID := range projects {
		err := k.snapshotProject(ctx, projectID)
		if err != nil {
			panic(err)
		}
	}
}

// snapshot project, create a snapshot of a project and reset the cu
func (k Keeper) snapshotProject(ctx sdk.Context, projectID string) error {
	var project types.Project
	if found := k.projectsFS.FindEntry(ctx, projectID, uint64(ctx.BlockHeight()), &project); !found {
		return utils.LavaFormatWarning("snapshot of project failed, project does not exist", fmt.Errorf("project not found"),
			utils.Attribute{Key: "projectID", Value: projectID},
		)
	}

	project.UsedCu = 0
	project.Snapshot += 1

	return k.projectsFS.AppendEntry(ctx, project.Index, uint64(ctx.BlockHeight()), &project)
}

func (k Keeper) DeleteProject(ctx sdk.Context, projectID string) error {
	var project types.Project
	if found := k.projectsFS.FindEntry(ctx, projectID, uint64(ctx.BlockHeight()), &project); !found {
		return utils.LavaFormatWarning("project to delete was not found", fmt.Errorf("project not found"),
			utils.Attribute{Key: "projectID", Value: projectID},
		)
	}

	project.Enabled = false
	// TODO: delete all developer keys from the fixation

	return k.projectsFS.AppendEntry(ctx, project.Index, uint64(ctx.BlockHeight()), &project)
}
