# STemplate - Simple Template parser for the Linux Shell

## What is it
STemplate is a simple template parser for the Linux Shell (bash, ksh, csh, etc).
The purpose is to easily parse templates and fill them in with values from a "dictionary" file.
See the [example](#example) below for a quick introduction.

The template file contains placeholders and the dictionary file contains key=value pairs. STemplate will
take the dictionary file and fill in the template file placeholders.

The dictionary file can be in YAML, JSON or TOML format. The template language is provided by the Golang 
[text/template](https://golang.org/pkg/text/template/) package.

The idea is similar to [Jinja templates](http://jinja.pocoo.org/) that you apply in
[Ansible](https://docs.ansible.com/ansible/latest/index.html) with the [template module](https://docs.ansible.com/ansible/latest/modules/template_module.html).
Now you can do it in plain shell.

## How to get it
Check the [releases](https://github.com/freshautomations/stemplate/releases) page.

## How to build it from source
Install [Golang](https://golang.org/doc/install), then run:
```cgo
GO111MODULE=on go get github.com/freshautomations/stemplate.git
```

## How to use
```$bash
stemplate dictionary.json file.template 
```

Check the `test.json` and `test.template` files for an [example](#example).

Optionally, you can use the `--output` or `-o` flags to add a file where the result will be written,
instead of the default `stdout`.

## Caveats
None that I know of. Admittedly, I use it for simple things. Usually, Bash does not require very complex configurations.

## Example

Template file `test.template`:
```
Hi {{ .user }}!

Welcome to this {{ .filename }} template demonstration.

You should see a few examples of
* List item: {{ index .list 0 }}
* Map item: {{ .map.test }}
* {{ index .gospecific 0 -}} {{ index .gospecific 1 }} specific {{ print "stuff" }}
```

Dictionary file `test.yaml`:
```yaml
user: guest
filename: test

list:
    - first
    - second
    - third

map:
    test: testmap
    nottest: "not a test map"

gospecific:
    - Go
    - lang
```

Run:
```bash
stemplate test.yaml test.template --output result.txt
```

Result in `result.txt`:
```
Hi guest!

Welcome to this test template demonstration.

You should see a few examples of
* List item: first
* Map item: testmap
* Golang specific stuff
```