package main

import (
	"context"
	"fmt"

	"github.com/slok/kahoy/internal/kubernetes"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/plan"
	"github.com/slok/kahoy/internal/resource"
	resourcekubectl "github.com/slok/kahoy/internal/resource/kubectl"
	"github.com/slok/kahoy/internal/storage"
	storagefs "github.com/slok/kahoy/internal/storage/fs"
)

// RunApply runs the apply command.
func RunApply(ctx context.Context, cmdConfig CmdConfig, globalConfig GlobalConfig) error {
	logger := globalConfig.Logger.WithValues(log.Kv{"cmd": "apply"})
	logger.Infof("running command")

	// Create dependencies.
	kubernetesSerializer := kubernetes.NewYAMLObjectSerializer(logger)

	currentRepoStorage, err := storagefs.NewRepository(storagefs.RepositoryConfig{
		ExcludeRegex:      cmdConfig.Apply.ExcludeManifests,
		IncludeRegex:      cmdConfig.Apply.IncludeManifests,
		Path:              cmdConfig.Apply.ManifestsPathOld,
		KubernetesDecoder: kubernetesSerializer,
		Logger:            logger,
	})
	if err != nil {
		return fmt.Errorf("could not create fs %q repository storage: %w", cmdConfig.Apply.ManifestsPathOld, err)
	}

	expectedRepoStorage, err := storagefs.NewRepository(storagefs.RepositoryConfig{
		ExcludeRegex:      cmdConfig.Apply.ExcludeManifests,
		IncludeRegex:      cmdConfig.Apply.IncludeManifests,
		Path:              cmdConfig.Apply.ManifestsPathNew,
		KubernetesDecoder: kubernetesSerializer,
		Logger:            logger,
	})
	if err != nil {
		return fmt.Errorf("could not create fs %q repository storage: %w", cmdConfig.Apply.ManifestsPathNew, err)
	}

	// Prepare for plan.
	currentRes, err := currentRepoStorage.ListResources(ctx, storage.ResourceListOpts{})
	if err != nil {
		return fmt.Errorf("could not retrieve the list of current resources: %w", err)
	}

	expectedRes, err := expectedRepoStorage.ListResources(ctx, storage.ResourceListOpts{})
	if err != nil {
		return fmt.Errorf("could not retrieve the list of expected resources: %w", err)
	}

	// Plan our actions/states.
	planner := plan.NewPlanner(logger)
	statePlan, err := planner.Plan(ctx, expectedRes.Items, currentRes.Items)
	if err != nil {
		return fmt.Errorf("could not get a plan: %w", err)
	}

	applyRes, deleteRes, err := splitPlan(statePlan)
	if err != nil {
		return err
	}

	// Execute them with the correct manager.
	var manager resource.Manager = resource.NewNoopManager(logger)
	switch {
	case cmdConfig.Apply.DryRun:
		manager = resource.NewDryRunManager(cmdConfig.Global.NoColor, nil)
	case cmdConfig.Apply.DiffMode:
		manager, err = resourcekubectl.NewDiffManager(resourcekubectl.DiffManagerConfig{
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			YAMLEncoder: kubernetesSerializer,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create diff resource manager: %w", err)
		}
	default:
		manager, err = resourcekubectl.NewManager(resourcekubectl.ManagerConfig{
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			YAMLEncoder: kubernetesSerializer,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create resource manager: %w", err)
		}
	}

	err = manager.Apply(ctx, applyRes)
	if err != nil {
		return fmt.Errorf("could not apply resources correctly: %w", err)
	}

	err = manager.Delete(ctx, deleteRes)
	if err != nil {
		return fmt.Errorf("could not delete resources correctly: %w", err)
	}

	return nil
}

func splitPlan(statePlan []plan.State) (apply, delete []model.Resource, err error) {
	applyRes := []model.Resource{}
	deleteRes := []model.Resource{}
	for _, s := range statePlan {
		switch s.State {
		case plan.ResourceStateExists:
			applyRes = append(applyRes, s.Resource)
		case plan.ResourceStateMissing:
			deleteRes = append(deleteRes, s.Resource)
		default:
			return nil, nil, fmt.Errorf("unknown resource state on plan: %s-%s", s.Resource.GroupID, s.Resource.ID)
		}
	}

	return applyRes, deleteRes, nil
}
