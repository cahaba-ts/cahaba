# Cahaba

Cahaba is the tool I wrote to handle building volumes. It has two modes, generation and
build. A volume is a collection of markdown chapters and images that are built into 
distributable epub and pdf documents. It can output four versions of a volume with a 
single command!

__You must install calibre, cahaba uses the ebook-polish command to optimize the epub__
__after generation, as well as the ebook-convert command to build the pdf version.__

## cahaba new

Creates a new volume and prompts you for the various customizable parts.
Call it with cahaba new volume_name to make the new volume in a new folder,
otherwise it uses the current folder.

## cahaba build

Builds the volume in the current directory or the directory passed as an 
argument. Will output a Title.epub and Title.pdf in the current
directory after building the volume. 