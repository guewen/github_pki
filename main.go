package main

import (
  "os"
  "os/exec"
  "io/ioutil"
  "strings"
  "fmt"
  "golang.org/x/oauth2"
  "github.com/google/go-github/github"
	"github.com/Sirupsen/logrus"
)

func main() {
  gh_token := os.Getenv("GITHUB_TOKEN")
  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: gh_token},
  )
  tc := oauth2.NewClient(oauth2.NoContext, ts)

  client := github.NewClient(tc)

  // Get users from teams
  users, err := getTeamUsers(client)
  if err != nil {
    logrus.Errorf("Failed to get team users")
  }

  users, err = getUsers(client, users)
  if err != nil {
    logrus.Errorf("Failed to add individual users")
  }

  //getUserKeys(client, users)
}

func getTeamUsers(client *github.Client) ([]github.User, error) {
  var users []github.User

  gh_org := os.Getenv("GITHUB_ORG")
  gh_team := os.Getenv("GITHUB_TEAM")

  if gh_org == "" {
    return users, nil
  }

  gh_teams, _, err := client.Organizations.ListTeams(gh_org, nil)
  if err != nil {
    logrus.Errorf("Failed to list teams: %v", err)
  }

  for _, team := range gh_teams {
    gh_users, _, err := client.Organizations.ListTeamMembers(*team.ID, nil)
    if err != nil {
      logrus.Errorf("Failed to list team members")
    }

    if gh_team == "" || *team.Name == gh_team {
      logrus.Infof("Adding users for team %v", *team.Name)
      for _, user := range gh_users {
        users = append(users, user)
      }
    }
  }

  return users, err
}

func getUsers(client *github.Client, users []github.User) ([]github.User, error) {
  var err error

  for _, u := range strings.Split(os.Getenv("GITHUB_USERS"), ",") {
    user, _, err := client.Users.Get(u)
    if err != nil {
      logrus.Errorf("Failed to find user %v", u)
    }
    users = append(users, *user)
  }

  return users, err
}

func getUserKeys(client *github.Client, users []github.User) (error) {
  var authorizedKeys []string
  var err error

  for _, user := range users {
    keys, _, err := client.Users.ListKeys(*user.Login, nil)
    if err != nil {
      logrus.Errorf("Failed to list keys for user %v", *user.Login)
    }

    for _, k := range keys {
      // Add to authorized_keys
      authorizedKeys = append(authorizedKeys, *k.Key)

      // And/or dump SSL key
      tmpfile, err := ioutil.TempFile("", "ssh-ssl")
      if err != nil {
        logrus.Errorf("Failed to create tempfile")
      }
      defer os.Remove(tmpfile.Name())

      cmd := exec.Command("ssh-keygen", "-f", tmpfile.Name(), "-e", "-m", "pem")

      // TODO: split stdout/stderr in case of errors
      ssl_key, err := cmd.CombinedOutput()
      if err != nil {
        logrus.Errorf("Failed to convert key to X509")
      }

      ssl_dir := "/tmp"
      ssl_keyfile := fmt.Sprintf("%s/%s.pem", ssl_dir, *k.ID)

      err = ioutil.WriteFile(ssl_keyfile, ssl_key, 0644)
      if err != nil {
        logrus.Errorf("Failed to write key to file")
      }
    }
  }
  return err
}

func dumpAuthorizedKeys() () {
}


