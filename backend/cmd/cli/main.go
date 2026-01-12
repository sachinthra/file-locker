package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
)

const (
	configDir  = ".filelocker"
	configFile = "config.json"
	apiBase    = "http://localhost:9010/api/v1"
)

type CLIConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

func cfgPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

func saveConfig(cfg CLIConfig) error {
	p, err := cfgPath()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(p, b, 0600)
}

func loadConfig() (*CLIConfig, error) {
	p, err := cfgPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var cfg CLIConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadToken() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	if cfg.Token == "" {
		return "", errors.New("no token found")
	}
	return cfg.Token, nil
}

func httpClient(token string) *http.Client {
	client := &http.Client{Timeout: 0}
	return client
}

func normalizeBaseURL(host string) string {
	// Remove trailing slash
	host = strings.TrimSuffix(host, "/")

	// If host doesn't end with /api or /api/v1, append /api/v1
	if !strings.HasSuffix(host, "/api") && !strings.HasSuffix(host, "/api/v1") {
		return host + "/api/v1"
	}

	// If ends with /api but not /api/v1, append /v1
	if strings.HasSuffix(host, "/api") && !strings.HasSuffix(host, "/api/v1") {
		return host + "/v1"
	}

	return host
}

func getBaseURL() string {
	cfg, err := loadConfig()
	if err == nil && cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return apiBase
}

func doRequest(method, path, token string, body io.Reader, contentType string) (*http.Response, error) {
	baseURL := getBaseURL()

	req, _ := http.NewRequest(method, baseURL+path, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := httpClient(token)
	resp, err := client.Do(req)

	// Handle 401 Unauthorized
	if err == nil && resp.StatusCode == 401 {
		fmt.Fprintln(os.Stderr, "Session expired or invalid token. Please run 'fl login'.")
		os.Exit(1)
	}

	return resp, err
}

func cmdLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	tokenPtr := fs.String("token", "", "personal access token")
	userPtr := fs.String("u", "", "username")
	passPtr := fs.String("p", "", "password")
	hostPtr := fs.String("host", "", "server URL (e.g., http://raspberrypi.local:8080)")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// Load existing config or create new one
	cfg, err := loadConfig()
	if err != nil {
		cfg = &CLIConfig{BaseURL: apiBase}
	}

	// Update base URL if --host is provided
	if *hostPtr != "" {
		cfg.BaseURL = normalizeBaseURL(*hostPtr)
		fmt.Printf("Using server: %s\n", cfg.BaseURL)
	}

	// Token-based login (preferred)
	if *tokenPtr != "" {
		// Save config with new host before validating token
		cfg.Token = *tokenPtr
		if err := saveConfig(*cfg); err != nil {
			return err
		}

		// Validate token by calling an auth-protected endpoint
		resp, err := doRequest("GET", "/files", *tokenPtr, nil, "")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != 200 {
			return fmt.Errorf("invalid token (status %d)", resp.StatusCode)
		}
		fmt.Println("✅ Successfully logged in with Personal Access Token!")
		return nil
	}

	// Username/Password login (legacy)
	if *userPtr != "" && *passPtr != "" {
		payload := map[string]string{
			"username": *userPtr,
			"password": *passPtr,
		}
		body, _ := json.Marshal(payload)

		resp, err := doRequest("POST", "/auth/login", "", strings.NewReader(string(body)), "application/json")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != 200 {
			return fmt.Errorf("login failed (status %d)", resp.StatusCode)
		}

		var result struct {
			Token string `json:"token"`
			User  struct {
				Username string `json:"username"`
			} `json:"user"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		cfg.Token = result.Token
		if err := saveConfig(*cfg); err != nil {
			return err
		}
		fmt.Printf("✅ Successfully logged in as %s!\n", result.User.Username)
		return nil
	}

	return errors.New("either --token or both -u and -p are required")
}

func cmdLs(jsonOut bool) error {
	token, err := loadToken()
	if err != nil {
		return err
	}
	resp, err := doRequest("GET", "/files", token, nil, "")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return fmt.Errorf("error: %s", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	if jsonOut {
		fmt.Println(string(body))
		return nil
	}

	var parsed struct {
		Files []struct {
			ID        string     `json:"file_id"`
			FileName  string     `json:"file_name"`
			Size      int64      `json:"size"`
			CreatedAt time.Time  `json:"created_at"`
			ExpiresAt *time.Time `json:"expires_at"`
		} `json:"files"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}

	if len(parsed.Files) == 0 {
		fmt.Println("No files found.")
		return nil
	}

	// Use tabwriter for clean table formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tSIZE\tUPLOADED\tEXPIRES")
	_, _ = fmt.Fprintln(w, "---\t----\t----\t--------\t-------")

	for _, f := range parsed.Files {
		id := f.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}

		size := humanize.Bytes(uint64(f.Size))
		uploaded := humanize.Time(f.CreatedAt)

		expires := "Never"
		if f.ExpiresAt != nil {
			if f.ExpiresAt.Before(time.Now()) {
				expires = "Expired"
			} else {
				expires = "In " + humanize.Time(*f.ExpiresAt)
			}
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, f.FileName, size, uploaded, expires)
	}
	w.Flush()

	return nil
}

