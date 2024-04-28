package connector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

var ImageToSign string
var Repo string
var PushPermission string

func SetImageToSign(img string) {
	ImageToSign = img
}

func GetImageToSign() (string, bool) {
	if ImageToSign != "" {
		fmt.Println("server.ImageToSign gotten is: ", ImageToSign)
		return ImageToSign, true
	}
	return "", false
}

func SetRepoAndPermission(repo string, permission string) {
	Repo = repo
	PushPermission = permission
}

func GetRepoAndPermission() (string, string) {
	if Repo != "" && PushPermission != "" {
		return Repo, PushPermission
	}
	return "", ""
}

type ErrorResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

type UserPermissions struct {
	Push bool `json:"push"`
}

func GetPermission(input string, username string, token string) (string, bool, error) {

	parts := strings.Split(input, "/")
	if len(parts) >= 4 {
		repo_owner := parts[1]
		reponame := parts[2]
		pathname := fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators/%s/permission", repo_owner, reponame, username)

		curlCmd := exec.Command("curl", "-L",
			"-H", "Accept: application/vnd.github+json",
			"-H", fmt.Sprintf("Authorization: Bearer %s", token),
			"-H", "X-GitHub-Api-Version: 2022-11-28",
			pathname)

		output, err := curlCmd.Output()
		if err != nil {
			return "", false, fmt.Errorf("error executing curl command: %v", err)
		}
		fmt.Println(len(output))

		if strings.Contains(string(output), "is not a user") {
			var errResponse ErrorResponse
			if err := json.Unmarshal(output, &errResponse); err != nil {
				return "", false, fmt.Errorf("error parsing JSON: %v", err)
			}
			return "", false, fmt.Errorf("%s. Documentation URL: %s", errResponse.Message, errResponse.DocumentationURL)
		}
		var permResponse struct {
			User struct {
				Permissions UserPermissions `json:"permissions"`
			} `json:"user"`
		}
		// Unmarshal the JSON into the struct
		if err := json.Unmarshal([]byte(output), &permResponse); err != nil {
			fmt.Println("Error:", err)
			return "", false, fmt.Errorf("error parsing Permission Response JSON: %v", err)
		}
		hostRepo := fmt.Sprintf("github.com/%s/%s", repo_owner, reponame)
		pushPermission := permResponse.User.Permissions.Push
		return hostRepo, pushPermission, nil

	} else {
		return "", false, fmt.Errorf("invalid image repo path format")
	}

}
