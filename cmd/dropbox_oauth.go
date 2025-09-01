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
	fnName := "cmd::RunDropboxOAuth"

	// Parse command line arguments
	action := cmd.String("action", "auth", "OAuth action: 'auth' to get authorization URL, 'exchange' to exchange code for tokens")
	code := cmd.String("code", "", "Authorization code received from Dropbox (for exchange action)")
	cmd.Parse(args)

	log.Info().Str("action", *action).Msgf("%s - starting Dropbox OAuth process", fnName)

	switch *action {
	case "auth":
		generateAuthURL()
	case "exchange":
		if *code == "" {
			log.Error().Msgf("%s - authorization code is required for exchange action", fnName)
			fmt.Println("Error: Authorization code is required for exchange action")
			fmt.Println("Usage: ./kaffah-finance dropbox-oauth --action=exchange --code=YOUR_AUTH_CODE")
			os.Exit(1)
		}
		exchangeCodeForTokens(*code)
	default:
		log.Error().Str("action", *action).Msgf("%s - invalid action", fnName)
		fmt.Printf("Error: Invalid action '%s'. Use 'auth' or 'exchange'\n", *action)
		os.Exit(1)
	}
}

// generateAuthURL generates and displays the OAuth2 authorization URL
func generateAuthURL() {
	fnName := "cmd::generateAuthURL"

	// Check if required configuration is present
	if config.Envs.Dropbox.AppKey == "" {
		log.Error().Msgf("%s - DROPBOX_APP_KEY is not configured", fnName)
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

	log.Info().Str("authURL", authURL).Msgf("%s - authorization URL generated", fnName)
}

// exchangeCodeForTokens exchanges the authorization code for access and refresh tokens
func exchangeCodeForTokens(authCode string) {
	fnName := "cmd::exchangeCodeForTokens"

	// Check if required configuration is present
	if config.Envs.Dropbox.AppKey == "" || config.Envs.Dropbox.AppSecret == "" {
		log.Error().Msgf("%s - DROPBOX_APP_KEY or DROPBOX_APP_SECRET is not configured", fnName)
		fmt.Println("Error: DROPBOX_APP_KEY and DROPBOX_APP_SECRET environment variables are required")
		os.Exit(1)
	}

	redirectURL := config.Envs.Dropbox.RedirectURL
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/dropbox/callback"
	}

	log.Info().Str("authCode", authCode).Msgf("%s - exchanging authorization code for tokens", fnName)

	// Exchange code for tokens
	tokenResp, err := adapter.ExchangeCodeForToken(
		authCode,
		config.Envs.Dropbox.AppKey,
		config.Envs.Dropbox.AppSecret,
		redirectURL,
	)
	if err != nil {
		log.Error().Err(err).Msgf("%s - failed to exchange code for tokens", fnName)
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
		log.Warn().Err(err).Msgf("%s - failed to read user input", fnName)
		return
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		updateEnvFile(tokenResp.AccessToken, tokenResp.RefreshToken)
	} else {
		fmt.Println("Please manually add the environment variables to your .env file")
	}

	log.Info().Msgf("%s - token exchange completed successfully", fnName)
}

// updateEnvFile updates the .env file with new tokens
func updateEnvFile(accessToken, refreshToken string) {
	fnName := "cmd::updateEnvFile"

	envFile := ".env"

	// Check if .env file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		log.Warn().Msgf("%s - .env file not found, creating new one", fnName)

		// Create new .env file
		file, err := os.Create(envFile)
		if err != nil {
			log.Error().Err(err).Msgf("%s - failed to create .env file", fnName)
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
			log.Error().Err(err).Msgf("%s - failed to write to .env file", fnName)
			fmt.Printf("Error: Failed to write to .env file: %v\n", err)
			return
		}

		fmt.Printf("Created .env file with Dropbox tokens\n")
		log.Info().Msgf("%s - created new .env file with tokens", fnName)
		return
	}

	// TODO: Update existing .env file (append or replace existing tokens)
	// This is a simple implementation - you might want to make it more sophisticated
	fmt.Println("Note: .env file exists. Please manually update the following variables:")
	fmt.Printf("DROPBOX_ACCESS_TOKEN=%s\n", accessToken)
	if refreshToken != "" {
		fmt.Printf("DROPBOX_REFRESH_TOKEN=%s\n", refreshToken)
	}

	log.Info().Msgf("%s - .env file update instructions provided", fnName)
}
