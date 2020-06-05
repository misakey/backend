# API Documentation

This is the documentation of the API exposed by Misakey backend.
At the moment it is primarily intended for internal use,
as a specification for our backend developers
and as documentation for our frontend developers.

It consists in a [Hugo][] static website,
so you need to install Hugo if you want to build it yourself,
but it should already be quite useful just by reading the source
(see the `content` directory).

When you have installed Hugo,
just run `hugo server` in the same directory as this README file
and go to the address Hugo is listenning to.
The website rebuilds automatically and your browser refreshes the page
when you change something in the `content` directory.

## Contributions Guideline

- do not use heading level 1 (lines starting with one `#`). It is used by Hugo for the page title.

## The `include` shortcode

You can factorize snippets of text that repeat over several pages
using a custom [Hugo shortcode][] named `include`:

    {{% include "include/hashParameters.json" %}}

Additionally to the path to the snippet (please put snippets in the `include` directory),
the shortcode takes two other parameters that are *optional*:
the first is an integer indicating the **indentation** (using spaces) that must be applied to the included snippet
(useful when including a JSON object inside another JSON object),
and the second is a boolean represented as `0` (false, the default) or `1` (true)
indicating whether the indent should be applied to the very first line of the snippet.

Actually, snippets can themselves include other snippets through a similar syntax.
This is still quite an experimental feature though.
For more information see the source code of the shortcode in `layouts/shortcodes/include.html`

[Hugo]: https://gohugo.io/
[Hugo shortcode]: https://gohugo.io/content-management/shortcodes/
