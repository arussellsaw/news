# Swan [![Build Status](https://github.com/thatguystone/swan/workflows/test/badge.svg)](https://github.com/thatguystone/swan/actions) [![GoDoc](https://godoc.org/github.com/thatguystone/swan?status.svg)](https://godoc.org/github.com/thatguystone/swan)

<img src="https://github.com/thatguystone/swan/raw/master/logo.png" alt="swan" align="left" hspace="20" vspace="0" />

An implementation of the Goose HTML Content / Article Extractor algorithm in golang.

Swan allows you to extract cleaned up text and HTML content from any webpage by removing all the extra junk that so many pages have these days.

Check out [the go documentation page](https://godoc.org/github.com/thatguystone/swan) for full usage and examples.

<br clear="all"/>

## Features

* Main content extraction from almost any source
* Extract HTML content with images
* Get article metadata, publish dates, and a lot more
* Recognize different content types and apply special extractions (currently only recognizes comic sites and normal sites)

## Planned

* Inline videos into HTML content when found in an article
* Recognize news sources and extract corresponding video / audio content
* Recognize and extract more types of content
* An interesting idea: https://github.com/buriy/python-readability/issues/57#issuecomment-67926023
