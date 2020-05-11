PROJECT_NAME := proto2asciidoc
PROJECT_REPO := https://github.com/productsupcom/proto2asciidoc
GIT_VERSION_NAME := $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q HEAD || git rev-parse HEAD)

.PHONY: man

man:
	@asciidoctor -a version=${GIT_VERSION_NAME} -b manpage man/proto2asciidoc.1.adoc
	@gzip man/proto2asciidoc.1

markdown:
	@asciidoctor docs/readme.adoc -b docbook -a leveloffset=+1 -o - | pandoc  --atx-headers --wrap=preserve -t gfm -f docbook - > README.md