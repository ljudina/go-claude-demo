package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"claude-test/internal/service"
)

func userCmd(svc *service.UserService, args []string) error {
	if len(args) < 1 {
		printUserUsage()
		return errors.New("subcommand required")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return userCreate(svc, subArgs)
	case "list":
		return userList(svc)
	case "get":
		return userGet(svc, subArgs)
	case "update":
		return userUpdate(svc, subArgs)
	case "delete":
		return userDelete(svc, subArgs)
	default:
		printUserUsage()
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func userCreate(svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("user create", flag.ExitOnError)
	email := fs.String("email", "", "user email (required)")
	name := fs.String("name", "", "user name (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *email == "" || *name == "" {
		return errors.New("--email and --name are required")
	}

	user, err := svc.Create(context.Background(), *email, *name)
	if err != nil {
		return err
	}

	fmt.Printf("Created user: ID=%d, Email=%s, Name=%s\n", user.ID, user.Email, user.Name)
	return nil
}

func userList(svc *service.UserService) error {
	users, err := svc.List(context.Background())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return nil
	}

	fmt.Printf("%-5s %-30s %-30s %-10s\n", "ID", "Email", "Name", "Theme")
	fmt.Println("----------------------------------------------------------------------------------")
	for _, u := range users {
		fmt.Printf("%-5d %-30s %-30s %-10s\n", u.ID, u.Email, u.Name, u.Theme)
	}
	return nil
}

func userGet(svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("user get", flag.ExitOnError)
	id := fs.Int64("id", 0, "user ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}

	user, err := svc.GetByID(context.Background(), *id)
	if err != nil {
		return err
	}

	fmt.Printf("ID:        %d\n", user.ID)
	fmt.Printf("Email:     %s\n", user.Email)
	fmt.Printf("Name:      %s\n", user.Name)
	fmt.Printf("Theme:     %s\n", user.Theme)
	fmt.Printf("Created:   %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:   %s\n", user.UpdatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func userUpdate(svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("user update", flag.ExitOnError)
	id := fs.Int64("id", 0, "user ID (required)")
	email := fs.String("email", "", "new email (required)")
	name := fs.String("name", "", "new name (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}
	if *email == "" || *name == "" {
		return errors.New("--email and --name are required")
	}

	user, err := svc.Update(context.Background(), *id, *email, *name)
	if err != nil {
		return err
	}

	fmt.Printf("Updated user: ID=%d, Email=%s, Name=%s\n", user.ID, user.Email, user.Name)
	return nil
}

func userDelete(svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("user delete", flag.ExitOnError)
	id := fs.Int64("id", 0, "user ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == 0 {
		return errors.New("--id is required")
	}

	if err := svc.Delete(context.Background(), *id); err != nil {
		return err
	}

	fmt.Printf("Deleted user with ID=%d\n", *id)
	return nil
}

func printUserUsage() {
	fmt.Println("Usage: cli user <subcommand> [options]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  create   Create a new user")
	fmt.Println("  list     List all users")
	fmt.Println("  get      Get user by ID")
	fmt.Println("  update   Update a user")
	fmt.Println("  delete   Delete a user")
	fmt.Println()
	fmt.Println("Options for create:")
	fmt.Println("  --email string   User email (required)")
	fmt.Println("  --name string    User name (required)")
	fmt.Println()
	fmt.Println("Options for get/delete:")
	fmt.Println("  --id int         User ID (required)")
	fmt.Println()
	fmt.Println("Options for update:")
	fmt.Println("  --id int         User ID (required)")
	fmt.Println("  --email string   New email (required)")
	fmt.Println("  --name string    New name (required)")
}
