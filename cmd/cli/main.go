package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"claude-test/internal/database"
	"claude-test/internal/repository"
	"claude-test/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "user":
		if err := userCmd(userService, args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "role":
		if err := roleCmd(roleService, userService, args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "policy":
		if err := policyCmd(db, args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cli <command> <subcommand> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  user    Manage users")
	fmt.Println("  role    Manage roles and user role assignments")
	fmt.Println("  policy  Manage authorization policies (Casbin)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli user list")
	fmt.Println("  cli user create --email user@example.com --name \"John Doe\"")
	fmt.Println("  cli user get --id 1")
	fmt.Println("  cli user update --id 1 --email new@example.com --name \"Jane Doe\"")
	fmt.Println("  cli user delete --id 1")
	fmt.Println()
	fmt.Println("  cli role list")
	fmt.Println("  cli role create --name admin --description \"Administrator role\"")
	fmt.Println("  cli role get --id 1")
	fmt.Println("  cli role update --id 1 --name \"Admin\" --description \"Updated\"")
	fmt.Println("  cli role delete --id 1")
	fmt.Println("  cli role assign --user-id 1 --role-ids 1,2")
	fmt.Println("  cli role remove --user-id 1")
	fmt.Println("  cli role user-roles --user-id 1")
	fmt.Println()
	fmt.Println("  cli policy list")
	fmt.Println("  cli policy init")
	fmt.Println("  cli policy add --role Admin --resource /web/users --action read")
	fmt.Println("  cli policy remove --role Admin --resource /web/users --action read")
	fmt.Println("  cli policy check --role Admin --resource /web/users --action read")
}
