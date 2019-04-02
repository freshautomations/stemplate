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
```$bash
stemplate my.template --file dictionary.json
```
OR
```$bash
export envvar1="value1"
export envvar2="value2"
export envlist1="listitem1,listitem2"
export envlist2="listitem3"
export envmap1="mapkey1=mapvalue1,mapkey2=mapvalue2"
stemplate my.template --string envvar1,envvar2 --list envlist1,envlist2 --map envmap1
```

Check the [examples](#examples) section for more.

* `--file` will load the content of the filename as the data structure
* `--string` will evaluate the content of the environment variables as a string.
* `--list` will evaluate the content of the environment variables as a comma-separated list of strings.
* `--map` will evaluate the content of the environment variables as a comma-separated list of key-value pairs where both key and value are strings.

When using the `--file` parameter, the dictionary file can contain complex variable definitions, like maps within a list.

When using environment variables with any of the `--string`, `--list` or `--map` parameters, the values have to be simple string values.

When using both `--file` and any of the environment variables flags, the resultant data structure is the combination of both data sets.
If the same variable name is used in both the file and an environment variable, the environment variable will take precedence.

Optionally, you can use the `--output` or `-o` flags to add a file where the result will be written,
instead of the default `stdout`.

## Caveats
Using the `--file` parameter will allow the full extent of the Golang text/template package to be used, while using environment variables will only allow string values.

## Examples

## Using `--file`
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
stemplate test.template --file test.yaml --output result.txt
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
