package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type Configuration struct {
	SteamCmdPath string `json:"steamcmd_path"`
}

func getTokenFromCache() (*oauth2.Token, error) {
	tokenFile := "token.json"
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}
	tk := &oauth2.Token{}
	err = json.Unmarshal(token, tk)
	if err != nil {
		return nil, err
	}
	return tk, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", ""}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	if runtime.GOOS == "windows" {
		args = append(args, "\""+url+"\"")
	} else {
		args = append(args, url)
	}

	return exec.Command(cmd, args...).Start()
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	// Generate OAuth2 URL and open it in the default web browser
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Your browser has been opened to visit:", authURL)
	err := openBrowser(authURL)
	if err != nil {
		log.Fatalf("Failed to open web browser: %s", err)
	}

	// Start a local HTTP server to listen for the OAuth2 callback
	listen := "localhost:80"
	fmt.Printf("Listening on http://%s/\n", listen)
	codeCh := make(chan string)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "Missing code", http.StatusBadRequest)
				return
			}
			codeCh <- code
		})
		http.ListenAndServe(listen, nil)
	}()

	// Wait for the authorization code
	code := <-codeCh

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func saveTokenToCache(token *oauth2.Token) {
	tokenFile := "token.json"
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Unable to marshal token: %v", err)
	}
	os.WriteFile(tokenFile, tokenJSON, 0600)
}

func getClient() (*gmail.Service, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, err
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, err
	}

	token, err := getTokenFromCache()
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveTokenToCache(token)
	}

	client := config.Client(context.Background(), token)
	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func readConfig() Configuration {
	configFile, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			// Create a default config.json file if it doesn't exist
			defaultConfig := Configuration{
				SteamCmdPath: "./steamcmd/steamcmd.exe",
			}
			err := createDefaultConfigFile(defaultConfig)
			if err != nil {
				log.Fatalf("Error creating default config file: %v", err)
			}
			return defaultConfig
		}
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}
	return config
}

func createDefaultConfigFile(config Configuration) error {
	configFile, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		return err
	}
	return nil
}

func fetchSteamGuardCode(srv *gmail.Service) string {
	user := "me"
	query := "from:noreply@steampowered.com is:unread"
	messages, err := srv.Users.Messages.List(user).Q(query).Do()
	if err != nil {
		log.Fatal(err)
	}
	if len(messages.Messages) == 0 {
		return ""
	}
	msg, err := srv.Users.Messages.Get(user, messages.Messages[0].Id).Do()
	if err != nil {
		log.Fatal(err)
	}
	for _, part := range msg.Payload.Parts {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			log.Fatal(err)
		}
		regex := regexp.MustCompile(`Login Code\s*([A-Za-z0-9]+)`)
		match := regex.FindStringSubmatch(string(data))
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func checkForSteamGuardInLogs(stdout, stderr *bytes.Buffer) bool {
	return strings.Contains(stdout.String(), "This computer has not been authenticated for your account using Steam Guard") || strings.Contains(stderr.String(), "This computer has not been authenticated for your account using Steam Guard")
}

func main() {
	config := readConfig()
	srv, err := getClient()
	if err != nil {
		log.Fatalf("Unable to create Gmail client: %v", err)
	}

	args := os.Args[1:]

	// Check if +quit is already present and remove it if it is as we add it later
	quitIndex := -1
	for i, arg := range args {
		if arg == "+quit" {
			quitIndex = i
			break
		}
	}
	if quitIndex != -1 {
		args = append(args[:quitIndex], args[quitIndex+1:]...)
	}

	for {
		var stderr, stdout bytes.Buffer // Buffers to capture stderr and stdout
		cmd := exec.Command(config.SteamCmdPath)
		cmd.Args = append(cmd.Args, args...)
		cmd.Args = append(cmd.Args, "+quit") // Add +quit to the end of the arguments to make sure SteamCmd quits after running the given arguments, must always be the very last argument

		fmt.Println("Running SteamCmd with given arguments: ", cmd.Args)

		// Use MultiWriter to write to both stdout and the buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

		err := cmd.Run()

		if err != nil {
			if checkForSteamGuardInLogs(&stdout, &stderr) {
				fmt.Println("Steam Guard code required. Fetching...")
				for i := 0; i < 12; i++ {
					code := fetchSteamGuardCode(srv)
					if code != "" {
						fmt.Printf("Fetched Steam Guard code: %s\n", code)

						// Check if +set_steam_guard_code is already present
						steamGuardIndex := -1
						for i, arg := range args {
							if arg == "+set_steam_guard_code" {
								steamGuardIndex = i
							}
						}

						// Append the Steam Guard code to the arguments or update it
						if steamGuardIndex != -1 {
							args[steamGuardIndex+1] = code
						} else {
							args = append(args, "+set_steam_guard_code", code)
						}

						break
					} else {
						fmt.Println("No Steam Guard code found. Waiting 10 seconds...")
					}

					time.Sleep(10 * time.Second)
				}
			} else {
				fmt.Println("Error running SteamCmd:", err)
				return
			}
		} else {
			fmt.Println("Successfully run SteamCmd with given arguments.")
			return
		}
	}
}