func uploadWithProgress(token, path string, tags string, expireHours int) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Create progress bar
	bar := progressbar.NewOptions64(
		stat.Size(),
		progressbar.OptionSetDescription("Uploading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Create pipe for streaming upload
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Error channel for goroutine
	done := make(chan error, 1)

	// Write multipart form in goroutine
	go func() {
		defer pw.Close()

		// Add file part
		part, err := writer.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			done <- err
			return
		}

		// Copy file through progress bar
		_, err = io.Copy(part, io.TeeReader(file, bar))
		if err != nil {
			done <- err
			return
		}

		// Add optional fields
		if tags != "" {
			_ = writer.WriteField("tags", tags)
		}
		if expireHours > 0 {
			_ = writer.WriteField("expire_after", fmt.Sprint(expireHours))
		}

		writer.Close()
		done <- nil
	}()

	// Get base URL
	baseURL := getBaseURL()

	// Create request
	req, err := http.NewRequest("POST", baseURL+"/upload", pr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := httpClient(token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(b))
	}

	// Wait for upload goroutine
	if err := <-done; err != nil {
		return err
	}

	// Parse response
	var result struct {
		FileID   string `json:"file_id"`
		FileName string `json:"file_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		fmt.Printf("Successfully uploaded: %s (ID: %s)\n", result.FileName, result.FileID[:8]+"...")
	} else {
		fmt.Println("Upload complete!")
	}

	return nil
}

func cmdUpload(args []string) error {
	if len(args) < 1 {
		return errors.New("file path required")
	}
	path := args[0]

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	args = args[1:]

	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	tags := fs.String("tags", "", "comma separated tags")
	expire := fs.Int("expire", 0, "expiration time in hours")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	token, err := loadToken()
	if err != nil {
		return err
	}

	return uploadWithProgress(token, path, *tags, *expire)
}

func cmdDownload(args []string) error {
	fs := flag.NewFlagSet("download", flag.ContinueOnError)
	output := fs.String("o", "", "output filename (default: from server)")
	fs.Parse(args)
	args = fs.Args()
	if len(args) < 1 {
		return errors.New("file id required")
	}
	id := args[0]
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/download/"+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(b))
	}

	// Determine output filename
	filename := *output
	if filename == "" {
		// Try to get filename from Content-Disposition header
		if cd := resp.Header.Get("Content-Disposition"); cd != "" {
			_, params, err := mime.ParseMediaType(cd)
			if err == nil && params["filename"] != "" {
				filename = params["filename"]
			}
		}
		// Fallback to file ID
		if filename == "" {
			filename = filepath.Base(id)
		}
	}

	// Create output file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create progress bar
	total := resp.ContentLength
	if total < 0 {
		total = 0
	}

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Download with progress
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded to: %s\n", filename)
	return nil
}

func cmdRm(args []string) error {
	fs := flag.NewFlagSet("rm", flag.ContinueOnError)
	fs.Parse(args)
	args = fs.Args()
	if len(args) < 1 {
		return errors.New("file id required")
	}
	id := args[0]
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("DELETE", "/files?id="+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Printf("Successfully deleted file: %s\n", id)
	return nil
}

func cmdLogout() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/auth/logout", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Clear config
	cfg := CLIConfig{BaseURL: getBaseURL()}
	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Println("✅ Logged out successfully")
	return nil
}

func cmdMe() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/auth/me", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get user info (status %d): %s", resp.StatusCode, string(b))
	}

	var user struct {
		ID        string    `json:"user_id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return err
	}

	fmt.Printf("User ID:      %s\n", user.ID)
	fmt.Printf("Username:     %s\n", user.Username)
	fmt.Printf("Email:        %s\n", user.Email)
	fmt.Printf("Role:         %s\n", user.Role)
	fmt.Printf("Member Since: %s\n", user.CreatedAt.Format("2006-01-02"))
	return nil
}

