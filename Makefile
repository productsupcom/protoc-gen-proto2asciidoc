# tag::extension[]
# Asciidoctor settings
ASCIIDOC_EXT := -r ./asciidoctor/extensions/proto2asciidoc-inline-macro.rb
# end::extension[]

.PHONY: markdown

# tag::markdown[]
markdown:
	@asciidoctor ${ASCIIDOC_EXT} docs/readme.adoc -b docbook -a leveloffset=+1 -o - | pandoc  --markdown-headings=atx --wrap=preserve -t gfm -f docbook - > README.md
# end::markdown[]