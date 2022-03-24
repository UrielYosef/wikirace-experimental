package main

import (
	"crawler/tree"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	baseUrl                    = "https://en.wikipedia.org/wiki/"
	maxDepth                   = 3
	maxChildren                = 5
	sleepTimeBetweenRequests   = time.Second * 30
	debugWikiNamesToFileSystem = false
)

var unwantedWikiPrefixes = [...]string{"Wikipedia:", "Category:", "File:",
	"Portal:", "Template:", "CS1_maint:", "Special:", "Template_talk:", "Help:", "Talk:"}

var depths = make([]bool, maxDepth)

var allNames = make(map[string]bool)
var lock = sync.RWMutex{}

//TODO: http client singleton
func main() {
	srcWikiPageName, destWikiPageName := getSourceAndDestination()
	fmt.Printf("Searching for route from Source: \"%v\" to Destination: \"%v\"...\n", srcWikiPageName, destWikiPageName)

	start := time.Now()
	root, searchJobsQueue := initSearchStructure(srcWikiPageName)
	destNode := executeBreadthFirstSearchJob(searchJobsQueue, destWikiPageName)
	elapsed := time.Since(start)
	if destNode != nil {
		fmt.Printf("Destination \"%v\" was found\n", destWikiPageName)
		destNode.PrintRouteToRoot()
		//fmt.Println("Depth was", depth)
	} else {
		fmt.Printf("Destination \"%v\" was not found within max depth of %v\n", destWikiPageName, maxDepth)
	}
	fmt.Println("Tree depth is", root.Depth())
	fmt.Println("Execution time was", elapsed)
}

func getSourceAndDestination() (string, string) {
	var srcWikiPageName, destWikiPageName string
	if len(os.Args) == 3 {
		return os.Args[1], os.Args[2]
	}
	fmt.Printf("Please Enter Source wiki page name: ")
	fmt.Scan(&srcWikiPageName)
	fmt.Printf("Please Enter Dest wiki page name: ")
	fmt.Scan(&destWikiPageName)
	strings.TrimSpace(srcWikiPageName)
	strings.TrimSpace(destWikiPageName)
	// srcWikiPageName := "Go_(programming_language)"
	// destWikiPageName := "Touchscreen"

	return srcWikiPageName, destWikiPageName
}

func initSearchStructure(src string) (*tree.Node, []*tree.Node) {
	var root = tree.NewTree(src)
	fmt.Println("Initializing first depth based on source name...")
	fillNode(root)

	var searchJobsQueue = make([]*tree.Node, 0)
	searchJobsQueue = append(searchJobsQueue, root)

	return root, searchJobsQueue
}

func executeBreadthFirstSearchJob(queue []*tree.Node, dest string) *tree.Node {
	if len(queue) == 0 {
		return nil
	}

	node := queue[0]
	if node.Level >= maxDepth {
		return nil
	}

	printCurrentSearchingDepth(node.Level + 1)
	destNode := searchInChildren(node, dest)
	if destNode != nil {
		return destNode
	}

	time.Sleep(sleepTimeBetweenRequests)
	fillNodeChildren(node)
	queue = append(queue, *node.Children...)

	return executeBreadthFirstSearchJob(queue[1:], dest)
}

func fillNodeChildren(node *tree.Node) {
	var wg sync.WaitGroup
	for _, child := range *node.Children {
		wg.Add(1)
		go func(child *tree.Node) {
			defer wg.Done()
			fillNode(child)
		}(child)
	}
	wg.Wait()
}

func fillNode(node *tree.Node) {
	content := getWikiContent(node.Name)
	wikiPageNames := getRelatedWikiPageNames(content)
	filteredNames := filterExistingNames(wikiPageNames)
	node.Insert(filteredNames)

	debugToFileSystem(node)
}

func printCurrentSearchingDepth(currentDepth int) {
	if depths[currentDepth-1] == false {
		fmt.Printf("Searching in depth %v...\n", currentDepth)
		depths[currentDepth-1] = true
	}
}

func searchInChildren(node *tree.Node, dest string) *tree.Node {
	for _, node := range *node.Children {
		if node.Name == dest {
			return node
		}
	}

	return nil
}

func getWikiContent(wikiPageName string) string {
	response, err := http.Get(baseUrl + wikiPageName)
	checkNilErr(err)
	defer response.Body.Close()

	dataBytes, err := ioutil.ReadAll(response.Body)
	checkNilErr(err)

	content := string(dataBytes)

	return content
}

func getRelatedWikiPageNames(content string) []string {
	wikiPageNames := make([]string, 0)
	doc, err := html.Parse(strings.NewReader(content))
	checkNilErr(err)
	var f func(*html.Node, *[]string)
	f = func(n *html.Node, wikiPageNames *[]string) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && strings.HasPrefix(a.Val, "/wiki/") {
					wikiPageName := strings.Trim(a.Val, "/wiki/")
					if !hasPrefix(unwantedWikiPrefixes[:], wikiPageName) {
						*wikiPageNames = append(*wikiPageNames, wikiPageName)
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, wikiPageNames)
		}
	}
	f(doc, &wikiPageNames)

	return wikiPageNames
}

func filterExistingNames(newNames []string) []string {
	lock.Lock()
	var filteredNames = make([]string, 0)
	for _, name := range newNames {
		isExists := allNames[name]
		if !isExists {
			filteredNames = append(filteredNames, name)
			allNames[name] = true
		}
	}
	lock.Unlock()

	if len(filteredNames) < maxChildren {
		return filteredNames
	}
	return filteredNames[:maxChildren]
}

func hasPrefix(names []string, str string) bool {
	for _, name := range names {
		if strings.HasPrefix(str, name) {
			return true
		}
	}
	return false
}

func debugToFileSystem(node *tree.Node) {
	if debugWikiNamesToFileSystem {
		var fileName string
		if node.Parent == nil {
			fileName = strconv.Itoa(node.Level) + "-" + node.Name + ".txt"
		} else {
			fileName = strconv.Itoa(node.Level) + "-" + node.Name + "-" + node.Parent.Name + ".txt"
		}
		_, err := os.Create(fileName)
		if err != nil {
			fmt.Println("error creating", fileName)
		}
	}
}

func checkNilErr(err error) {
	if err != nil {
		panic(err)
	}
}
