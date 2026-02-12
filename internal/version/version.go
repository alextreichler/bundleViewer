package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	// Current is the version of the binary, set at build time via ldflags.
	Current = "dev"
	
	// Repo is the GitHub repository name.
	Repo = "alextreichler/bundleViewer"
)

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
}

// CheckUpdate checks if a new version is available on GitHub.
// It returns the latest version tag if an update is available, or an empty string otherwise.
func CheckUpdate() (string, error) {
	if Current == "dev" {
		return "", nil
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Current, "v")

	if isNewer(latest, current) {
		return release.TagName, nil
	}

	return "", nil
}

// isNewer is a simple version comparison.
// It assumes semantic versioning without pre-release tags for simplicity.
func isNewer(latest, current string) bool {
	if latest == current {
		return false
	}
	
	lParts := strings.Split(latest, ".")
	cParts := strings.Split(current, ".")
	
	for i := 0; i < len(lParts) && i < len(cParts); i++ {
		lNum, errL := strconv.Atoi(lParts[i])
		cNum, errC := strconv.Atoi(cParts[i])
		
		if errL == nil && errC == nil {
			if lNum > cNum {
				return true
			}
			if lNum < cNum {
				return false
			}
		} else {
			// Fallback to string comparison if not numeric
			if lParts[i] > cParts[i] {
				return true
			}
			if lParts[i] < cParts[i] {
				return false
			}
		}
	}
	
	return len(lParts) > len(cParts)
}

// UpdateCommand returns the command to update the application.
func UpdateCommand() string {
	return fmt.Sprintf("curl -sSL https://raw.githubusercontent.com/%s/main/install.sh | bash", Repo)
}
