package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"

	"github.com/casbin/casbin/v3"

	"claude-test/internal/authz"
)

func policyCmd(db *sql.DB, args []string) error {
	if len(args) < 1 {
		printPolicyUsage()
		return errors.New("subcommand required")
	}

	enforcer, err := authz.NewEnforcer(db)
	if err != nil {
		return fmt.Errorf("failed to create enforcer: %w", err)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list":
		return policyList(enforcer)
	case "add":
		return policyAdd(enforcer, subArgs)
	case "remove":
		return policyRemove(enforcer, subArgs)
	case "check":
		return policyCheck(enforcer, subArgs)
	case "init":
		return policyInit(enforcer)
	default:
		printPolicyUsage()
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func policyList(e *casbin.Enforcer) error {
	policies, _ := e.GetPolicy()

	if len(policies) == 0 {
		fmt.Println("No policies found")
		return nil
	}

	fmt.Printf("%-15s %-30s %-10s\n", "Role", "Resource", "Action")
	fmt.Println("----------------------------------------------------------")
	for _, p := range policies {
		if len(p) >= 3 {
			fmt.Printf("%-15s %-30s %-10s\n", p[0], p[1], p[2])
		}
	}
	return nil
}

func policyAdd(e *casbin.Enforcer, args []string) error {
	fs := flag.NewFlagSet("policy add", flag.ExitOnError)
	role := fs.String("role", "", "role name (required)")
	resource := fs.String("resource", "", "resource path (required)")
	action := fs.String("action", "", "action: read or write (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *role == "" || *resource == "" || *action == "" {
		return errors.New("--role, --resource, and --action are required")
	}

	if *action != "read" && *action != "write" {
		return errors.New("--action must be 'read' or 'write'")
	}

	added, err := e.AddPolicy(*role, *resource, *action)
	if err != nil {
		return err
	}

	if !added {
		fmt.Println("Policy already exists")
		return nil
	}

	if err := e.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policy: %w", err)
	}

	fmt.Printf("Added policy: %s can %s %s\n", *role, *action, *resource)
	return nil
}

func policyRemove(e *casbin.Enforcer, args []string) error {
	fs := flag.NewFlagSet("policy remove", flag.ExitOnError)
	role := fs.String("role", "", "role name (required)")
	resource := fs.String("resource", "", "resource path (required)")
	action := fs.String("action", "", "action: read or write (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *role == "" || *resource == "" || *action == "" {
		return errors.New("--role, --resource, and --action are required")
	}

	removed, err := e.RemovePolicy(*role, *resource, *action)
	if err != nil {
		return err
	}

	if !removed {
		fmt.Println("Policy not found")
		return nil
	}

	if err := e.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policy: %w", err)
	}

	fmt.Printf("Removed policy: %s can %s %s\n", *role, *action, *resource)
	return nil
}

func policyCheck(e *casbin.Enforcer, args []string) error {
	fs := flag.NewFlagSet("policy check", flag.ExitOnError)
	role := fs.String("role", "", "role name (required)")
	resource := fs.String("resource", "", "resource path (required)")
	action := fs.String("action", "", "action: read or write (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *role == "" || *resource == "" || *action == "" {
		return errors.New("--role, --resource, and --action are required")
	}

	allowed, err := e.Enforce(*role, *resource, *action)
	if err != nil {
		return err
	}

	if allowed {
		fmt.Printf("ALLOWED: %s can %s %s\n", *role, *action, *resource)
	} else {
		fmt.Printf("DENIED: %s cannot %s %s\n", *role, *action, *resource)
	}
	return nil
}

func policyInit(e *casbin.Enforcer) error {
	if err := authz.SetupDefaultPolicies(e); err != nil {
		return err
	}
	fmt.Println("Default policies initialized")
	return nil
}

func printPolicyUsage() {
	fmt.Println("Usage: cli policy <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list     List all policies")
	fmt.Println("  add      Add a policy")
	fmt.Println("  remove   Remove a policy")
	fmt.Println("  check    Check if a role has permission")
	fmt.Println("  init     Initialize default policies")
	fmt.Println()
	fmt.Println("Options for add/remove/check:")
	fmt.Println("  --role string      Role name (required)")
	fmt.Println("  --resource string  Resource path, e.g., /web/users (required)")
	fmt.Println("  --action string    Action: read or write (required)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli policy list")
	fmt.Println("  cli policy add --role Editor --resource /web/users --action read")
	fmt.Println("  cli policy remove --role Editor --resource /web/users --action read")
	fmt.Println("  cli policy check --role Admin --resource /web/users --action write")
	fmt.Println("  cli policy init")
}
