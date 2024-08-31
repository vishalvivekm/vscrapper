package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	//"os"
)
var (
ambassadorsData = make(map[string]AmbassadorDetail) 
inPersonBlock bool
divDepth int
currentAmbassadorName string
)

type AmbassadorDetail struct {
	Name, LinkedInURL, GitHubURL, TwitterUrl  string
}

var socialMap = map[string]string{
	"GitHub":   "github.com",
	"LinkedIn": "linkedin.com",
	"Twitter":  "twitter.com",
}

const url = "https://www.cncf.io/people/ambassadors/"

func main() {
	pageContent, err := GetPageContent(url)
	if err != nil {
		log.Fatal(err)
	}

	isNextLineName := false
	re := regexp.MustCompile(`<a.*?href=["'](.*?)["']`)
	for _, line := range strings.Split(string(pageContent), "\n") {
		if strings.Contains(line, "<div") && inPersonBlock{
			divDepth++
		}
		if strings.Contains(line, "<div class=\"person has-animation-scale-2\">") {
			inPersonBlock = true
			divDepth++
		} else if strings.Contains(line, "</div>") && inPersonBlock {
			divDepth--
			if divDepth == 0 {
				inPersonBlock = false
				currentAmbassadorName = ""
			}
		}


		if isNextLineName {
			nameSlice := strings.Split(strings.TrimSpace(line), " ")[:2]
			name := strings.Join(nameSlice, " ")
			currentAmbassadorName = name
			isNextLineName = false
		}
		if strings.Contains(line, `<h3 class="person__name">`) {
			isNextLineName = true
		}

		personn := AmbassadorDetail{}
		if currentAmbassadorName != "" {
			person, exists := ambassadorsData[currentAmbassadorName]
			if !exists {
				person = AmbassadorDetail{Name: currentAmbassadorName}
				personn = person
			}
			personn = person
		}

		if inPersonBlock {
			match := re.FindStringSubmatch(line)
			if match != nil {
				link := match[1]
				// slice := ""
				// for i, val := range match{
				// 	s := fmt.Sprintf("%v : %v", i, val)
				//     slice += "\n " + s
				// 	fmt.Println(s)
					
				// }
				// os.WriteFile("match.text", []byte(slice), 0644)


				switch {
				case strings.Contains(link, socialMap["LinkedIn"]):
					//fmt.Println(link)
					personn.LinkedInURL = link
				case strings.Contains(link, socialMap["Twitter"]):
					personn.TwitterUrl = link
				case strings.Contains(link, socialMap["GitHub"]):
					personn.GitHubURL = link

				}
				ambassadorsData[currentAmbassadorName] = personn
			}
		}
		// _, exists := ambData[currentPersonName]
		// if !exists {
		// 	ambData[currentPersonName] = personn
		// }
	}

	// fmt.Printf("%+v", ambData)
	file, err := os.OpenFile("data.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Println(len(ambassadorsData))
	// _, err = fmt.Fprint(file, ambData)
	for key, val := range ambassadorsData {
		fmt.Println(key, "   ", val)
	_, err = fmt.Fprintf(file, "\n%v: %v %v %v",key, val.LinkedInURL, val.GitHubURL, val.TwitterUrl)
	if err != nil {
		log.Fatal(err)
	}
	}

}

func GetPageContent(url string) (body []byte, err error) {
	filename := "amb.html"
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// page's html file doesn not not exist; get page content
			body, err := fetchBody(url)
			if err != nil {
				return nil, err
			}
			 //_ = os.WriteFile(filename, body, 0644) // write file to disk
			return body, nil
		} else {
			return nil, fmt.Errorf("unknown error during stating: %w", err)
		}
	}
	log.Printf("amb file %v exists\nReading file...\n", file.Name())
	body, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	return
}
func fetchBody(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error forming request: %w", err)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting resp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("req not successful, non 200 statusCode: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}
