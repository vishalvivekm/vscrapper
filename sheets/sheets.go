package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"vishalvivekm/vcrawler/models"
)

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "./sheets/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}


func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Head to following link and then type the "+
		"authorization code: \n%v\n", authURL) // does not work 
		// fix: after allowing, copy the code form thr url after code=<code> and paste that into terminal

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func SaveToSheets(ambassadorData map[string]models.AmbassadorDetail) {
	if envErr := godotenv.Load(".env"); envErr != nil {
        fmt.Println("can not load .env file")
    }
	ctx := context.Background()
	b, err := os.ReadFile("./sheets/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file ( credentials.json): %v", err)
	}

	// If modifying these scopes, rm previously saved token.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets") // readonly scope: https://www.googleapis.com/auth/spreadsheets
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}


spreadsheetId := os.Getenv("SPREADSHEET_ID")
sheetId := 0

response1, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
if err != nil || response1.HTTPStatusCode != 200 {
	log.Fatal(err)
	return
}
sheetName := ""
for _, v := range response1.Sheets {
	prop := v.Properties
	if prop.SheetId == int64(sheetId) {
		sheetName = prop.Title
		break
	}
}

	var values [][]interface{}
	values = append(values, []interface{}{"Name", "LinkedIn URL", "GitHub URL", "Twitter URL"}) // header row
	
	for _, ambassador := range ambassadorData {
		values = append(values, []interface{}{
			ambassador.Name,
			ambassador.LinkedInURL,
			ambassador.GitHubURL,
			ambassador.TwitterUrl,
		})
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

response2, err := srv.Spreadsheets.Values.Append(spreadsheetId, sheetName, valueRange).
ValueInputOption("USER_ENTERED").
InsertDataOption("INSERT_ROWS").
Context(ctx).Do()
if err != nil || response2.HTTPStatusCode != 200 {
	log.Fatal(err)
	return
}
fmt.Printf("Data appended to range: %v\n", response2.Updates.UpdatedRange)
}
