package cmd

import (
	"bufio"
	"codebase-app/internal/adapter"
	"codebase-app/internal/infrastructure/config"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// RunDropboxOAuth handles OAuth2 setup for Dropbox integration
func RunDropboxOAuth(cmd *flag.FlagSet, args []string) {
	// Parse command line arguments
	action := cmd.String("action", "auth", "OAuth action: 'auth' to get authorization URL, 'exchange' to exchange code for tokens")
	code := cmd.String("code", "", "Authorization code received from Dropbox (for exchange action)")
	if err := cmd.Parse(args); err != nil {
		log.Error().Err(err).Msg("failed to parse args")
		return
	}

	log.Info().Str("action", *action).Msg("starting Dropbox OAuth process")

	switch *action {
	case "auth":
		generateAuthURL()
	case "exchange":
		if *code == "" {
			log.Error().Msg("authorization code is required for exchange action")
			fmt.Println("Error: Authorization code is required for exchange action")
			fmt.Println("Usage: ./kaffah-finance dropbox-oauth --action=exchange --code=YOUR_AUTH_CODE")
			os.Exit(1)
		}
		exchangeCodeForTokens(*code)
	default:
		log.Error().Str("action", *action).Msg("invalid action")
		fmt.Printf("Error: Invalid action '%s'. Use 'auth' or 'exchange'\n", *action)
		os.Exit(1)
	}
}

// generateAuthURL generates and displays the OAuth2 authorization URL
func generateAuthURL() {
	// Check if required configuration is present
	if config.Envs.Dropbox.AppKey == "" {
		log.Error().Msg("DROPBOX_APP_KEY is not configured")
		fmt.Println("Error: DROPBOX_APP_KEY environment variable is required")
		fmt.Println("Please set your Dropbox app key in the environment variables")
		os.Exit(1)
	}

	redirectURL := config.Envs.Dropbox.RedirectURL
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/dropbox/callback"
	}

	// Generate authorization URL
	authURL := adapter.GetDropboxAuthURL(config.Envs.Dropbox.AppKey, redirectURL)

	fmt.Println("=== Dropbox OAuth2 Setup ===")
	fmt.Println()
	fmt.Printf("App Key: %s\n", config.Envs.Dropbox.AppKey)
	fmt.Printf("Redirect URL: %s\n", redirectURL)
	fmt.Println()
	fmt.Println("Step 1: Open the following URL in your browser:")
	fmt.Println()
	fmt.Printf("    %s\n", authURL)
	fmt.Println()
	fmt.Println("Step 2: Authorize your application and copy the authorization code from the callback URL")
	fmt.Println()
	fmt.Println("Step 3: Run the following command with your authorization code:")
	fmt.Println()
	fmt.Printf("    ./kaffah-finance dropbox-oauth --action=exchange --code=YOUR_AUTH_CODE\n")
	fmt.Println()

	log.Info().Str("authURL", authURL).Msg("authorization URL generated")
}

// exchangeCodeForTokens exchanges the authorization code for access and refresh tokens
func exchangeCodeForTokens(authCode string) {
	// Check if required configuration is present
	if config.Envs.Dropbox.AppKey == "" || config.Envs.Dropbox.AppSecret == "" {
		log.Error().Msg("DROPBOX_APP_KEY or DROPBOX_APP_SECRET is not configured")
		fmt.Println("Error: DROPBOX_APP_KEY and DROPBOX_APP_SECRET environment variables are required")
		os.Exit(1)
	}

	redirectURL := config.Envs.Dropbox.RedirectURL
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/dropbox/callback"
	}

	log.Info().Str("authCode", authCode).Msg("exchanging authorization code for tokens")

	// Exchange code for tokens
	tokenResp, err := adapter.ExchangeCodeForToken(
		authCode,
		config.Envs.Dropbox.AppKey,
		config.Envs.Dropbox.AppSecret,
		redirectURL,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to exchange code for tokens")
		fmt.Printf("Error: Failed to exchange code for tokens: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== OAuth2 Token Exchange Successful ===")
	fmt.Println()
	fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
	fmt.Printf("Token Type: %s\n", tokenResp.TokenType)
	fmt.Printf("Expires In: %d seconds\n", tokenResp.ExpiresIn)
	if tokenResp.RefreshToken != "" {
		fmt.Printf("Refresh Token: %s\n", tokenResp.RefreshToken)
	}
	if tokenResp.Scope != "" {
		fmt.Printf("Scope: %s\n", tokenResp.Scope)
	}
	fmt.Println()
	fmt.Println("=== Environment Variables ===")
	fmt.Println()
	fmt.Printf("Add these to your .env file:\n")
	fmt.Println()
	fmt.Printf("DROPBOX_ACCESS_TOKEN=%s\n", tokenResp.AccessToken)
	if tokenResp.RefreshToken != "" {
		fmt.Printf("DROPBOX_REFRESH_TOKEN=%s\n", tokenResp.RefreshToken)
	}
	fmt.Println()

	// Ask if user wants to update the configuration automatically
	fmt.Print("Do you want to automatically update your .env file? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Warn().Err(err).Msg("failed to read user input")
		return
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		updateEnvFile(tokenResp.AccessToken, tokenResp.RefreshToken)
	} else {
		fmt.Println("Please manually add the environment variables to your .env file")
	}

	log.Info().Msg("token exchange completed successfully")
}

// updateEnvFile updates the .env file with new tokens
func updateEnvFile(accessToken, refreshToken string) {
	envFile := ".env"

	// Check if .env file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		log.Warn().Msg(".env file not found, creating new one")

		// Create new .env file
		file, err := os.Create(envFile)
		if err != nil {
			log.Error().Err(err).Msg("failed to create .env file")
			fmt.Printf("Error: Failed to create .env file: %v\n", err)
			return
		}
		defer file.Close()

		// Write tokens to new file
		content := fmt.Sprintf("DROPBOX_ACCESS_TOKEN=%s\n", accessToken)
		if refreshToken != "" {
			content += fmt.Sprintf("DROPBOX_REFRESH_TOKEN=%s\n", refreshToken)
		}

		if _, err := file.WriteString(content); err != nil {
			log.Error().Err(err).Msg("failed to write to .env file")
			fmt.Printf("Error: Failed to write to .env file: %v\n", err)
			return
		}

		fmt.Printf("Created .env file with Dropbox tokens\n")
		log.Info().Msg("created new .env file with tokens")
		return
	}

	// TODO: Update existing .env file (append or replace existing tokens)
	// This is a simple implementation - you might want to make it more sophisticated
	fmt.Println("Note: .env file exists. Please manually update the following variables:")
	fmt.Printf("DROPBOX_ACCESS_TOKEN=%s\n", accessToken)
	if refreshToken != "" {
		fmt.Printf("DROPBOX_REFRESH_TOKEN=%s\n", refreshToken)
	}

	log.Info().Msg(".env file update instructions provided")
}
