package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	internalkubernetes "github.com/slok/kahoy/internal/kubernetes"
	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/plan"
	resourcemanage "github.com/slok/kahoy/internal/resource/manage"
	managebatch "github.com/slok/kahoy/internal/resource/manage/batch"
	managedryrun "github.com/slok/kahoy/internal/resource/manage/dryrun"
	managekubectl "github.com/slok/kahoy/internal/resource/manage/kubectl"
	manageTimeout "github.com/slok/kahoy/internal/resource/manage/timeout"
	managewait "github.com/slok/kahoy/internal/resource/manage/wait"
	resourceprocess "github.com/slok/kahoy/internal/resource/process"
	"github.com/slok/kahoy/internal/storage"
	storagefs "github.com/slok/kahoy/internal/storage/fs"
	storagegit "github.com/slok/kahoy/internal/storage/git"
	storagekubernetes "github.com/slok/kahoy/internal/storage/kubernetes"
	storagereport "github.com/slok/kahoy/internal/storage/report"
)

// RunApply runs the apply command.
func RunApply(ctx context.Context, cmdConfig CmdConfig, globalConfig GlobalConfig) error {
	report, err := model.NewState()
	if err != nil {
		return fmt.Errorf("could not start the app report: %w", err)
	}

	logger := globalConfig.Logger.WithValues(log.Kv{
		"id":       report.ID,
		"cmd":      "apply",
		"provider": cmdConfig.Apply.Provider,
	})
	logger.Infof("running command")

	// Create YAML serializer.
	kubernetesSerializer := internalkubernetes.NewYAMLObjectSerializer(logger)

	// Aggregate options (cmd flags + kahoy config files).
	fsExclude := append(cmdConfig.Apply.ExcludeManifests, globalConfig.AppConfig.Fs.Exclude...)
	fsInclude := append(cmdConfig.Apply.IncludeManifests, globalConfig.AppConfig.Fs.Include...)

	// Create Kubernetes client.
	kubeCfg, err := loadKubernetesConfig(cmdConfig)
	if err != nil {
		return fmt.Errorf("could not load Kubernetes configuration: %w", err)
	}
	kubeRawCli, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("could not create client-go kubernetes client: %w", err)
	}
	kubeCli := internalkubernetes.NewClient(kubeRawCli, logger)

	modelResGroupFactory, err := model.NewResourceAndGroupFactory(kubeCli, logger)
	if err != nil {
		return fmt.Errorf("could not create resource and group models factory: %w", err)
	}

	var (
		oldResourceRepo, newResourceRepo storage.ResourceRepository
		newGroupRepo                     storage.GroupRepository
		stateRepo                        storage.StateRepository = storage.NewNoopStateRepository(logger)
	)
	switch cmdConfig.Apply.Provider {
	case ApplyProviderGit:
		oldRepo, newRepo, err := storagegit.NewRepositories(storagegit.RepositoriesConfig{
			ExcludeRegex:       fsExclude,
			IncludeRegex:       fsInclude,
			OldRelPath:         cmdConfig.Apply.ManifestsPathOld,
			NewRelPath:         cmdConfig.Apply.ManifestsPathNew,
			GitBeforeCommitSHA: cmdConfig.Apply.GitBeforeCommit,
			GitDefaultBranch:   cmdConfig.Apply.GitDefaultBranch,
			KubernetesDecoder:  kubernetesSerializer,
			AppConfig:          &globalConfig.AppConfig,
			Logger:             logger,
			ModelFactory:       modelResGroupFactory,
		})
		if err != nil {
			return fmt.Errorf("could not create git based fs repos storage: %w", err)
		}

		oldResourceRepo = oldRepo
		newResourceRepo = newRepo
		newGroupRepo = newRepo

	case ApplyProviderPaths:
		oldRepo, newRepo, err := storagefs.NewRepositories(storagefs.RepositoriesConfig{
			Ctx:               ctx,
			StdIn:             globalConfig.Stdin,
			ExcludeRegex:      fsExclude,
			IncludeRegex:      fsInclude,
			OldPath:           cmdConfig.Apply.ManifestsPathOld,
			NewPath:           cmdConfig.Apply.ManifestsPathNew,
			KubernetesDecoder: kubernetesSerializer,
			AppConfig:         &globalConfig.AppConfig,
			ModelFactory:      modelResGroupFactory,
			Logger:            logger,
		})
		if err != nil {
			return fmt.Errorf("could not create fs repos storage: %w", err)
		}

		oldResourceRepo = oldRepo
		newResourceRepo = newRepo
		newGroupRepo = newRepo

	case ApplyProviderK8s:
		k8sRepo, err := storagekubernetes.NewRepository(storagekubernetes.RepositoryConfig{
			Namespace:    cmdConfig.Apply.KubeProviderNs,
			StorageID:    cmdConfig.Apply.KubeProviderID,
			Serializer:   kubernetesSerializer,
			Client:       kubeCli,
			ModelFactory: modelResGroupFactory,
			Logger:       logger.WithValues(log.Kv{"repo-state": "old"}),
		})
		if err != nil {
			return fmt.Errorf("could not create state storer: %w", err)
		}

		_, newRepo, err := storagefs.NewRepositories(storagefs.RepositoriesConfig{
			Ctx:               ctx,
			StdIn:             globalConfig.Stdin,
			ExcludeRegex:      fsExclude,
			IncludeRegex:      fsInclude,
			OldPath:           os.DevNull,
			NewPath:           cmdConfig.Apply.ManifestsPathNew,
			KubernetesDecoder: kubernetesSerializer,
			AppConfig:         &globalConfig.AppConfig,
			ModelFactory:      modelResGroupFactory,
			Logger:            logger.WithValues(log.Kv{"repo-state": "new"}),
		})
		if err != nil {
			return fmt.Errorf("could not create fs repos storage: %w", err)
		}

		// State store and old repository is from Kubernetes.
		stateRepo = k8sRepo
		oldResourceRepo = k8sRepo
		newResourceRepo = newRepo
		newGroupRepo = newRepo

	default:
		return fmt.Errorf("unknown apply provider: %s", cmdConfig.Apply.Provider)
	}

	// Get resources from repositories.
	oldRes, err := oldResourceRepo.ListResources(ctx, storage.ResourceListOpts{})
	if err != nil {
		return fmt.Errorf("could not retrieve the list of current resources: %w", err)
	}

	newRes, err := newResourceRepo.ListResources(ctx, storage.ResourceListOpts{})
	if err != nil {
		return fmt.Errorf("could not retrieve the list of expected resources: %w", err)
	}

	// Plan our actions/states.
	planner := plan.NewPlanner(cmdConfig.Apply.IncludeChanges, logger)
	statePlan, err := planner.Plan(ctx, oldRes.Items, newRes.Items)
	if err != nil {
		return fmt.Errorf("could not get a plan: %w", err)
	}

	applyRes, deleteRes, err := splitPlan(statePlan)
	if err != nil {
		return err
	}

	// Process planned resources.
	resProc, err := newResourceProcessor(cmdConfig, logger)
	if err != nil {
		return err
	}

	resQBefore := len(applyRes)
	applyRes, err = resProc.Process(ctx, applyRes)
	if err != nil {
		return fmt.Errorf("error while processing apply state resources: %w", err)
	}
	resQAfter := len(applyRes)
	logger.Infof("apply resources before filter %d, after %d", resQBefore, resQAfter)

	resQBefore = len(deleteRes)
	deleteRes, err = resProc.Process(ctx, deleteRes)
	if err != nil {
		return fmt.Errorf("error while processing delete state resources: %w", err)
	}

	if len(applyRes)+len(deleteRes) <= 0 {
		logger.Infof("no resources to apply/delete, exiting...")
		return nil
	}
	resQAfter = len(deleteRes)
	logger.Infof("delete resources before filter %d, after %d", resQBefore, resQAfter)

	// Select the execution logic based on diff, dry-run...
	var (
		manager    resourcemanage.ResourceManager
		reportRepo storage.StateRepository = storage.NewNoopStateRepository(logger)
	)
	switch {
	case cmdConfig.Apply.DryRun:
		stateRepo = storage.NewNoopStateRepository(logger)
		manager = managedryrun.NewManager(cmdConfig.Global.NoColor, nil)

	case cmdConfig.Apply.DiffMode:
		stateRepo = storage.NewNoopStateRepository(logger)
		manager, err = managekubectl.NewDiffManager(managekubectl.DiffManagerConfig{
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			KubectlCmd:  cmdConfig.Apply.KubectlPath,
			YAMLEncoder: kubernetesSerializer,
			YAMLDecoder: kubernetesSerializer,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create diff resource manager: %w", err)
		}

	default:
		manager, err = managekubectl.NewManager(managekubectl.ManagerConfig{
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			KubectlCmd:  cmdConfig.Apply.KubectlPath,
			YAMLEncoder: kubernetesSerializer,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create resource manager: %w", err)
		}

		// Wrap the executor manger with wait manager. This is wrapped here because
		// wait manager should only wait on real executions.
		manager, err = managewait.NewManager(managewait.ManagerConfig{
			Manager:         manager,
			GroupRepository: newGroupRepo,
			Logger:          logger,
		})
		if err != nil {
			return fmt.Errorf("could not create wait resource manager: %w", err)
		}

		// Set up report output.
		switch cmdConfig.Apply.ReportPath {
		case "":
			// NOOP.

		// Write output to stdout.
		case "-":
			reportRepo = storagereport.NewJSONStateRepository(globalConfig.Stdout)

		// Anything else write as if it was a path to a file.
		default:
			outFile, err := os.Create(cmdConfig.Apply.ReportPath)
			if err != nil {
				return fmt.Errorf("could not open file %q for out report: %w", cmdConfig.Apply.ReportPath, err)
			}
			logger.Infof("report will be written to %q", cmdConfig.Apply.ReportPath)
			defer outFile.Close()

			reportRepo = storagereport.NewJSONStateRepository(outFile)
		}
	}

	// Wrap resource manager with timeout manager if timeout is properly set
	if cmdConfig.Apply.ExecutionTimeout != 0 {
		manager, err = manageTimeout.NewTimeoutManager(manageTimeout.TimeoutManagerConfig{
			Timeout: cmdConfig.Apply.ExecutionTimeout,
			Manager: manager,
			Logger:  logger,
		})
		if err != nil {
			return fmt.Errorf("could not create timeout manager: %w", err)
		}
	}

	// Wrap manager with batch manager. This should wrap the executors managers
	manager, err = managebatch.NewPriorityManager(managebatch.PriorityManagerConfig{
		Manager:         manager,
		Logger:          logger,
		GroupRepository: newGroupRepo,
	})
	if err != nil {
		return fmt.Errorf("could not create batch manager: %w", err)
	}

	// If we need to create namespaces before apply, we wrap the manager with the
	// ns ensure manager that will ensure the namespace exists.
	if cmdConfig.Apply.CreateNamespace && !cmdConfig.Apply.DryRun {
		manager, err = managekubectl.NewNamespaceEnsurer(managekubectl.NamespaceEnsurerConfig{
			Manager:     manager,
			KubeConfig:  cmdConfig.Apply.KubeConfig,
			KubeContext: cmdConfig.Apply.KubeContext,
			KubectlCmd:  cmdConfig.Apply.KubectlPath,
			Logger:      logger,
		})
		if err != nil {
			return fmt.Errorf("could not create namespace ensurer manager: %w", err)
		}
	}

	// Ask for confirmation
	if !cmdConfig.Apply.DryRun && !cmdConfig.Apply.DiffMode && !cmdConfig.Apply.AutoApprove {
		proceed, err := askYesNo(globalConfig.Stdout, globalConfig.Stdin)
		if err != nil {
			return fmt.Errorf("could not read confirmation: %w", err)
		}

		if !proceed {
			return nil
		}
	}

	// Execute actions on resources.
	err = applyDeleteResources(ctx, manager, applyRes, deleteRes, cmdConfig.Apply.DeleteFirst)
	if err != nil {
		return err
	}

	// Set applied/deleted resources on report.
	report.EndedAt = time.Now().UTC()
	report.AppliedResources = applyRes
	report.DeletedResources = deleteRes

	// Store executed state.
	err = stateRepo.StoreState(ctx, *report)
	if err != nil {
		return fmt.Errorf("could not store state: %w", err)
	}

	// Show report.
	err = reportRepo.StoreState(ctx, *report)
	if err != nil {
		return fmt.Errorf("could not store report: %w", err)
	}

	return nil
}

// splitPlan takes a list of resources from the plan and splits them by state.
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

// newResourceProcessor will create the resource processor using a chain of multiple resource processors that will
// be executed after the resource plan.
func newResourceProcessor(cmdConfig CmdConfig, logger log.Logger) (resourceprocess.ResourceProcessor, error) {
	exclKubeTypeProc, err := resourceprocess.NewExcludeKubeTypeProcessor(cmdConfig.Apply.ExcludeKubeTypeResources, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes resource type exclude processor: %w", err)
	}

	incNamespacesProc, err := resourceprocess.NewIncludeNamespaceProcessor(cmdConfig.Apply.IncludeNamespaces, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes resource namespace include processor: %w", err)
	}

	includeLabelProc, err := resourceprocess.NewKubeLabelSelectorProcessor(cmdConfig.Apply.KubeLabelSelector, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes label selector processor: %w", err)
	}

	includeAnnotationProc, err := resourceprocess.NewKubeAnnotationSelectorProcessor(cmdConfig.Apply.KubeAnnotationSelector, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes annotation selector processor: %w", err)
	}

	resProc := resourceprocess.NewResourceProcessorChain(
		exclKubeTypeProc,
		incNamespacesProc,
		includeLabelProc,
		includeAnnotationProc,
	)

	return resProc, nil
}

// askYesNo prompts the user with a dialog to ask whether wants to proceed
// or not
func askYesNo(writer io.Writer, reader io.Reader) (bool, error) {
	var s string

	fmt.Fprintf(writer, "Do you want to proceed? (y/N): ")
	_, err := fmt.Fscan(reader, &s)
	if err != nil {
		return false, err
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true, nil
	}

	return false, nil
}

// loadKubernetesConfig loads kubernetes configuration based on flags.
func loadKubernetesConfig(cmdCfg CmdConfig) (*rest.Config, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: cmdCfg.Apply.KubeConfig,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: cmdCfg.Apply.KubeContext,
			// TODO(slok): Timeout.
		},
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("could not load Kubernetes configuration: %w", err)
	}

	// Set better cli rate limiter.
	config.QPS = 100
	config.Burst = 100

	return config, nil
}

func applyDeleteResources(ctx context.Context, manager resourcemanage.ResourceManager, applyRes, deleteRes []model.Resource, deleteFirst bool) error {
	// Prepare.
	apply := func() error {
		err := manager.Apply(ctx, applyRes)
		if err != nil {
			return fmt.Errorf("could not apply resources correctly: %w", err)
		}
		return nil
	}
	delete := func() error {
		err := manager.Delete(ctx, deleteRes)
		if err != nil {
			return fmt.Errorf("could not delete resources correctly: %w", err)
		}
		return nil
	}

	// Sort.
	funcs := []func() error{apply, delete}
	if deleteFirst {
		funcs = []func() error{delete, apply}
	}

	// Execute.
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}