func cmdSearch(args []string) error {
	if len(args) < 1 {
		return errors.New("search query required")
	}
	query := args[0]
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/files/search?q="+query, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("search failed (status %d)", resp.StatusCode)
	}

	var result struct {
		Files []struct {
			ID        string    `json:"file_id"`
			FileName  string    `json:"file_name"`
			Size      int64     `json:"size"`
			CreatedAt time.Time `json:"created_at"`
			Tags      []string  `json:"tags"`
		} `json:"files"`
		Count int `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Count == 0 {
		fmt.Println("No files found matching query.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tSIZE\tTAGS\n")
	fmt.Fprintf(w, "---\t----\t----\t----\n")

	for _, f := range result.Files {
		id := f.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		tags := strings.Join(f.Tags, ", ")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, f.FileName, humanize.Bytes(uint64(f.Size)), tags)
	}
	w.Flush()

	fmt.Printf("\nFound %d file(s)\n", result.Count)
	return nil
}

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	output := fs.String("o", "filelocker-export.zip", "output filename")
	fs.Parse(args)

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/files/export", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("export failed (status %d): %s", resp.StatusCode, string(b))
	}

	f, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer f.Close()

	total := resp.ContentLength
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription("Exporting"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
	)

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Exported to: %s\n", *output)
	return nil
}

func cmdUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	tags := fs.String("tags", "", "comma separated tags")
	name := fs.String("name", "", "new filename")
	fs.Parse(args)
	args = fs.Args()

	if len(args) < 1 {
		return errors.New("file id required")
	}
	id := args[0]

	if *tags == "" && *name == "" {
		return errors.New("either --tags or --name required")
	}

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := make(map[string]interface{})
	if *tags != "" {
		payload["tags"] = strings.Split(*tags, ",")
	}
	if *name != "" {
		payload["file_name"] = *name
	}

	body, _ := json.Marshal(payload)
	resp, err := doRequest("PATCH", "/files/"+id, token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ File updated successfully")
	return nil
}

func cmdTokens(args []string) error {
	if len(args) < 1 {
		return errors.New("subcommand required: list, create, revoke")
	}

	subcmd := args[0]
	switch subcmd {
	case "list":
		return cmdTokensList()
	case "create":
		return cmdTokensCreate(args[1:])
	case "revoke":
		return cmdTokensRevoke(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcmd)
	}
}

func cmdTokensList() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/auth/tokens", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list tokens (status %d)", resp.StatusCode)
	}

	var tokens []struct {
		ID        string     `json:"id"`
		Name      string     `json:"name"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt *time.Time `json:"expires_at"`
		LastUsed  *time.Time `json:"last_used"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return err
	}

	if len(tokens) == 0 {
		fmt.Println("No tokens found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tCREATED\tEXPIRES\tLAST USED\n")
	fmt.Fprintf(w, "---\t----\t-------\t-------\t---------\n")

	for _, t := range tokens {
		id := t.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		created := humanize.Time(t.CreatedAt)
		expires := "Never"
		if t.ExpiresAt != nil {
			expires = humanize.Time(*t.ExpiresAt)
		}
		lastUsed := "Never"
		if t.LastUsed != nil {
			lastUsed = humanize.Time(*t.LastUsed)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, t.Name, created, expires, lastUsed)
	}
	w.Flush()
	return nil
}

func cmdTokensCreate(args []string) error {
	if len(args) < 1 {
		return errors.New("token name required")
	}
	name := args[0]

	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	expire := fs.String("expire", "", "expiration date (YYYY-MM-DD)")
	fs.Parse(args[1:])

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"name": name,
	}
	if *expire != "" {
		payload["expires_at"] = *expire
	}

	body, _ := json.Marshal(payload)
	resp, err := doRequest("POST", "/auth/tokens", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create token (status %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Token   string `json:"token"`
		TokenID string `json:"token_id"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Println("✅ Token created successfully!")
	fmt.Printf("Name:  %s\n", result.Name)
	fmt.Printf("Token: %s\n\n", result.Token)
	fmt.Println("⚠️  Save this token now - you won't be able to see it again!")
	return nil
}

