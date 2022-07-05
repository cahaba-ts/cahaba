package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cahaba-ts/epub"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
)

func numericValidator(input string) error {
	_, err := strconv.Atoi(input)
	return err
}
func New(c *cli.Context) error {
	args := c.Args()
	vpath, _ := os.Getwd()
	if len(args) == 0 {
		if c.Bool("debug") {
			fmt.Println("Creating new Volume in Current Directory")
		}
	} else {
		folder := strings.Join(args, " ")
		os.Mkdir(folder, os.ModeDir|os.ModePerm)
		if c.Bool("debug") {
			fmt.Printf("Creating new Volume in '%s'\n", folder)
		}
		vpath = filepath.Join(vpath, folder)
	}

	name, _ := (&promptui.Prompt{Label: "Name of the Volume"}).Run()
	desc, _ := (&promptui.Prompt{Label: "Description of the Volume"}).Run()
	chapters, _ := (&promptui.Prompt{Label: "Number of Chapters", Validate: numericValidator}).Run()

	tpath := filepath.Join(vpath, "text")
	if c.Bool("debug") {
		fmt.Println("Making text and image folders")
	}
	os.Mkdir(tpath, os.ModeDir|os.ModePerm)
	os.Mkdir(filepath.Join(vpath, "images"), os.ModeDir|os.ModePerm)
	os.Mkdir(filepath.Join(vpath, "ln_images"), os.ModeDir|os.ModePerm)
	os.Mkdir(filepath.Join(vpath, "assets"), os.ModeDir|os.ModePerm)

	chapterCount, _ := strconv.Atoi(chapters)
	chapterList := []string{}
	for i := 1; i <= chapterCount; i++ {
		if c.Bool("debug") {
			fmt.Println("Making file for chapter", i)
		}
		os.WriteFile(
			filepath.Join(tpath, fmt.Sprintf("chapter%02d.md", i)),
			[]byte(fmt.Sprintf(ChapterBody, i)),
			os.ModePerm,
		)
		chapterList = append(chapterList, fmt.Sprintf(`"chapter%02d"`, i))
	}
	if c.Bool("debug") {
		fmt.Println("Making config file")
	}
	os.WriteFile(
		filepath.Join(vpath, "volume.toml"),
		[]byte(fmt.Sprintf(
			ConfigBody,
			name,
			desc,
			strings.Join(chapterList, ",\n    "),
		)),
		os.ModePerm,
	)
	data, _ := epub.RetrieveTemplate("default.css")
	os.WriteFile(
		filepath.Join(vpath, "volume.css"),
		data,
		os.ModePerm,
	)

	return nil
}

const ConfigBody = `Title = "%s"
Cover = "images/cover.png"
Author = "Unknown"
Description = "%s"
ImageFolder = "images"
LNImageFolder = "ln_images"
ReleaseDate = "2022-02-02"
Header = """<header style="justify-content: space-between">
<div class="even-page">_TITLE_ by _AUTHOR_</div>
<div></div>
<div class="odd-page"><i>Find me at example.com</i></div>
</header>"""
Sections = [
    %s
]`

const ChapterBody = `---
title: "Chapter %d: Something"
---

Add body here
`
