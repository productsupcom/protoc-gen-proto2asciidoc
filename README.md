# proto2asciidoc

## About

proto2asciidoc generates asciidoc documentation from a Protobuffer definition
file. The goal of the project was to prevent API documentation inconsistenties
between the actual API and the docs.

By generating the code from the Protobuf definition the documentation is always
in sync.

It can either generate only the Protobuffer Messages/Enums as output or full
api-docs output depending on the flags passed.

Can be used in conjunction with [code2asciidoc](https://github.com/productsupcom/code2asciidoc)
to produce even more consistent API documentation.

## Usage

**--source string**
Source file to parse into AsciiDoc, recommended is to set the absolute path.

**--out**
File to write to, if left empty writes to stdout

**--f**
Overwrite the existing out file

**--no-header**
Do not set a document header and ToC

**--api-dir**
Relative path from the out to the api dir. E.g. docs/generated/api.adoc is the out,
the api dir is docs/api/
then set --api-dir ../api

This path will be used to set the includes for asciidoc

**--api-docs**
Generate a full API documentation, including files from the out file relative dir
../api/SERVICE/endpoint.adoc.

Do not set if you only want Messages and/or Enums
