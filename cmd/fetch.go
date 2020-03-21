package cmd

import (
	"time"
	"sync"
	"net/url"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kondoumh/scrapbox-viz/pkg/file"

	"github.com/kondoumh/scrapbox-viz/pkg"
	"github.com/kondoumh/scrapbox-viz/pkg/fetch"
	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch all pages of the project",
	Long:  `fetch all pages of the project`,
	Run: func(cmd *cobra.Command, args []string) {
		doFetch(cmd)
	},
}

func init() {
	fetchCmd.PersistentFlags().StringP("project", "p", "help-jp", "Name of Scrapbox project (required)")
	rootCmd.AddCommand(fetchCmd)
}

func doFetch(cmd *cobra.Command) {
	projectName, _ := cmd.PersistentFlags().GetString("project")
	project, err := fetchIndex(projectName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("fetch all pages, %s : %d\n", project.Name, project.Count)
	err = fetchPageList(project)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	groups, err := dividePagesList(3, projectName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	path := fmt.Sprintf("%s/%s", config.WorkDir, projectName)
	file.CreateDir(path)
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(len(groups))
	for _, pages := range groups {
		go fetchPagesByGroup(projectName, pages, &wg)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("took %s\n", elapsed)
}

func fetchIndex(projectName string) (pkg.Project, error) {
	url := fmt.Sprintf("%s/%s?limit=1", fetch.BaseURL, projectName)
	data, err := fetch.FetchData(url)
	var project pkg.Project
	if err != nil {
		return project, err
	}
	err = json.Unmarshal(data, &project)
	if err != nil {
		return project, err
	}
	return project, nil
}

func fetchPageList(project pkg.Project) error {
	pages := []pkg.Page{}
	for skip := 0; skip < project.Count; skip += fetch.Limit {
		url := fmt.Sprintf("%s/%s?skip=%d&limit=%d&sort=updated", fetch.BaseURL, project.Name, skip, fetch.Limit)
		data, err := fetch.FetchData(url)
		if err != nil {
			return err
		}
		var proj pkg.Project
		err = json.Unmarshal(data, &proj)
		for _, page := range proj.Pages {
			pages = append(pages, page)
		}
	}
	project.Pages = pages
	data, _ := json.Marshal(project)
	if err := file.WriteBytes(data, project.Name+".json", config.WorkDir); err != nil {
		return err
	}
	return nil
}

func dividePagesList(multiplicity int, projectName string) ([][]pkg.Page, error) {
	var divided [][]pkg.Page
	proj, err := readProject(projectName)
	if err != nil {
		return divided, err
	}
	fmt.Printf("Total pages : %d\n", len(proj.Pages))
	chunkSize := len(proj.Pages) / multiplicity
	fmt.Printf("Chunk size : %d\n", chunkSize)
	for i := 0; i < len(proj.Pages); i += chunkSize {
		end := i + chunkSize
		if end > len(proj.Pages) {
			end = len(proj.Pages)
		}
		divided = append(divided, proj.Pages[i:end])
	}
	totalCount := 0
	for _, pages := range divided {
		totalCount += len(pages)
		fmt.Printf("Size of chunk %d\n", len(pages))
	}
	fmt.Printf("Total pages to be fetched %d\n", totalCount)
	return divided, nil
}

func readProject(projectName string) (pkg.Project, error) {
	bytes, err := file.ReadBytes(projectName+".json", config.WorkDir)
	var project pkg.Project
	if err != nil {
		return project, err
	}
	if err := json.Unmarshal(bytes, &project); err != nil {
		return project, err
	}
	return project, err
}

func fetchPagesByGroup(projectName string, pages []pkg.Page, wg *sync.WaitGroup) error {
	defer wg.Done()
	for _, page := range pages {
		fmt.Println(page.Title)
		if err := fetchPage(projectName, page.Title, page.ID); err != nil {
			return err
		}
	}
	return nil
}

func fetchPage(projectName string, title string, index string) error {
	url := fmt.Sprintf("%s/%s/%s", fetch.BaseURL, projectName, url.PathEscape(title))
	data, err := fetch.FetchData(url)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s.json", index)
	path := fmt.Sprintf("%s/%s", config.WorkDir, projectName)
	if err := file.WriteBytes(data, fileName, path); err != nil {
		return err
	}
	return nil
}
