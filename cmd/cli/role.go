package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"claude-test/internal/service"
)

func roleCmd(svc *service.RoleService, userSvc *service.UserService, args []string) error {
	if len(args) < 1 {
		printRoleUsage()
		return errors.New("subcommand required")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return roleCreate(svc, subArgs)
	case "list":
		return roleList(svc)
	case "get":
		return roleGet(svc, subArgs)
	case "update":
		return roleUpdate(svc, subArgs)
	case "delete":
		return roleDelete(svc, subArgs)
	case "assign":
		return roleAssign(svc, userSvc, subArgs)
	case "remove":
		return roleRemove(svc, userSvc, subArgs)
	case "user-roles":
		return roleUserRoles(svc, userSvc, subArgs)
	default:
		printRoleUsage()
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func roleCreate(svc *service.RoleService, args []string) error {
	fs := flag.NewFlagSet("role create", flag.ExitOnError)
	name := fs.String("name", "", "role name (required)")
	description := fs.String("description", "", "role description")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return errors.New("--name is required")
	}

	role, err := svc.Create(context.Background(), *name, *description)
	if err != nil {
		return err
	}

	fmt.Printf("Created role: ID=%d, Name=%s, Description=%s\n", role.ID, role.Name, role.Description)
	return nil
}

func roleList(svc *service.RoleService) error {
	roles, err := svc.List(context.Background())
	if err != nil {
		return err
	}

	if len(roles) == 0 {
		fmt.Println("No roles found")
		return nil
	}

	fmt.Printf("%-5s %-20s %-40s\n", "ID", "Name", "Description")
	fmt.Println("----------------------------------------------------------------------")
	for _, r := range roles {
		fmt.Printf("%-5d %-20s %-40s\n", r.ID, r.Name, r.Description)
	}
	return nil
}

func roleGet(svc *service.RoleService, args []string) error {
	fs := flag.NewFlagSet("role get", flag.ExitOnError)
	id := fs.Int64("id", 0, "role ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}

	role, err := svc.GetByID(context.Background(), *id)
	if err != nil {
		return err
	}

	fmt.Printf("ID:          %d\n", role.ID)
	fmt.Printf("Name:        %s\n", role.Name)
	fmt.Printf("Description: %s\n", role.Description)
	return nil
}

func roleUpdate(svc *service.RoleService, args []string) error {
	fs := flag.NewFlagSet("role update", flag.ExitOnError)
	id := fs.Int64("id", 0, "role ID (required)")
	name := fs.String("name", "", "new name (required)")
	description := fs.String("description", "", "new description")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}
	if *name == "" {
		return errors.New("--name is required")
	}

	role, err := svc.Update(context.Background(), *id, *name, *description)
	if err != nil {
		return err
	}

	fmt.Printf("Updated role: ID=%d, Name=%s, Description=%s\n", role.ID, role.Name, role.Description)
	return nil
}

func roleDelete(svc *service.RoleService, args []string) error {
	fs := flag.NewFlagSet("role delete", flag.ExitOnError)
	id := fs.Int64("id", 0, "role ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}

	if err := svc.Delete(context.Background(), *id); err != nil {
		return err
	}

	fmt.Printf("Deleted role with ID=%d\n", *id)
	return nil
}

func roleAssign(svc *service.RoleService, userSvc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("role assign", flag.ExitOnError)
	userID := fs.Int64("user-id", 0, "user ID (required)")
	roleIDs := fs.String("role-ids", "", "comma-separated role IDs (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *userID == 0 {
		return errors.New("--user-id is required")
	}
	if *roleIDs == "" {
		return errors.New("--role-ids is required")
	}

	// Verify user exists
	user, err := userSvc.GetByID(context.Background(), *userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Parse role IDs
	ids, err := parseCommaSeparatedIDs(*roleIDs)
	if err != nil {
		return err
	}

	// Verify roles exist
	for _, id := range ids {
		if _, err := svc.GetByID(context.Background(), id); err != nil {
			return fmt.Errorf("role %d not found: %w", id, err)
		}
	}

	if err := svc.SetUserRoles(context.Background(), *userID, ids); err != nil {
		return err
	}

	fmt.Printf("Assigned roles %v to user %s (ID=%d)\n", ids, user.Email, user.ID)
	return nil
}

func roleRemove(svc *service.RoleService, userSvc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("role remove", flag.ExitOnError)
	userID := fs.Int64("user-id", 0, "user ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *userID == 0 {
		return errors.New("--user-id is required")
	}

	// Verify user exists
	user, err := userSvc.GetByID(context.Background(), *userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Clear all roles
	if err := svc.SetUserRoles(context.Background(), *userID, []int64{}); err != nil {
		return err
	}

	fmt.Printf("Removed all roles from user %s (ID=%d)\n", user.Email, user.ID)
	return nil
}

func roleUserRoles(svc *service.RoleService, userSvc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("role user-roles", flag.ExitOnError)
	userID := fs.Int64("user-id", 0, "user ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *userID == 0 {
		return errors.New("--user-id is required")
	}

	// Verify user exists
	user, err := userSvc.GetByID(context.Background(), *userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	roles, err := svc.GetUserRoles(context.Background(), *userID)
	if err != nil {
		return err
	}

	fmt.Printf("Roles for user %s (ID=%d):\n", user.Email, user.ID)
	if len(roles) == 0 {
		fmt.Println("  No roles assigned")
		return nil
	}

	for _, r := range roles {
		fmt.Printf("  - %s (ID=%d)\n", r.Name, r.ID)
	}
	return nil
}

func parseCommaSeparatedIDs(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	var ids []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", p)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func printRoleUsage() {
	fmt.Println("Usage: cli role <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  create      Create a new role")
	fmt.Println("  list        List all roles")
	fmt.Println("  get         Get role by ID")
	fmt.Println("  update      Update a role")
	fmt.Println("  delete      Delete a role")
	fmt.Println("  assign      Assign roles to a user")
	fmt.Println("  remove      Remove all roles from a user")
	fmt.Println("  user-roles  List roles for a user")
	fmt.Println()
	fmt.Println("Options for create:")
	fmt.Println("  --name string         Role name (required)")
	fmt.Println("  --description string  Role description")
	fmt.Println()
	fmt.Println("Options for get/delete:")
	fmt.Println("  --id int              Role ID (required)")
	fmt.Println()
	fmt.Println("Options for update:")
	fmt.Println("  --id int              Role ID (required)")
	fmt.Println("  --name string         New name (required)")
	fmt.Println("  --description string  New description")
	fmt.Println()
	fmt.Println("Options for assign:")
	fmt.Println("  --user-id int         User ID (required)")
	fmt.Println("  --role-ids string     Comma-separated role IDs (required)")
	fmt.Println()
	fmt.Println("Options for remove/user-roles:")
	fmt.Println("  --user-id int         User ID (required)")
}
