package ssh

import (
	"context"
	"fmt"
	"strings"

	"github.com/unweave/unweave/api/types"
)

func Connect(ctx context.Context, connectionInfo types.ExecNetwork, prvKeyPath string, args []string, command []string) error {
	proxiedssh, err := NewProxied(
		connectionInfo.User,
		"localhost:2233",
		fmt.Sprintf("%s:%d", connectionInfo.Host, 50505),
		"localhost:22",
		prvKeyPath,
	)
	if err != nil {
		return fmt.Errorf("create proxied ssh: %w", err)
	}

	overrideUserKnownHostsFile := false
	overrideStrictHostKeyChecking := false

	for _, arg := range args {
		if strings.Contains(arg, "UserKnownHostsFile") {
			overrideUserKnownHostsFile = true
		}
		if strings.Contains(arg, "StrictHostKeyChecking") {
			overrideStrictHostKeyChecking = true
		}
	}

	if prvKeyPath != "" {
		args = append(args, "-i", prvKeyPath)
	}

	if !overrideUserKnownHostsFile {
		args = append(args, "-o", "UserKnownHostsFile=/dev/null")
	}
	if !overrideStrictHostKeyChecking {
		args = append(args, "-o", "StrictHostKeyChecking=no")
	}

	if err := proxiedssh.RunCommand(args, command); err != nil {
		return fmt.Errorf("run proxied ssh command: %w", err)
	}

	return nil
}
