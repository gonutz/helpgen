# help

Create HTML and RTF help files using a simple, text-based markup language.

This program takes as input a special markup, described below, and produces as output a single-file, either HTML or RTF, useful as help or documentation. It supports stylized text, multiple sizes of captions, in-document links and images. All content is packed into a single output file, which means no need to ship a folder for your HTML help and thus no errors with relative or absolute path references.

# Installation and Usage

You need to have the [Go programming language](https://golang.org/) installed to build this program. Download and install it by simply running:

`go get github.com/gonutz/helpgen`

To then transform a markup file `doc.help` just type:

`helpgen -html doc.help > output.html`

or

`helpgen -rtf doc.help > output.rtf`

# Syntax

Besides simple text, a help file can contain special commands to insert links, captions, images and more into the file. Below is a description of all special syntax elements.

## Text

The most basic syntax element is simple text. All non-special syntax text will appear as regular text in the output. The input file is expected to be encoded in UTF-8 (or plain ASCII, which is a subset of UTF-8).

## Font Styles

To make text bold, put it in '*' characters like so

`*bold*`

For italic, use '/'

`/italic/`

and to have both styles, use either

`*/bold and italic/*`

or

`/*bold and italic*/`

Stylized text can only span a single line so if you want a whole paragraph to appear bold, enclose all lines in '*' characters.

## Images

To insert an image file into the document, put its name in brackets like so

`[mona lisa.jpeg]`

Only give the file name, not the full path. This will go through all folders under the current working directory in a breadth-first manner, looking for the first file of the given name. This means that images in the same folder as the code file have the highest priority, after that images in sub-folders, then images in sub-sub-folders, etc. This means that you could have different versions of the same image, maybe as a backup, in one or more backup folders, the current one lying in the code folder and it would be used correctly.

## Variables

If there is an expression that you want to use multiple times throughout the document, but it might change or is very long, you might want to define a variable for it like so

`[\varname=this is some text]`

This will define a variable `varname` with the content `this is some text`. You can now insert this variable in your text by putting its name in brackets

`[varname]`

This will replace the `[varname]` with the actual text.

The definition line itself (starting with `[\`) will be removed from the text. Variables can be defined anywhere in the text, even after their actual usage. Each variable name can only be defined once.

Note that variable names can only contain letters, digits and underscores, no spaces. Also note that the text is used verbatim in the output, meaning all spaces and characters that are special syntax in other contexts, are copied verbatim. If you want the variable context to appear bold for example, write `*[varname]*`.

## Captions

There are four sizes of captions: the title, chapters, sub-chapters and sub-sub-chapters.

The document title can only appear once in the whole document, it uses the largest font and will also appear as the HTML document's title, i.e. your browser tab will display the title when you open the generated HTML file. The title is enclosed by a line of `=` right before and right after it, like so

```
==============
Document Title
==============
```

A document can contain any number of chapters and sub-chapters, they are marked with a special line right below them, like so

```
1. Chapter One
==============

1.1 Sub-Chapter with minus line
-------------------------------

1.1.1 Sub-Sub-Chapter with dotted line
......................................
```

All chapters and sub-chapters can be used as link targets (see below).

## Links

You can link to any chapter or sub-chapter in your document by putting its caption inside brackets like so

```
Introduction
============

...

Here is a link to the [Introduction].
```

This will make the text `Introduction` a link in the HTML output.

If you want the link text to be different from the actual caption, you can include an alternative text inside the brackets like this

```
1.1 Some Caption
----------------

Here is a link with an [alternative text[1.1 Some Caption]].
```

## Special Characters

These characters are used to start special syntax elements: `[`, `*`, `/`, `=`, `-`, `.`

To use these characters verbatim in the text, you have to escape them by enclosing them in brackets, e.g. `[[]` or `[*]`.

Note that the special characters do not always start a special element, e.g. a single `=` will always appear as is in the output text. Try using no brackets and see if the output is fine before escaping them, this way your text is not cluttered with brackets.