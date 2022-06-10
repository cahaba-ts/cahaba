package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/adrg/frontmatter"
	"github.com/cahaba-ts/epub"
	"github.com/urfave/cli"
)

type BuildConfig struct {
	Title       string   `toml:"Title" yaml:"title"`
	Author      string   `toml:"Author" yaml:"author"`
	Cover       string   `toml:"Cover" yaml:"cover"`
	Description string   `toml:"Description" yaml:"description"`
	ImageFolder string   `toml:"ImageFolder" yaml:"image_folder"`
	Sections    []string `toml:"Sections" yaml:"sections"`
}

var onlyLN = false

func Build(c *cli.Context) error {
	args := c.Args()
	vpath, _ := os.Getwd()
	if len(args) == 0 {
		fmt.Println("Building Volume in Current Directory")
	} else {
		folder := strings.Join(args, " ")
		os.Mkdir(folder, os.ModeDir|os.ModePerm)
		fmt.Printf("Building Volume in '%s'\n", folder)
		vpath = filepath.Join(vpath, folder)
	}

	configureShortcodes()

	config := loadConfig(vpath)
	err := buildBookEpub(vpath, config)
	if err != nil {
		return err
	}
	config.Title += " (LN Images Only)"
	onlyLN = true

	return buildBookEpub(vpath, config)
}

func buildBookEpub(vpath string, config BuildConfig) error {
	output := configureEpub(vpath, config)
	pastPrologue := false
	for _, section := range config.Sections {
		cfg, body := loadChapter(vpath, section)
		if strings.HasPrefix(section, "chapter") {
			pastPrologue = true
			output.AddChapterMD(cfg.Title, body)
		} else if pastPrologue {
			output.AddPostscriptMD(cfg.Title, body)
		} else {
			output.AddIntroductionMD(cfg.Title, body)
		}
	}
	return output.Write(config.Title + ".epub")
}

func loadConfig(vpath string) BuildConfig {
	f, err := os.Open(filepath.Join(vpath, "volume.toml"))
	if err != nil {
		log.Fatal("Cannot open config: ", err)
	}
	c := BuildConfig{}
	toml.NewDecoder(f).Decode(&c)
	return c
}

func loadChapter(vpath, chapterFile string) (BuildConfig, string) {
	f, err := os.Open(filepath.Join(vpath, "text", chapterFile+".md"))
	if err != nil {
		log.Fatal("Open Chapter File: ", chapterFile, err)
	}
	bc := BuildConfig{}
	body, err := frontmatter.Parse(f, &bc)
	if err != nil {
		log.Fatal("Parse Chapter File: ", chapterFile, err)
	}
	return bc, string(body)

}

func configureEpub(vpath string, config BuildConfig) *epub.Book {
	book := epub.NewBook(config.Title)
	if config.Author != "" {
		book.SetAuthor(config.Author)
	}
	book.SetDescription(config.Description)
	book.AddImageFolder(filepath.Join(vpath, config.ImageFolder))
	book.SetCover(config.Cover)
	book.SetReleaseDate(time.Now().Format("2006-01-02"))
	book.SetPublisher("Cahaba Translations")
	if _, err := os.Stat("main.css"); err == nil {
		book.SetCSS("main.css")
	}
	return book
}

func configureShortcodes() {
	epub.RegisterShortcode(
		"normalimage",
		func(b *epub.Book, s1 string, m map[string]any, s2 string) (string, error) {
			if onlyLN {
				return "", nil
			}
			img := fmt.Sprint(m["image"])
			alt := fmt.Sprint(m["alt"])
			imgPath, ok := b.LookupImage(img)
			if !ok {
				return "", errors.New("Missing Image")
			}
			return fmt.Sprintf(
				`<img class="cahaba--normal-image" src="%s" alt="%s" />`,
				imgPath,
				alt,
			), nil
		},
	)
	epub.RegisterShortcode(
		"narrowimage",
		func(b *epub.Book, s1 string, m map[string]any, s2 string) (string, error) {
			if onlyLN {
				return "", nil
			}
			img := fmt.Sprint(m["image"])
			alt := fmt.Sprint(m["alt"])
			imgPath, ok := b.LookupImage(img)
			if !ok {
				return "", errors.New("Missing Image")
			}
			return fmt.Sprintf(
				`<img class="cahaba--narrow-image" src="%s" alt="%s" />`,
				imgPath,
				alt,
			), nil
		},
	)
	epub.RegisterShortcode(
		"fullimage",
		func(b *epub.Book, s1 string, m map[string]any, s2 string) (string, error) {
			img := fmt.Sprint(m["image"])
			alt := fmt.Sprint(m["alt"])
			imgPath, ok := b.LookupImage(img)
			if !ok {
				return "", errors.New("Missing Image")
			}
			return fmt.Sprintf(
				`<!-- PAGE BREAK --><img class="cahaba--page-image" src="%s" alt="%s" /><!-- PAGE BREAK -->`,
				imgPath,
				alt,
			), nil
		},
	)
	epub.RegisterShortcode(
		"clickimage",
		func(b *epub.Book, s1 string, m map[string]any, s2 string) (string, error) {
			if onlyLN {
				return "", nil
			}
			img := fmt.Sprint(m["image"])
			alt := fmt.Sprint(m["alt"])
			imgPath, ok := b.LookupImage(img)
			if !ok {
				return "", errors.New("Missing Image")
			}
			return fmt.Sprintf(
				`<img class="cahaba--normal-image" src="%s" alt="%s" />`,
				imgPath,
				alt,
			), nil
		},
	)
}
