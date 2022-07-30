package main

import (
	"automation/internal/config"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"golang.design/x/clipboard"
)

const (
	jenkinsURL = "https://jenkins-mgmt-sys-dev-ew1.pmc.pearsondev.tech/job/DevOps/job/checkout-reactjs-gold/job/reimagined-he/"
)

var conf config.Config

func init() {
	var err error
	conf, err = config.LoadConfig("./env/.env")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	envToDeploy := flag.String("env", "dev2", "Environment to deploy")
	branchToDeploy := flag.String("branch", "develop", "Branch name to deploy")

	flag.Parse()

	browser := rod.New().MustConnect().NoDefaultDevice()
	page := browser.MustPage(jenkinsURL).MustWindowFullscreen()

	authenticate(page)

	repoNames := []string{"pmc-react-browse", "pmc-react-login", "pmc-react-myaccount", "pmc-react-replocator"}

	for _, repo := range repoNames {
		// visit job's page
		page.MustElement(fmt.Sprintf("a[href='job/%s/']", repo)).MustClick()

		// click on build link
		page.MustElement("a[href*='/build?']").MustClick()

		// select env
		dropdown, _ := page.MustElement("input[value='ENV_TYPE']").Next()
		dropdown.MustSelect(*envToDeploy)

		// choose branch to deploy
		input, _ := page.MustElement("input[value='BRANCH']").Next()
		input.MustSelectAllText().MustInput(*branchToDeploy)

		// submit
		page.MustElement("button[type='submit']").MustClick()

		// wait for pipeline to appear
		page.MustElement("#pipeline-box")

		// go to the list
		page.MustNavigate(jenkinsURL)
	}

	alertMessage := createAlertMessage(*envToDeploy)
	clipboard.Write(clipboard.FmtText, []byte(alertMessage))
	fmt.Println("Warning info copied to clipboard!")

	time.Sleep(time.Minute * 10)
}

func authenticate(page *rod.Page) {
	// fill auth credentials
	page.MustElement("#j_username").MustInput(conf.JenkinsUsername)
	page.MustElement("input[name=j_password]").MustInput(conf.JenkinsPassword)

	// submit
	page.MustElement("form button[type=submit]").MustClick()
}

func createAlertMessage(env string) (msg string) {
	msg = "(paloudspeaker) React deployment alert\n \n"
	msg += fmt.Sprintf("ENV: %s\n", env)
	msg += "Status: In progress (hourglassdone)"

	return
}
