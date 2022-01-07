# protoc-gen-proto2asciidoc

## About

proto2asciidoc is a [protoc](https://github.com/protocolbuffers/protobuf/releases) plugin
that generates asciidoc documentation from the Protobuffer IDL files.

The goal of the project was to prevent API documentation inconsistenties
between the actual API and the docs.
By generating the code from the Protobuf definition the documentation is always
in sync.

The current version is a plugin, which is a from scratch rewrite of the previous one that
generated asciidoc too but parsed the Protobuffer files itself.

Can be used in conjunction with [code2asciidoc](https://github.com/productsupcom/code2asciidoc)
to produce even more consistent API documentation.

## Options

The following options for the plugin can be passed to protoc, the default for everything is off.
Ensure the options are comma separated.

<div class="formalpara-title">

**Example**

</div>

``` shell
--proto2asciidoc_opt=rest=on,sorted=on
```

<div class="note">

All boolean values are either `on` or `off`

</div>

|             |                |                                                                               |
|-------------|----------------|-------------------------------------------------------------------------------|
| Option      | Accepted Value | Description                                                                   |
| extension   | boolean        | Enables the optional Ruby asciidoctor extension for more stylized hyperlinks  |
| collapsible | boolean        | When enabled the tables produced are collapsible through HTML                 |
| rest        | boolean        | Output the REST information for the Service endpoints too                     |
| icons       | boolean        | Instead of string for true and false asciidoc icons are used                  |
| sorted      | boolean        | When enabled the Services, Enum and Messages are sorted alphabetically vs the 
                                order they appear in the file in.                                              |

## Optional Extension

This tool assumes the extension in `asciidoc/extension` will be loaded when using
AsciiDoctor.

The following snippet can be used inside a Makefile.

<div class="note">

Of course you have to ensure that the directory is in a location your
Makefile can find it.

</div>

<div class="formalpara-title">

**Makefile**

</div>

``` Makefile
# Asciidoctor settings
ASCIIDOC_EXT := -r ./asciidoctor/extensions/proto2asciidoc-inline-macro.rb
```

The following shows how the README (if on Github) you’re currently reading
has been produced.

<div class="formalpara-title">

**Optional formats**

</div>

    markdown:
        @asciidoctor ${ASCIIDOC_EXT} docs/readme.adoc -b docbook -a leveloffset=+1 -o - | pandoc  --markdown-headings=atx --wrap=preserve -t gfm -f docbook - > README.md

## Installation

``` shell
go get github.com/productsupcom/protoc-gen-proto2asciidoc/cmd/protoc-gen-proto2asciidoc
```