func cmdTokensRevoke(args []string) error {
	if len(args) < 1 {
		return errors.New("token id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("DELETE", "/auth/tokens/"+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to revoke token (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Token revoked successfully")
	return nil
}

func cmdPassword(args []string) error {
	fs := flag.NewFlagSet("password", flag.ContinueOnError)
	old := fs.String("old", "", "old password")
	new := fs.String("new", "", "new password")
	fs.Parse(args)

	if *old == "" || *new == "" {
		return errors.New("both --old and --new passwords required")
	}

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]string{
		"old_password": *old,
		"new_password": *new,
	}

	body, _ := json.Marshal(payload)
	resp, err := doRequest("PATCH", "/user/password", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to change password (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Password changed successfully")
	return nil
}

func cmdAnnouncements(args []string) error {
	if len(args) == 0 {
		return cmdAnnouncementsList()
	}

	subcmd := args[0]
	if subcmd == "dismiss" {
		return cmdAnnouncementsDismiss(args[1:])
	}

	return fmt.Errorf("unknown subcommand: %s", subcmd)
}

func cmdAnnouncementsList() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/announcements", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list announcements (status %d)", resp.StatusCode)
	}

	var announcements []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&announcements); err != nil {
		return err
	}

	if len(announcements) == 0 {
		fmt.Println("No announcements.")
		return nil
	}

	for i, a := range announcements {
		var emoji string
		switch a.Severity {
		case "warning":
			emoji = "⚠️"
		case "error":
			emoji = "❌"
		default:
			emoji = "ℹ️"
		}
		fmt.Printf("%s [%s] %s\n", emoji, a.Severity, a.Title)
		fmt.Printf("   %s\n", a.Message)
		fmt.Printf("   ID: %s\n", a.ID)
		if i < len(announcements)-1 {
			fmt.Println()
		}
	}
	return nil
}

