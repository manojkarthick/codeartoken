package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Configuration stores Maven and CodeArtifact related information
type Configuration struct {
	domain      string // AWS CodeArtifact domain
	domainOwner string // AWS CodeArtifact domain owner
	server      string // Server ID for AWS CodeArtifact in your Maven settings
	settings    string // AWS CodeArtifact Maven settings path
}

// Get the default location for the maven settings file
func getDefaultMavenSettings() string {
	log.Debug("Parsing settings.xml")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not get user home directory")
		os.Exit(1)
	}

	settingsPath := filepath.Join(homeDir, ".m2", "settings.xml")
	return settingsPath
}

// Get CodeArtifact Authentication Token using AWS Credentials
func getCodeArtifactToken(cfg Configuration) (string, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	client := codeartifact.NewFromConfig(awsConfig)

	output, err := client.GetAuthorizationToken(context.TODO(), &codeartifact.GetAuthorizationTokenInput{
		Domain:      &cfg.domain,
		DomainOwner: &cfg.domainOwner,
	})

	token := aws.ToString(output.AuthorizationToken)
	log.Debugf("CodeArtifact Token: %s", token)
	return token, nil
}

// Get currently used CodeArtifact Token present in the maven settings XML file
func getCodeArtifactTokenFromSettings(cfg Configuration) (string, error) {
	xmlFile, err := os.Open(cfg.settings)
	if err != nil {
		return "", err
	}

	doc, err := xmlquery.Parse(xmlFile)
	if err != nil {
		return "", err
	}

	nodes := xmlquery.Find(doc, "/settings/servers/server")
	for _, node := range nodes {
		element := node.SelectElement("id")
		if element.InnerText() == cfg.server {
			password := node.SelectElement("password").InnerText()
			log.Debugf("CodeArtifact Maven Password: %s", password)
			return password, nil

		}
	}

	defer xmlFile.Close()
	return "", errors.New(fmt.Sprintf("Could not find server %s", cfg.server))
}

// Update Maven settings with the new CodeArtifact token
func updateSettings(cfg Configuration, existingToken string, newToken string) {
	input, err := ioutil.ReadFile(cfg.settings)
	if err != nil {
		log.Fatal(err)
	}

	output := bytes.ReplaceAll(input, []byte(existingToken), []byte(newToken))
	if err = ioutil.WriteFile(cfg.settings, output, 0644); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Info("✅ Updated token!")
}

// Main entrypoint for managing the CodeArtifact tokens for Maven
func manageTokens(cfg Configuration) error {
	log.Infof("Using owner: %s", cfg.domain)
	log.Infof("Using domain owner: %s", cfg.domainOwner)
	log.Infof("Using server: %s", cfg.server)

	existingToken, err := getCodeArtifactTokenFromSettings(cfg)
	if existingToken == "" || err != nil {
		return err
	}

	newToken, err := getCodeArtifactToken(cfg)
	if newToken == "" || err != nil {
		return err
	}

	updateSettings(cfg, existingToken, newToken)
	return nil
}

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	cfg := Configuration{}

	app := &cli.App{
		Name:  "codeartoken",
		Usage: "Refresh AWS CodeArtifact token for maven",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "domain",
				Aliases:     []string{"d"},
				Destination: &cfg.domain,
			},
			&cli.StringFlag{
				Name:        "owner",
				Aliases:     []string{"o"},
				Destination: &cfg.domainOwner,
			},
			&cli.StringFlag{
				Name:        "server",
				Aliases:     []string{"s"},
				Destination: &cfg.server,
				Value:       "codeartifact",
			},
			&cli.StringFlag{
				Name:        "settings",
				Aliases:     []string{"x"},
				Destination: &cfg.settings,
				Value:       getDefaultMavenSettings(),
			},
		},
		Before: func(context *cli.Context) error {
			return nil
		},
		Action: func(c *cli.Context) error {
			if c.NumFlags() == 0 {
				cli.ShowAppHelpAndExit(c, 0)
			}

			err := manageTokens(cfg)
			if err != nil {
				log.Fatal("Encountered exception, exiting: ", err)
				os.Exit(1)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Encountered error: %s", err.Error())
		log.Fatal("Exiting...")
		os.Exit(1)
	}
}
