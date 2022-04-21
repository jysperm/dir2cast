package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	id3 "github.com/mikkyang/id3-go"
	id3v2 "github.com/mikkyang/id3-go/v2"
)

//go:embed feed.tmpl
var feedTmpl string

type FeedInfo struct {
	Title    string
	BaseURL  string
	Episodes []Episode
}

type Episode struct {
	Title       string
	PubDate     string
	Description string
	FileName    string
	ArtworkData []byte
}

func main() {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	files, err := ioutil.ReadDir(".")

	if err != nil {
		panic(err)
	}

	feedInfo := FeedInfo{
		Title: path.Base(cwd),
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".mp3") {
			fmt.Printf("[Parse] Processing %s\n", file.Name())
			feedInfo.Episodes = append(feedInfo.Episodes, parseEpisode(file))
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[HTTP] Accessed by %s %s\n", r.RemoteAddr, r.RequestURI)

		feedInfo.BaseURL = fmt.Sprintf("http://%s", r.Host)

		feed := bytes.Buffer{}

		err = template.Must(template.New("feed.xml").Parse(feedTmpl)).Execute(&feed, feedInfo)

		if err != nil {
			panic(err)
		}

		if r.URL.Path == "/" || r.URL.Path == "/feed.xml" {
			w.Write(feed.Bytes())
		} else if strings.HasPrefix(r.URL.Path, "/download") {
			filename := r.URL.Path[len("/download/"):]
			episode := findEpisodes(&feedInfo, filename)
			http.ServeFile(w, r, episode.FileName)
		} else if r.URL.Path == "/cover.jpg" {
			w.Header().Add("Content-Type", "image/jpeg")
			w.Write(feedInfo.Episodes[0].ArtworkData)
		}
	})

	port := 8080

	fmt.Printf("[HTTP] Listening on port %d ...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func findEpisodes(feedInfo *FeedInfo, filename string) *Episode {
	for _, episode := range feedInfo.Episodes {
		if episode.FileName == filename {
			return &episode
		}
	}

	return nil
}

func parseEpisode(file fs.FileInfo) Episode {
	tag, err := id3.Open(file.Name())

	if err != nil {
		panic(err)
	}

	defer tag.Close()

	dataFrameULT := tag.Frame("ULT").(*id3v2.DataFrame)
	dataFramePIC := tag.Frame("PIC").(*id3v2.DataFrame)

	frameULT := id3v2.ParseUnsynchTextFrame(dataFrameULT.FrameHead, dataFrameULT.Bytes()).(*id3v2.UnsynchTextFrame)
	framePIC := id3v2.ParseImageFrame(dataFramePIC.FrameHead, dataFramePIC.Bytes()).(*id3v2.ImageFrame)

	return Episode{
		Title:       strings.TrimSuffix(file.Name(), ".mp3"),
		PubDate:     file.ModTime().Format(time.RFC1123Z),
		Description: strings.Replace(frameULT.Text(), "\x00", "", -1),
		FileName:    file.Name(),
		ArtworkData: append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00}, framePIC.Data()...),
	}
}