func cmdAnnouncementsDismiss(args []string) error {
	if len(args) < 1 {
		return errors.New("announcement id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/announcements/"+id+"/dismiss", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to dismiss announcement (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Announcement dismissed")
	return nil
}

func cmdAdmin(args []string) error {
	if len(args) < 1 {
		return errors.New("admin subcommand required: stats, users, settings, files, storage, logs, announcements")
	}

	subcmd := args[0]
	switch subcmd {
	case "stats":
		return cmdAdminStats()
	case "users":
		return cmdAdminUsers(args[1:])
	case "settings":
		return cmdAdminSettings(args[1:])
	case "files":
		return cmdAdminFiles(args[1:])
	case "storage":
		return cmdAdminStorage(args[1:])
	case "logs":
		return cmdAdminLogs(args[1:])
	case "announcements":
		return cmdAdminAnnouncements(args[1:])
	default:
		return fmt.Errorf("unknown admin subcommand: %s", subcmd)
	}
}

func cmdAdminStats() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/admin/stats", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get stats (status %d)", resp.StatusCode)
	}

	var stats struct {
		TotalUsers        int   `json:"total_users"`
		TotalFiles        int   `json:"total_files"`
		TotalStorageBytes int64 `json:"total_storage_bytes"`
		ActiveSessions    int   `json:"active_sessions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return err
	}

	fmt.Println("📊 System Statistics")
	fmt.Printf("Total Users:     %d\n", stats.TotalUsers)
	fmt.Printf("Total Files:     %d\n", stats.TotalFiles)
	fmt.Printf("Storage Used:    %s\n", humanize.Bytes(uint64(stats.TotalStorageBytes)))
	fmt.Printf("Active Sessions: %d\n", stats.ActiveSessions)
	return nil
}

func cmdAdminUsers(args []string) error {
	if len(args) == 0 {
		return cmdAdminUsersList(args)
	}

	subcmd := args[0]
	switch subcmd {
	case "approve":
		return cmdAdminUsersApprove(args[1:])
	case "reject":
		return cmdAdminUsersReject(args[1:])
	case "delete":
		return cmdAdminUsersDelete(args[1:])
	default:
		// Check if it's a user ID for status/role/reset-password/logout
		if len(args) >= 2 {
			userID := args[0]
			action := args[1]
			switch action {
			case "status":
				return cmdAdminUsersStatus(userID, args[2:])
			case "role":
				return cmdAdminUsersRole(userID, args[2:])
			case "reset-password":
				return cmdAdminUsersResetPassword(userID)
			case "logout":
				return cmdAdminUsersLogout(userID)
			}
		}
		return cmdAdminUsersList(args)
	}
}

func cmdAdminUsersList(args []string) error {
	fs := flag.NewFlagSet("users", flag.ContinueOnError)
	status := fs.String("status", "", "filter by status: active, inactive, pending")
	fs.Parse(args)

	token, err := loadToken()
	if err != nil {
		return err
	}

	path := "/admin/users"
	if *status != "" {
		path += "?status=" + *status
	}

	resp, err := doRequest("GET", path, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list users (status %d)", resp.StatusCode)
	}

	var result struct {
		Users []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		} `json:"users"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tUSERNAME\tEMAIL\tROLE\n")
	fmt.Fprintf(w, "---\t--------\t-----\t----\n")

	for _, u := range result.Users {
		id := u.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, u.Username, u.Email, u.Role)
	}
	w.Flush()
	return nil
}

