package main

import (
	"context"
	"fmt"

	"github.com/slok/kahoy/internal/kubernetes"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/plan"
	resourcemanage "github.com/slok/kahoy/internal/resource/manage"
	managekubectl "github.com/slok/kahoy/internal/resource/manage/kubectl"
	resourceprocess "github.com/slok/kahoy/internal/resource/process"
	"github.com/slok/kahoy/internal/storage"
	storagefs "github.com/slok/kahoy/internal/storage/fs"
	storagegit "github.com/slok/kahoy/internal/storage/git"
)

// RunApply runs the apply command.
func RunApply(ctx context.Context, cmdConfig CmdConfig, globalConfig GlobalConfig) error {
	logger := globalConfig.Logger.WithValues(log.Kv{
		"cmd":  "apply",
		"mode": cmdConfig.Apply.Mode,
	})
	logger.Infof("running command")

	// Create YAML serializer.
	kubernetesSerializer := kubernetes.NewYAMLObjectSerializer(logger)

	var (
		oldResourceRepo, newResourceRepo storage.ResourceRepository
	)
	switch cmdConfig.Apply.Mode {
	case ApplyModeGit:
		oldRepo, newRepo, err := storagegit.NewRepositories(storagegit.RepositoriesConfig{
			ExcludeRegex:         cmdConfig.Apply.ExcludeManifests,
			IncludeRegex:         cmdConfig.Apply.IncludeManifests,
			OldRelPath:           cmdConfig.Apply.ManifestsPathOld,
			NewRelPath:           cmdConfig.Apply.ManifestsPathNew,
			GitBeforeCommitSHA:   cmdConfig.Apply.GitBeforeCommit,
			GitDefaultBranch:     cmdConfig.Apply.GitDefaultBranch,
			GitDiffIncludeFilter: cmdConfig.Apply.GitDiffFilter,
			KubernetesDecoder:    kubernetesSerializer,
			Logger:               logger,
		})
		if err != nil {
			return fmt.Errorf("could not create git based fs repos storage: %w", err)
		}

		oldResourceRepo = oldRepo
		newResourceRepo = newRepo

	case ApplyModePaths:
		oldRepo, newRepo, err := storagefs.NewRepositories(storagefs.RepositoriesConfig{
			ExcludeRegex:      cmdConfig.Apply.ExcludeManifests,
			IncludeRegex:      cmdConfig.Apply.IncludeManifests,
			OldPath:           cmdConfig.Apply.ManifestsPathOld,
			NewPath:           cmdConfig.Apply.ManifestsPathNew,
			KubernetesDecoder: kubernetesSerializer,
			Logger:            logger,
		})
		if err != nil {
			return fmt.Errorf("could not create fs repos storage: %w", err)
		}

		oldResourceRepo = oldRepo
		newResourceRepo = newRepo
	default:
		return fmt.Errorf("unknown apply mode: %s", cmdConfig.Apply.Mode)
	}

	// Get resources from repositories.
	currentRes, err := oldResourceRepo.ListResources(ctx, storage.ResourceListOpts{})
	if err != nil {
		return fmt.Errorf("could not retrieve the list of current resources: %w", err)
	}

	expectedRes, err := newResourceRepo.ListResources(ctx, storage.ResourceListOpts{})
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

	// Process planned resources.
	exclKubeTypeProc, err := resourceprocess.NewExcludeKubeTypeProcessor(cmdConfig.Apply.ExcludeKubeTypeResources, logger)
	if err != nil {
		return fmt.Errorf("could not create Kubernetes resorce type exclude processor: %w", err)
	}
	resProc := resourceprocess.NewResourceProcessorChain(exclKubeTypeProc)

	applyRes, err = resProc.Process(ctx, applyRes)
	if err != nil {
		return fmt.Errorf("error while processing apply state resources: %w", err)
	}

	deleteRes, err = resProc.Process(ctx, deleteRes)
	if err != nil {
		return fmt.Errorf("error while processing delete state resources: %w", err)
	}

	if len(applyRes)+len(deleteRes) <= 0 {
		logger.Infof("no resources to apply/delete, exiting...")
		return nil
	}

	// Execute them with the correct manager.
	var manager resourcemanage.ResourceManager = resourcemanage.NewNoopManager(logger)
	switch {
	case cmdConfig.Apply.DryRun:
		manager = resourcemanage.NewDryRunManager(cmdConfig.Global.NoColor, nil)
	case cmdConfig.Apply.DiffMode:
		manager, err = managekubectl.NewDiffManager(managekubectl.DiffManagerConfig{
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			YAMLEncoder: kubernetesSerializer,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create diff resource manager: %w", err)
		}
	default:
		manager, err = managekubectl.NewManager(managekubectl.ManagerConfig{
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
