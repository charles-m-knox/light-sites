# Light Sites

Light Sites is a Go-powered, no-JavaScript, template-friendly, lightweight static markdown parser and server. It aims to combine Bootstrap CSS (and other customizable CSS if you choose) with Markdown with minimal configuration.

Light Sites was created to bring together a few different niches into a single package:

* No-JavaScript, inspired by:
  * The Tor Browser, and how JavaScript is used to violate privacy when browsing the web - we should be able to acquire all the information we need from a website without running JS
  * The [nojs.club](https://nojs.club/)
* Lightweight, Go-based codebase
* Docker-first deployment
* Markdown-based static blog tool
* HTML Template rendering engine
* Slightly opinionated preference to a common responsiveBootstrap layout

Due to its youth, it only boasts a couple features, but simple templating is probably its most useful feature. See [templating](#template-tag) for more info.

Code unit test coverage is currently at 96.7% (for packages with code worth testing).

## Table of Contents

- [Light Sites](#light-sites)
  - [Table of Contents](#table-of-contents)
  - [Usage](#usage)
  - [Deployment](#deployment)
  - [Configuration](#configuration)
  - [Special Tags/Behavior](#special-tagsbehavior)
    - [Special Behavior](#special-behavior)
      - [Auto-refresh](#auto-refresh)
    - [Important Tags](#important-tags)
      - [`attributes` Tag (Required)](#attributes-tag-required)
      - [`directory` Tag](#directory-tag)
      - [`template` Tag](#template-tag)
    - [Behind the Scenes Tags](#behind-the-scenes-tags)
      - [`title` Tag](#title-tag)
      - [`body` Tag](#body-tag)
      - [`head` Tag](#head-tag)
      - [`table` Tag](#table-tag)
      - [`img` Tag](#img-tag)
  - [Roadmap](#roadmap)

## Usage

Simply create an `index.md` file in `./src/content`. The only requirement is that the markdown document requires the [`<attributes>`](#attributes-tag-required) tag to be defined with a `title="Home"` attribute (change as desired). Everything else can be normal markdown.

Continue to the [deployment](#deployment) steps next.

## Deployment

You need:

* GNU Make
* Docker or Go 1.15

For Docker, run:

```bash
make build # make gobuild, if using Go
make run   # make gorun, if using Go
```

Finally, navigate to `http://localhost:8099/content/index.html` to view `src/content/index.md` in its rendered form.

> *Note: If you change the configured `listenAddr` in `config.yml`, or want to change the port mapping in the `docker run` command, please update the Makefile accordingly.

To add new documents, ensure that the [`<attributes title="Hello World!"></attributes>`]((#attributes-tag-required) tag is placed preferably at the top of your Markdown document.

## Configuration

Edit `config.yml` to meet your needs.

## Special Tags/Behavior

There are a few custom HTML tags that are processed by the Light Sites rendering engine.

### Special Behavior

#### Auto-refresh

Every 30 minutes (configurable), the documents are reloaded. This means that documents are served from memory for fastest performance, but are potentially outdated if a recent change was made.

*TODO: enable/disable this feature in `config.yml`.*

### Important Tags

Before spending a lot of time creating markdown files, take a look at the following tags and see if they are useful.

#### `attributes` Tag (Required)

The `<attributes>` tag must be placed in the first line of the document. Currently, the `title` attribute **must be set** or else the document may not function correctly. Example:

```xml
<attributes title="Hello World!"></attributes>
```

#### `directory` Tag

Use the `<directory>` tag to render links to all documents in the `src/content` directory as a `<ul><li>...</li></ul>` tree. To hide a document, prefix it with a `.`, such as `src/content/.page2.md`. To visit a hidden page, visit `http://localhost:8099/content/.page2.html`. Traversing folders is supported. *This behavior may change in the future.*

To avoid issues and ensure smoothest functionality, ensure that your `config.yml` specifies `directories.documents` as `src/content` for example, and NOT as `./src/content`.

#### `template` Tag

Templating is the most useful part of Light Sites. It allows you to reuse HTML elements and pass-in custom variables. Example:

```html
<template file="alert.html" alert-text="Heads up!"></template>
```

And the contents of `src/templates/alert.html` are:

```html
<div class="alert alert-primary">
	{{alert-text}}
</div>
```

`{{alert-text}}` wil render the text `Heads up!` when the static HTML document is produced.

Recursive/nested templating is currently not tested and likely does not work.

### Behind the Scenes Tags

The following tags are all handled by the Light Sites engine, and do not require any interaction. Consider this a behavioral documentation section rather than actual instructions.

#### `title` Tag

The `<title>Page Title</title>` tag should be set by you in your documents. Currently, it is not controlled by Light Sites.

#### `body` Tag

When the `<body>` tag is encountered, its first child is given a parent of:

```html
<main role="main">
    <div class="container">
        <div class="row">
            <div class="col-lg-12">
            </div>
        </div>
    </div>
</main>
```

For example, if at first the HTML content looks like this...

```html
<body>
    <p>
        Some text
    </p>
</body>
```

... it will become:

```html
<body>
    <main role="main">
        <div class="container">
            <div class="row">
                <div class="col-lg-12">
                    <p>
                        Some text
                    </p>
                </div>
            </div>
        </div>
    </main>
</body>
```

It is worth noting that the markdown-to-HTML engine automatically produces a `<body>` tag, whose children are the rendered HTML contents of your markdown document. You don't need to worry about this behavior at all, but it helps to understand how the markdown document is manipulated to the result that you see on screen.

#### `head` Tag

When encountered, the `<head>` tag is given a standard Bootstrap CSS import, using `<link href="/assets/bootstrap.min.css>` with a few other attributes. Additionally, it imports `/assets/custom.css` the same way. These are configurable in the `config.yml`.

#### `table` Tag

`<table>` tags are updated to use Bootstrap's responsive table classes, as well as adding zebra striping and active mouse hover highlighting to row elements. Tables are always wrapped in a `<div class="table-responsive"></div>` element. Currently, this behavior is not configurable.

#### `img` Tag

All `<img>` tags should be given a `max-width` of 100%, as shown in the default `config.yml`. This is to prevent occasional issues where the image may not respect the width of the parent container(s).

## Roadmap

* Recursive/nested templating
* Template `heading="true"` should break outside the container div and actually be at the top
* Directory listing should use doc titles instead of their relative path names
* Configurable behavior for tables instead of enforcing `<div class="table-responsive">` wrapping
