package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		fmt.Println("Creating new Volume in Current Directory")
	} else {
		folder := strings.Join(args, " ")
		os.Mkdir(folder, os.ModeDir|os.ModePerm)
		fmt.Printf("Creating new Volume in '%s'\n", folder)
		vpath = filepath.Join(vpath, folder)
	}

	name, _ := (&promptui.Prompt{Label: "Name of the Volume"}).Run()
	desc, _ := (&promptui.Prompt{Label: "Description of the Volume"}).Run()
	chapters, _ := (&promptui.Prompt{Label: "Number of Chapters", Validate: numericValidator}).Run()

	tpath := filepath.Join(vpath, "text")
	os.Mkdir(tpath, os.ModeDir|os.ModePerm)
	os.Mkdir(filepath.Join(vpath, "images"), os.ModeDir|os.ModePerm)

	chapterCount, _ := strconv.Atoi(chapters)
	chapterList := []string{}
	for i := 1; i <= chapterCount; i++ {
		os.WriteFile(
			filepath.Join(tpath, fmt.Sprintf("chapter%02d.md", i)),
			[]byte(fmt.Sprintf(ChapterBody, i)),
			os.ModePerm,
		)
		chapterList = append(chapterList, fmt.Sprintf(`"chapter%02d"`, i))
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

	return nil
}

const ConfigBody = `Title = "%s"
Cover = "cover.png"
Description = "%s"
Sections = [
    %s
]`

const ChapterBody = `---
title: "Chapter %d: Something"
---

Add body here
`
