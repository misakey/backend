# Email templates

This folder contains the emails templates we send.

## File structure

The HTML and CSS has to be compiled to be compatible with emails clients. So in this folder we have 3 types of files:
- `.source.html` files: the files used to create the templates. We work on those files on a browser to build the email. They are not compatible with all email clients.
- `.html.tpl` files: the files used for HTML version of emails. Compiled to be readable by most of the email clients, with templates vars.
- `.txt.tpl` files: the text version of the emails, with templates vars. They are mandatory to be send aside of the HTML version of the emails

## Tools 

### HTML to email compatible HTML

#### Make buttons, and border radius

You can use [buttons.cm](https://buttons.cm/)

#### Minify CSS

You can use [this tool](https://cssminifier.com/)

#### Inlinize CSS

You can use [this tool](https://premailer.io/) (previously we used: [Mailchimp inline-css tool](https://templates.mailchimp.com/resources/inline-css/), but it's less complete)


### HTML to TXT

You can use [Mailchimp html-to-text tool](https://templates.mailchimp.com/resources/html-to-text/) as a base