func cmdAdminUsersApprove(args []string) error {
	if len(args) < 1 {
		return errors.New("user id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/admin/users/"+id+"/approve", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to approve user (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User approved")
	return nil
}

func cmdAdminUsersReject(args []string) error {
	if len(args) < 1 {
		return errors.New("user id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/admin/users/"+id+"/reject", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to reject user (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User rejected")
	return nil
}

func cmdAdminUsersDelete(args []string) error {
	if len(args) < 1 {
		return errors.New("user id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("DELETE", "/admin/users/"+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User deleted")
	return nil
}

func cmdAdminUsersStatus(userID string, args []string) error {
	if len(args) < 1 {
		return errors.New("status required: active or inactive")
	}
	status := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]string{"status": status}
	body, _ := json.Marshal(payload)

	resp, err := doRequest("PATCH", "/admin/users/"+userID+"/status", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update status (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User status updated")
	return nil
}

func cmdAdminUsersRole(userID string, args []string) error {
	if len(args) < 1 {
		return errors.New("role required: user or admin")
	}
	role := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]string{"role": role}
	body, _ := json.Marshal(payload)

	resp, err := doRequest("PATCH", "/admin/users/"+userID+"/role", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update role (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User role updated")
	return nil
}

func cmdAdminUsersResetPassword(userID string) error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/admin/users/"+userID+"/reset-password", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to reset password (status %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Message           string `json:"message"`
		TemporaryPassword string `json:"temporary_password"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Println("✅ Password reset successfully")
	fmt.Printf("Temporary Password: %s\n", result.TemporaryPassword)
	return nil
}

func cmdAdminUsersLogout(userID string) error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/admin/users/"+userID+"/logout", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to logout user (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ User logged out")
	return nil
}

func cmdAdminSettings(args []string) error {
	if len(args) == 0 {
		return cmdAdminSettingsGet()
	}

	if len(args) >= 2 {
		return cmdAdminSettingsUpdate(args[0], args[1])
	}

	return errors.New("usage: admin settings [key value]")
}

func cmdAdminSettingsGet() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/admin/settings", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get settings (status %d)", resp.StatusCode)
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return err
	}

	fmt.Println("⚙️  System Settings:")
	for key, value := range settings {
		fmt.Printf("%s: %v\n", key, value)
	}
	return nil
}

func cmdAdminSettingsUpdate(key, value string) error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]string{"key": key, "value": value}
	body, _ := json.Marshal(payload)

	resp, err := doRequest("PATCH", "/admin/settings", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update setting (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Setting updated")
	return nil
}

func cmdAdminFiles(args []string) error {
	if len(args) >= 2 && args[0] == "delete" {
		return cmdAdminFilesDelete(args[1])
	}
	return cmdAdminFilesList()
}

func cmdAdminFilesList() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/admin/files", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list files (status %d)", resp.StatusCode)
	}

	var result struct {
		Files []struct {
			ID       string `json:"id"`
			FileName string `json:"filename"`
			Size     int64  `json:"size"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tFILENAME\tSIZE\n")
	fmt.Fprintf(w, "---\t--------\t----\n")

	for _, f := range result.Files {
		id := f.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", id, f.FileName, humanize.Bytes(uint64(f.Size)))
	}
	w.Flush()
	return nil
}

func cmdAdminFilesDelete(id string) error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("DELETE", "/admin/files/"+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete file (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ File deleted")
	return nil
}

func cmdAdminStorage(args []string) error {
	if len(args) < 1 {
		return errors.New("storage subcommand required: analyze or cleanup")
	}

	subcmd := args[0]
	switch subcmd {
	case "analyze":
		return cmdAdminStorageAnalyze()
	case "cleanup":
		return cmdAdminStorageCleanup()
	default:
		return fmt.Errorf("unknown storage subcommand: %s", subcmd)
	}
}

func cmdAdminStorageAnalyze() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/admin/storage/analyze", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to analyze storage (status %d)", resp.StatusCode)
	}

	var analysis struct {
		TotalFiles     int   `json:"total_files"`
		TotalSizeBytes int64 `json:"total_size_bytes"`
		OrphanedFiles  int   `json:"orphaned_files"`
		ExpiredFiles   int   `json:"expired_files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		return err
	}

	fmt.Println("💾 Storage Analysis:")
	fmt.Printf("Total Files:     %d\n", analysis.TotalFiles)
	fmt.Printf("Total Size:      %s\n", humanize.Bytes(uint64(analysis.TotalSizeBytes)))
	fmt.Printf("Orphaned Files:  %d\n", analysis.OrphanedFiles)
	fmt.Printf("Expired Files:   %d\n", analysis.ExpiredFiles)
	return nil
}

func cmdAdminStorageCleanup() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("POST", "/admin/storage/cleanup", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cleanup storage (status %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Message         string `json:"message"`
		FilesDeleted    int    `json:"files_deleted"`
		SpaceFreedBytes int64  `json:"space_freed_bytes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Println("🧹 Cleanup completed!")
	fmt.Printf("Files Deleted:  %d\n", result.FilesDeleted)
	fmt.Printf("Space Freed:    %s\n", humanize.Bytes(uint64(result.SpaceFreedBytes)))
	return nil
}

func cmdAdminLogs(args []string) error {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	action := fs.String("action", "", "filter by action")
	userID := fs.String("user_id", "", "filter by user id")
	fs.Parse(args)

	token, err := loadToken()
	if err != nil {
		return err
	}

	path := "/admin/logs?"
	if *action != "" {
		path += "action=" + *action + "&"
	}
	if *userID != "" {
		path += "user_id=" + *userID
	}

	resp, err := doRequest("GET", path, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get logs (status %d)", resp.StatusCode)
	}

	var result struct {
		Logs []struct {
			ID        string    `json:"id"`
			UserID    string    `json:"user_id"`
			Action    string    `json:"action"`
			Details   string    `json:"details"`
			IPAddress string    `json:"ip_address"`
			Timestamp time.Time `json:"timestamp"`
		} `json:"logs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, log := range result.Logs {
		fmt.Printf("[%s] %s - %s (%s) - %s\n",
			log.Timestamp.Format("2006-01-02 15:04:05"),
			log.Action,
			log.UserID[:8]+"...",
			log.IPAddress,
			log.Details)
	}
	return nil
}

func cmdAdminAnnouncements(args []string) error {
	if len(args) == 0 {
		return cmdAdminAnnouncementsList()
	}

	subcmd := args[0]
	switch subcmd {
	case "create":
		return cmdAdminAnnouncementsCreate(args[1:])
	case "delete":
		return cmdAdminAnnouncementsDelete(args[1:])
	default:
		return fmt.Errorf("unknown announcements subcommand: %s", subcmd)
	}
}

func cmdAdminAnnouncementsList() error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("GET", "/admin/announcements", token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to list announcements (status %d)", resp.StatusCode)
	}

	var announcements []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&announcements); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tSEVERITY\tTITLE\n")
	fmt.Fprintf(w, "---\t--------\t-----\n")

	for _, a := range announcements {
		id := a.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", id, a.Severity, a.Title)
	}
	w.Flush()
	return nil
}

func cmdAdminAnnouncementsCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	title := fs.String("title", "", "announcement title")
	message := fs.String("message", "", "announcement message")
	severity := fs.String("severity", "info", "severity: info, warning, error")
	fs.Parse(args)

	if *title == "" || *message == "" {
		return errors.New("--title and --message required")
	}

	token, err := loadToken()
	if err != nil {
		return err
	}

	payload := map[string]string{
		"title":    *title,
		"message":  *message,
		"severity": *severity,
	}

	body, _ := json.Marshal(payload)
	resp, err := doRequest("POST", "/admin/announcements", token, strings.NewReader(string(body)), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create announcement (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Announcement created")
	return nil
}

func cmdAdminAnnouncementsDelete(args []string) error {
	if len(args) < 1 {
		return errors.New("announcement id required")
	}
	id := args[0]

	token, err := loadToken()
	if err != nil {
		return err
	}

	resp, err := doRequest("DELETE", "/admin/announcements/"+id, token, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete announcement (status %d): %s", resp.StatusCode, string(b))
	}

	fmt.Println("✅ Announcement deleted")
	return nil
}

func printUsage() {
	fmt.Println("fl - File Locker CLI")
	fmt.Println("\n🔐 Authentication:")
	fmt.Println("  login --token <token>              Login with Personal Access Token")
	fmt.Println("  login -u <user> -p <pass>          Login with username/password")
	fmt.Println("  login --host <url>                 Set server URL")
	fmt.Println("  logout                             Logout and clear credentials")
	fmt.Println("  me                                 Show current user info")
	fmt.Println("  whoami                             Alias for 'me'")

	fmt.Println("\n📁 File Operations:")
	fmt.Println("  ls [--json]                        List files (table or JSON)")
	fmt.Println("  upload <file> [--tags t1,t2]       Upload file with optional tags")
	fmt.Println("                [--expire 24]        Set expiration in hours")
	fmt.Println("  download <file_id> [-o filename]   Download file")
	fmt.Println("  rm <file_id>                       Delete file")
	fmt.Println("  search <query>                     Search files by name or tags")
	fmt.Println("  export [-o output.zip]             Export all files as zip")
	fmt.Println("  update <file_id> --tags t1,t2      Update file metadata")
	fmt.Println("         <file_id> --name newname    Rename file")

	fmt.Println("\n🔑 Personal Access Tokens:")
	fmt.Println("  tokens list                        List all PATs")
	fmt.Println("  tokens create <name> [--expire]    Create new PAT")
	fmt.Println("  tokens revoke <token_id>           Revoke PAT")

	fmt.Println("\n👤 User Management:")
	fmt.Println("  password                           Change password (interactive)")
	fmt.Println("  password --old <old> --new <new>   Change password")

	fmt.Println("\n📢 Announcements:")
	fmt.Println("  announcements                      List announcements")
	fmt.Println("  announcements dismiss <id>         Dismiss announcement")

	fmt.Println("\n🛡️  Admin Commands:")
	fmt.Println("  admin stats                        System statistics")
	fmt.Println("  admin users [--status pending]     List users")
	fmt.Println("  admin users approve <id>           Approve user")
	fmt.Println("  admin users reject <id>            Reject user")
	fmt.Println("  admin users delete <id>            Delete user")
	fmt.Println("  admin users <id> status <active>   Update user status")
	fmt.Println("  admin users <id> role <admin>      Update user role")
	fmt.Println("  admin users <id> reset-password    Reset user password")
	fmt.Println("  admin users <id> logout            Force logout user")
	fmt.Println("  admin settings                     View system settings")
	fmt.Println("  admin settings <key> <value>       Update setting")
	fmt.Println("  admin files                        List all files")
	fmt.Println("  admin files delete <id>            Delete any file")
	fmt.Println("  admin storage analyze              Analyze storage usage")
	fmt.Println("  admin storage cleanup              Cleanup orphaned files")
	fmt.Println("  admin logs [--action] [--user_id]  View audit logs")
	fmt.Println("  admin announcements                List all announcements")
	fmt.Println("  admin announcements create         Create announcement")
	fmt.Println("  admin announcements delete <id>    Delete announcement")

	fmt.Println("\n📖 Examples:")
	fmt.Println("  fl login --token fl_abc123...")
	fmt.Println("  fl upload document.pdf --tags work,important --expire 72")
	fmt.Println("  fl search \"project files\"")
	fmt.Println("  fl tokens create \"CI/CD Pipeline\"")
	fmt.Println("  fl admin users --status pending")
	fmt.Println("  fl admin storage cleanup")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	cmd := os.Args[1]
	switch cmd {
	case "login":
		if err := cmdLogin(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "ls":
		fs := flag.NewFlagSet("ls", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "output json")
		fs.Parse(os.Args[2:])
		if err := cmdLs(*jsonOut); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "upload":
		if err := cmdUpload(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "download":
		if err := cmdDownload(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "rm":
		if err := cmdRm(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "logout":
		if err := cmdLogout(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "me", "whoami":
		if err := cmdMe(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "search":
		if err := cmdSearch(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "export":
		if err := cmdExport(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "update":
		if err := cmdUpdate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "tokens":
		if err := cmdTokens(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "password":
		if err := cmdPassword(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "announcements":
		if err := cmdAnnouncements(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "admin":
		if err := cmdAdmin(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		printUsage()
	}
}
