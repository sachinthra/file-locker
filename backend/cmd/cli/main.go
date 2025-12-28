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
	BaseURL string `json:"base_url,omitempty"`
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

func saveToken(token string) error {
	p, err := cfgPath()
	if err != nil {
		return err
	}
	cfg := CLIConfig{Token: token, BaseURL: apiBase}
	b, _ := json.Marshal(cfg)
	return os.WriteFile(p, b, 0600)
}

func loadToken() (string, error) {
	p, err := cfgPath()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	var cfg CLIConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
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

func getBaseURL(cfg *CLIConfig) string {
	if cfg != nil && cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return apiBase
}

func doRequest(method, path, token string, body io.Reader, contentType string) (*http.Response, error) {
	baseURL := apiBase
	if p, err := cfgPath(); err == nil {
		if b, err := os.ReadFile(p); err == nil {
			var cfg CLIConfig
			if json.Unmarshal(b, &cfg) == nil {
				baseURL = getBaseURL(&cfg)
			}
		}
	}
	
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
	fs.Parse(args)
	
	// Token-based login (preferred)
	if *tokenPtr != "" {
		// Validate token by calling an auth-protected endpoint
		resp, err := doRequest("GET", "/files", *tokenPtr, nil, "")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("invalid token (status %d)", resp.StatusCode)
		}
		if err := saveToken(*tokenPtr); err != nil {
			return err
		}
		fmt.Println("Successfully logged in with Personal Access Token!")
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
		defer resp.Body.Close()
		
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
		
		if err := saveToken(result.Token); err != nil {
			return err
		}
		fmt.Printf("Successfully logged in as %s!\n", result.User.Username)
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
	defer resp.Body.Close()
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
	fmt.Fprintln(w, "ID\tNAME\tSIZE\tUPLOADED\tEXPIRES")
	fmt.Fprintln(w, "---\t----\t----\t--------\t-------")
	
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
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, f.FileName, size, uploaded, expires)
	}
	w.Flush()
	
	return nil
}

func uploadWithProgress(token, path string, tags string, expireHours int) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
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
			writer.WriteField("tags", tags)
		}
		if expireHours > 0 {
			writer.WriteField("expire_hours", fmt.Sprint(expireHours))
		}
		
		writer.Close()
		done <- nil
	}()
	
	// Get base URL
	baseURL := apiBase
	if p, err := cfgPath(); err == nil {
		if b, err := os.ReadFile(p); err == nil {
			var cfg CLIConfig
			if json.Unmarshal(b, &cfg) == nil {
				baseURL = getBaseURL(&cfg)
			}
		}
	}
	
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
	
	if resp.StatusCode != 200 {
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
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	tags := fs.String("tags", "", "comma separated tags")
	expire := fs.Int("expire", 0, "expiration time in hours")
	fs.Parse(args)
	args = fs.Args()
	if len(args) < 1 {
		return errors.New("file path required")
	}
	path := args[0]
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

func printUsage() {
	fmt.Println("fl - File Locker CLI")
	fmt.Println("\nCommands:")
	fmt.Println("  login --token <token>              Login with Personal Access Token")
	fmt.Println("  login -u <user> -p <pass>          Login with username/password")
	fmt.Println("  ls [--json]                        List files (table or JSON)")
	fmt.Println("  upload <file> [--tags t1,t2]       Upload file with optional tags")
	fmt.Println("                [--expire 24]        Set expiration in hours")
	fmt.Println("  download <file_id> [-o filename]   Download file")
	fmt.Println("  rm <file_id>                       Delete file")
	fmt.Println("\nExamples:")
	fmt.Println("  fl login --token fl_abc123...")
	fmt.Println("  fl upload document.pdf --tags work,important --expire 72")
	fmt.Println("  fl ls")
	fmt.Println("  fl download a1b2c3d4")
	fmt.Println("  fl rm a1b2c3d4")
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
	default:
		printUsage()
	}
}
