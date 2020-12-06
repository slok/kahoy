package hook

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"

	"github.com/slok/kahoy/internal/log"
	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
	"github.com/slok/kahoy/internal/storage"
)

// CmdRunner knows how to run exec.Cmd commands. This can be used to
// hijack the command execution.
type CmdRunner interface {
	Start(*exec.Cmd) error
	Wait(*exec.Cmd) error
	CombinedOutputPipe(*exec.Cmd) (io.Reader, error)
}

//go:generate mockery --case underscore --output hookmock --outpkg hookmock --name CmdRunner

type stdCmdRunner int

const defaultCmdRunner = stdCmdRunner(0)

func (stdCmdRunner) Start(c *exec.Cmd) error { return c.Start() }
func (stdCmdRunner) Wait(c *exec.Cmd) error  { return c.Wait() }
func (stdCmdRunner) CombinedOutputPipe(cmd *exec.Cmd) (io.Reader, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	return io.MultiReader(stdoutPipe, stderrPipe), nil
}

// ManagerConfig is the configuration of the Wait manager.
type ManagerConfig struct {
	Manager         manage.ResourceManager
	GroupRepository storage.GroupRepository
	CmdRunner       CmdRunner
	KubectlCmd      string
	KubeConfig      string
	KubeContext     string
	Logger          log.Logger
}

func (c *ManagerConfig) defaults() error {
	if c.Manager == nil {
		return fmt.Errorf("manager is required")
	}

	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"app-svc": "hook.Manager"})

	if c.GroupRepository == nil {
		return fmt.Errorf("group repository is required")
	}

	if c.CmdRunner == nil {
		c.CmdRunner = defaultCmdRunner
	}

	if c.KubectlCmd == "" {
		c.KubectlCmd = "kubectl"
	}

	return nil
}

// hookManager knows how to execute cmds before and after resources management.
type hookManager struct {
	groupRepo   storage.GroupRepository
	manager     manage.ResourceManager
	cmdRunner   CmdRunner
	kubectlCmd  string
	kubeConfig  string
	kubeContext string
	logger      log.Logger
}

// NewManager returns a manager that knows how to execute hooks `Apply`
func NewManager(config ManagerConfig) (manage.ResourceManager, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("could not create wait manager: %w", err)
	}
	return hookManager{
		groupRepo:   config.GroupRepository,
		manager:     config.Manager,
		cmdRunner:   config.CmdRunner,
		kubectlCmd:  config.KubectlCmd,
		kubeConfig:  config.KubeConfig,
		kubeContext: config.KubeContext,
		logger:      config.Logger,
	}, err
}

// Apply will apply and then wait using the most important waiting policy of all
// the applied resources.
func (h hookManager) Apply(ctx context.Context, resources []model.Resource) error {
	preHooks, postHooks, err := h.createHooks(ctx, resources)
	if err != nil {
		return fmt.Errorf("could not create group hooks: %w", err)
	}

	// Pre hooks.
	err = h.executeHooks(ctx, preHooks)
	if err != nil {
		return fmt.Errorf("pre hooks error: %w", err)
	}

	// Apply.
	err = h.manager.Apply(ctx, resources)
	if err != nil {
		return err
	}

	// Post hooks.
	err = h.executeHooks(ctx, postHooks)
	if err != nil {
		return fmt.Errorf("post hooks error: %w", err)
	}

	return nil
}

// Delete is NOOP on hooks.
func (h hookManager) Delete(ctx context.Context, resources []model.Resource) error {
	return h.manager.Delete(ctx, resources)
}

// createHooks will create the hooks from the resource groups.
func (h hookManager) createHooks(ctx context.Context, resources []model.Resource) (pre []hook, post []hook, err error) {
	var preHooks, postHooks []hook

	groups := map[string]*model.Group{}
	for _, r := range resources {
		group, err := h.groupRepo.GetGroup(ctx, r.GroupID)
		if err != nil {
			return nil, nil, fmt.Errorf("could not get group %q: %w", r.GroupID, err)
		}
		groups[group.ID] = group
	}

	for _, group := range groups {
		if group.Hooks.Pre != nil {
			preHooks = append(preHooks, h.createHook(group, group.Hooks.Pre, hookPreType))
		}
		if group.Hooks.Post != nil {
			postHooks = append(postHooks, h.createHook(group, group.Hooks.Post, hookPostType))
		}
	}

	return preHooks, postHooks, nil
}

func (h hookManager) executeHooks(ctx context.Context, hooks []hook) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, hook := range hooks {
		hook := hook
		g.Go(func() error {
			return hook(ctx)
		})
	}

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("hooks execution stopped due to failure: %w", err)
	}

	return nil

}

const (
	hookPreType  = "pre"
	hookPostType = "post"
)

func (h hookManager) createHook(group *model.Group, config *model.GroupHookSpec, hookType string) hook {
	if config == nil || len(config.Cmd) == 0 {
		return func(ctx context.Context) error { return nil } // Noop.
	}

	logger := h.logger.WithValues(log.Kv{
		"group":     group.ID,
		"hook-type": hookType,
		"hook-cmd":  config.Cmd[0],
	})

	return func(ctx context.Context) error {
		logger.Infof("executing group hook")

		// Add timeout.
		if config.Timeout > 0 {
			ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
			defer cancel()
			ctx = ctxTimeout
		}

		// Prepare command.
		cmd := exec.CommandContext(ctx, config.Cmd[0], config.Cmd[1:]...)
		out, err := h.cmdRunner.CombinedOutputPipe(cmd)
		if err != nil {
			return err
		}

		// Set correct env for the commands.
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("KAHOY_KUBECTL_CMD=%s", h.kubectlCmd),
			fmt.Sprintf("KAHOY_KUBE_CONFIG=%s", h.kubeConfig),
			fmt.Sprintf("KAHOY_KUBE_CONTEXT=%s", h.kubeContext),
			fmt.Sprintf("KAHOY_HOOK_TYPE=%s", hookType),
			fmt.Sprintf("KAHOY_HOOK_GROUP=%s", group.ID),
		)

		// Execute.
		err = h.cmdRunner.Start(cmd)
		if err != nil {
			return fmt.Errorf("could not start hook: %w", err)
		}

		// Log in background.
		go func() {
			outStream := bufio.NewScanner(out)
			for outStream.Scan() {
				select {
				case <-ctx.Done():
					break
				default:
				}
				logger.Infof(outStream.Text())
			}
		}()

		// Wait until the command is done.
		err = h.cmdRunner.Wait(cmd)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("timeout: %w", err)
			}
			return fmt.Errorf("hook execution failed: %w", err)
		}

		return nil
	}
}

type hook func(ctx context.Context) error
