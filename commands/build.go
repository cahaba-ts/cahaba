package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/frontmatter"
	"github.com/cahaba-ts/epub"
	"github.com/urfave/cli"
)

type BuildConfig struct {
	Title         string   `toml:"Title" yaml:"title"`
	Format        string   `toml:"Format" yaml:"format"`
	Author        string   `toml:"Author" yaml:"author"`
	Cover         string   `toml:"Cover" yaml:"cover"`
	Description   string   `toml:"Description" yaml:"description"`
	ReleaseDate   string   `toml:"ReleaseDate" yaml:"release_date"`
	ImageFolder   string   `toml:"ImageFolder" yaml:"image_folder"`
	LNImageFolder string   `toml:"LNImageFolder" yaml:"ln_image_folder"`
	Sections      []string `toml:"Sections" yaml:"sections"`
}

var onlyLN = false

func Build(c *cli.Context) error {
	if c.Bool("debug") {
		epub.Debug = true
	}
	args := c.Args()
	vpath, _ := os.Getwd()
	if len(args) == 0 {
		if epub.Debug {
			fmt.Println("Building Volume in Current Directory")
		}
	} else {
		folder := strings.Join(args, " ")
		os.Mkdir(folder, os.ModeDir|os.ModePerm)
		if epub.Debug {
			fmt.Printf("Building Volume in '%s'\n", folder)
		}
		vpath = filepath.Join(vpath, folder)
	}

	configureShortcodes()

	if epub.Debug {
		fmt.Println("Loading book config")
	}
	config := loadConfig(vpath)
	if epub.Debug {
		fmt.Println("Generating epub")
	}
	err := buildBookEpub(vpath, config)
	if err != nil {
		return err
	}
	if config.LNImageFolder == "" {
		if epub.Debug {
			fmt.Println("No LN Image Folder, skipping LN Image only output")
		}
		return nil
	}
	if epub.Debug {
		fmt.Println("Swapping to LN Image Only book")
	}
	config.Title += " (LN Images Only)"
	config.ImageFolder = config.LNImageFolder
	onlyLN = true

	return buildBookEpub(vpath, config)
}

func buildBookEpub(vpath string, config BuildConfig) error {
	output := configureEpub(vpath, config)
	pastPrologue := false
	var err error
	for _, section := range config.Sections {
		cfg, body := loadChapter(vpath, section)
		if strings.HasPrefix(section, "chapter") {
			pastPrologue = true
			if cfg.Format == "md" {
				err = output.AddChapterMD(cfg.Title, body)
			} else {
				err = output.AddChapterHTML(cfg.Title, []string{body})
			}
			if err != nil {
				log.Fatal("Add Chapter: ", err)
			}
		} else if pastPrologue {
			if cfg.Format == "md" {
				err = output.AddPostscriptMD(cfg.Title, body)
			} else {
				err = output.AddPostscriptHTML(cfg.Title, []string{body})
			}
			if err != nil {
				log.Fatal("Add Chapter: ", err)
			}
		} else {
			if cfg.Format == "md" {
				err = output.AddIntroductionMD(cfg.Title, body)
			} else {
				err = output.AddIntroductionHTML(cfg.Title, []string{body})
			}
			if err != nil {
				log.Fatal("Add Chapter: ", err)
			}
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
	var f *os.File
	xhtml := false

	mdPath := filepath.Join(vpath, "text", chapterFile+".md")
	if _, err := os.Stat(mdPath); err == nil {
		f, err = os.Open(mdPath)
		if err != nil {
			log.Fatal("Open Chapter File: ", chapterFile, err)
		}
	}

	htmlPath := filepath.Join(vpath, "text", chapterFile+".html")
	if _, err := os.Stat(htmlPath); f == nil && err == nil {
		f, err = os.Open(htmlPath)
		if err != nil {
			log.Fatal("Open Chapter File: ", chapterFile, err)
		}
		xhtml = true
	}

	xhtmlPath := filepath.Join(vpath, "text", chapterFile+".html")
	if _, err := os.Stat(xhtmlPath); f == nil && err == nil {
		f, err = os.Open(xhtmlPath)
		if err != nil {
			log.Fatal("Open Chapter File: ", chapterFile, err)
		}
		xhtml = true
	}

	bc := BuildConfig{}
	body, err := frontmatter.Parse(f, &bc)
	if err != nil {
		log.Fatal("Parse Chapter File: ", chapterFile, err)
	}

	if xhtml {
		bc.Format = "html"
	} else {
		bc.Format = "md"
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
	if err := book.SetCover(filepath.Join(vpath, config.Cover)); err != nil {
		log.Fatal("Set Cover: ", err)
	}

	book.SetReleaseDate(config.ReleaseDate)
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
				return "", errors.New("Missing Image: " + img)
			}
			return fmt.Sprintf(
				`<img class="cahaba--normal-image" src="%s" alt="%s" />`,
				imgPath,
				alt,
			), nil
		},
	)
	epub.RegisterShortcode(
		"alwaysimage",
		func(b *epub.Book, s1 string, m map[string]any, s2 string) (string, error) {
			img := fmt.Sprint(m["image"])
			alt := fmt.Sprint(m["alt"])
			imgPath, ok := b.LookupImage(img)
			if !ok {
				return "", errors.New("Missing Image: " + img)
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
				return "", errors.New("Missing Image: " + img)
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
				return "", errors.New("Missing Image: " + img)
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
				return "", errors.New("Missing Image: " + img)
			}
			return fmt.Sprintf(
				`<img class="cahaba--normal-image" src="%s" alt="%s" />`,
				imgPath,
				alt,
			), nil
		},
	)
}
