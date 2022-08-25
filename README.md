# Mini Shortener

<img src="logo.png" width="80" />

Configuration-based, short link service in less than 200 lines.

## Config

The configuration file is in plain text format, and each line contains a redirect rule:

```
"/ping" => "https://github.com/soulteary/mini-shortener"
```

The program startup will read the `rules` file in the current working directory by default.

## Docker

Download the rules configuration file in the project, or refer to the documentation above, create a local file, and execute the following commands:

```bash
docker run -v `pwd`/rules:/app/rules -p 8080:8901 soulteary/mini-shortener -d
```
