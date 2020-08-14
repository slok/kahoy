package kubectl

import "fmt"

// newKubectlCmdArgs will return an slice of args based on an array
// of kubectl cmd options.
func newKubectlCmdArgs(opts []kubectlCmdOption) []string {
	args := []string{}
	for _, opt := range opts {
		args = opt(args)
	}

	return args
}

type kubectlCmdOption func([]string) []string

func withApplyCmd() kubectlCmdOption {
	return func(args []string) []string {
		return append(args, "apply")
	}
}

func withDeleteCmd() kubectlCmdOption {
	return func(args []string) []string {
		return append(args, "delete")
	}
}

func withDiffCmd() kubectlCmdOption {
	return func(args []string) []string {
		return append(args, "diff")
	}
}

func withContext(context string) kubectlCmdOption {
	return func(args []string) []string {
		if context == "" {
			return args
		}
		return append(args, "--context", context)
	}
}

func withConfig(path string) kubectlCmdOption {
	return func(args []string) []string {
		if path == "" {
			return args
		}
		return append(args, "--kubeconfig", path)
	}
}

func withForceConflicts(force bool) kubectlCmdOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("--force-conflicts=%t", force))
	}
}

func withFieldManager(manager string) kubectlCmdOption {
	return func(args []string) []string {
		if manager == "" {
			return args
		}
		return append(args, "--field-manager", manager)
	}
}

func withServerSide(serverSide bool) kubectlCmdOption {
	return func(args []string) []string {
		return append(args, fmt.Sprintf("--server-side=%t", serverSide))
	}
}

func withStdIn() kubectlCmdOption {
	return func(args []string) []string {
		return append(args, "-f", "-")
	}
}
