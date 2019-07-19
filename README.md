# STemplate - Simple Template parser for the Linux Shell

## What is it
STemplate is a simple template parser for the Linux Shell (bash, ksh, csh, etc).
The purpose is to easily parse templates and fill them in with values from a "dictionary" file.
See the [examples](#examples) below for a quick introduction.

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
```bash
stemplate my.template --file dictionary.json
```
OR
```bash
export envvar1="value1"
export envvar2="value2"
export envlist1="listitem1,listitem2"
export envlist2="listitem3"
export envmap1="mapkey1=mapvalue1,mapkey2=mapvalue2"
stemplate my.template --string envvar1,envvar2 --list envlist1,envlist2 --map envmap1
```
OR
```bash
stemplate my.template --env
```

Check the [examples](#examples) section for more.

* `--file` will load the content of the filename as the data structure
* `--string` will evaluate the content of the environment variables as a string.
* `--list` will evaluate the content of the environment variables as a comma-separated list of strings.
* `--map` will evaluate the content of the environment variables as a comma-separated list of key-value pairs where both key and value are strings.
* `--env` will evaluate _all_ environment variables as strings.

When using the `--file` parameter, the dictionary file can contain complex variable definitions, like maps within a list.

When using environment variables with any of the `--string`, `--list` or `--map` parameters, the values have to be simple string values.

When using both `--file` and any of the environment variables flags, the resultant data structure is the combination of both data sets.
If the same variable name is used in both the file and an environment variable, the environment variable will take precedence.

When using the `--env` option, all environment variables will be evaluated as strings. This has the lowest precedence and
all other options will overwrite what was received from the environment by this flag.

In short, precedence from lowest to highest: `--env`, `--file`, `--string`, `--list`, `--map`.

Optionally, you can use the `--output` or `-o` flags to add a file where the result will be written,
instead of the default `stdout`.

### Special functions
STemplate introduces special functions to make templates more versatile.

#### substitute
Transform a variable's content to variable name. (Something like *.(.var)* would be, if it was valid.) For example:

Dictionary file `test.yaml`:
```yaml
environment: dev

dev:
    description: "developer's environment"
    account_id: 0

prod:
    description: "production environment - exercise caution"
    account_id: 1
```

Template file `test.template`:
```gotemplate
This environment is the {{ index (substitute .environment) "description" }}.
```

Run:
```bash
stemplate test.template --file test.yaml
```

Result:
```gotemplate
This environment is the developer's environment.
```

Note: `(.environment | substitute)` is also valid. Use whichever makes the code more readable.

#### counter
Create a list of numbers, starting with 0. (Something like `range [5]int{}` in Go but with a variable instead of a constant.) Useful for the `range` function. For example:

Dictionary file `test.yaml`:
```yaml
howmany: 3
```

Template file `test.template`:
```gotemplate
Write some dots {{- range counter .howmany }}.{{end}}
```

Run:
```bash
stemplate test.template --file test.yaml
```

Result:
```gotemplate
Write some dots...
```

## Caveats
Using the `--file` parameter will allow the full extent of the Golang text/template package to be used, while using environment variables will only allow string values.

## Examples

## Using `--file`
Template file `test.template`:
```gotemplate
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
stemplate test.template --file test.yaml --output result.txt
```

Result in `result.txt`:
```gotemplate
Hi guest!

Welcome to this test template demonstration.

You should see a few examples of
* List item: first
* Map item: testmap
* Golang specific stuff
```

## Using environment variables
Assume the `test.template` file from the above example. Run
```bash
export user="guest"
export filename="test"
export list="first,second,third"
export gospecific="Go,lang"
export map="test=testmap,nottest=not a test map"
stemplate test.template --string user,filename --list list,gospecific --map map
```

The result will be the same as `result.txt`, only this time it will be printed on the standard output.
